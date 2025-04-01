package k8s

import (
	"context"
	"log/slog"
	"time"

	"github.com/heathcliff26/fleetlock/pkg/k8s/utils"
	systemdutils "github.com/heathcliff26/fleetlock/pkg/systemd-utils"

	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

type Client struct {
	client              kubernetes.Interface
	namespace           string
	drainTimeoutSeconds int32
	drainRetries        int
}

// Create a new kubernetes client, defaults to in-cluster if no kubeconfig is provided
func NewClient(config Config) (*Client, error) {
	client, err := utils.CreateNewClientset(config.Kubeconfig)
	if err == rest.ErrNotInCluster {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	ns, err := utils.GetNamespace()
	if err != nil {
		return nil, err
	}

	if config.DrainTimeoutSeconds < 1 {
		return nil, NewErrorDrainTimeoutSecondsInvalid()
	}

	return &Client{
		client:              client,
		namespace:           ns,
		drainTimeoutSeconds: config.DrainTimeoutSeconds,
		drainRetries:        config.DrainRetries,
	}, nil
}

// Create a test client with a fake kubernetes clientset
func NewFakeClient() (*Client, *fake.Clientset) {
	fakeclient := fake.NewSimpleClientset()
	return &Client{
		client:              fakeclient,
		namespace:           "fleetlock",
		drainTimeoutSeconds: 300,
	}, fakeclient
}

// Drain a node from all pods and set it to unschedulable.
// Status will be tracked in lease, only one drain will be run at a time.
func (c *Client) DrainNode(node string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.drainTimeoutSeconds)*time.Second)
	defer cancel()

	lease := NewLease(drainLeaseName(node), c.client.CoordinationV1().Leases(c.namespace))
	err := lease.Lock(ctx, c.drainTimeoutSeconds)
	if err != nil {
		return err
	}

	err = c.drainNode(ctx, node)
	if err != nil {
		err2 := lease.Error(ctx)
		if err2 != nil {
			slog.Error("Failed to set drain lease to error state", slog.String("node", node), "err", err)
		}
		return err
	}

	return lease.Done(ctx)
}

// Drain a node of all pods, skipping daemonsets
func (c *Client) drainNode(ctx context.Context, node string) error {
	_, err := c.client.CoreV1().Nodes().Patch(ctx, node, types.MergePatchType, nodeUnschedulablePatch(true), metav1.PatchOptions{})
	if err != nil {
		return err
	}

	pods, err := c.client.CoreV1().Pods(v1.NamespaceAll).List(ctx, metav1.ListOptions{
		FieldSelector: fields.SelectorFromSet(fields.Set{"spec.nodeName": node}).String(),
	})
	if err != nil {
		return err
	}

	var returnError error
	for _, pod := range pods.Items {
		// Skip mirror pods
		if _, ok := pod.Annotations[v1.MirrorPodAnnotationKey]; ok {
			continue
		}
		// Skip daemonsets
		controller := metav1.GetControllerOf(&pod)
		if controller != nil && controller.Kind == "DaemonSet" {
			continue
		}

		err = c.client.PolicyV1().Evictions(pod.GetNamespace()).Evict(ctx, &policyv1.Eviction{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "policy/v1",
				Kind:       "Eviction",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      pod.GetName(),
				Namespace: pod.GetNamespace(),
			},
			DeleteOptions: metav1.NewDeleteOptions(*pod.Spec.TerminationGracePeriodSeconds),
		})
		if err != nil {
			slog.Info("Failed to evict pod", "err", err, slog.String("node", node), slog.String("pod", pod.GetName()), slog.String("namespace", pod.GetNamespace()))
			returnError = NewErrorFailedToEvictAllPods()
			continue
		}
		slog.Info("Evicted pod", slog.String("node", node), slog.String("pod", pod.GetName()), slog.String("namespace", pod.GetNamespace()))

		done := false
		select {
		case <-ctx.Done():
			slog.Error("Aborting node drain", slog.String("node", node), "err", ctx.Err())
			if returnError == nil {
				returnError = ctx.Err()
			}
			done = true
		default:
		}

		if done {
			break
		}
	}

	return returnError
}

// Find the node in the cluster with the matching machine id
func (c *Client) FindNodeByZincatiID(zincatiID string) (string, error) {
	nodes, err := c.client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	for _, node := range nodes.Items {
		machineID := node.Status.NodeInfo.MachineID
		appID, err := systemdutils.ZincatiMachineID(machineID)
		if err != nil {
			return "", err
		}

		if appID == zincatiID {
			slog.Info("Matched node with zincati app id", slog.String("node", node.GetName()), slog.String("appid", zincatiID))
			return node.Name, nil
		}
	}

	return "", nil
}

// Uncordon a node
func (c *Client) UncordonNode(node string) error {
	_, err := c.client.CoreV1().Nodes().Patch(context.Background(), node, types.MergePatchType, nodeUnschedulablePatch(false), metav1.PatchOptions{})
	if err != nil {
		return err
	}

	return NewLease(drainLeaseName(node), c.client.CoordinationV1().Leases(c.namespace)).Delete(context.Background())
}

// Check if a node has been drained
func (c *Client) IsDrained(node string) (bool, error) {
	ctx := context.Background()
	lease := NewLease(drainLeaseName(node), c.client.CoordinationV1().Leases(c.namespace))
	done, err := lease.IsDone(ctx)
	if err != nil || done {
		return done, err
	}
	fails, err := lease.GetFailCounter(ctx)
	if err != nil || fails == 0 {
		return false, err
	}

	if c.drainRetries > 0 && fails >= c.drainRetries {
		slog.Info("Exhausted retries for draining node, marking as drained", slog.String("node", node), slog.Int("fails", fails), slog.Int("maxRetries", c.drainRetries))
		return true, nil
	}
	return false, nil
}

package k8s

import (
	"context"
	"log/slog"
	"time"

	"github.com/heathcliff26/fleetlock/pkg/k8s/utils"
	systemdutils "github.com/heathcliff26/fleetlock/pkg/systemd-utils"

	coordv1 "k8s.io/api/coordination/v1"
	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

type Client struct {
	client    kubernetes.Interface
	namespace string
}

// Create a new kubernetes client, defaults to in-cluster if no kubeconfig is provided
func NewClient(kubeconfig string) (*Client, error) {
	client, err := utils.CreateNewClientset(kubeconfig)
	if err == rest.ErrNotInCluster {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	ns, err := utils.GetNamespace()
	if err != nil {
		return nil, err
	}

	return &Client{
		client:    client,
		namespace: ns,
	}, nil
}

// Create a test client with a fake kubernetes clientset
func NewFakeClient() (*Client, *fake.Clientset) {
	fakeclient := fake.NewSimpleClientset()
	return &Client{
		client:    fakeclient,
		namespace: "fleetlock",
	}, fakeclient
}

// Drain a node from all pods and set it to unschedulable.
// Status will be tracked in lease, only one drain will be run at a time.
func (c *Client) DrainNode(node string) error {
	lease, err := c.client.CoordinationV1().Leases(c.namespace).Get(context.TODO(), drainLeaseName(node), metav1.GetOptions{})
	if errors.IsNotFound(err) {
		lease = &coordv1.Lease{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: c.namespace,
				Name:      drainLeaseName(node),
			},
			Spec: coordv1.LeaseSpec{
				HolderIdentity:       utils.Pointer("draining"),
				LeaseDurationSeconds: utils.Pointer(int32(300)),
				AcquireTime:          &metav1.MicroTime{Time: time.Now()},
			},
		}

		lease, err = c.client.CoordinationV1().Leases(c.namespace).Create(context.TODO(), lease, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if lease.Spec.AcquireTime != nil && time.Now().After(lease.Spec.AcquireTime.Time.Add(5*time.Minute)) {
		lease.Spec.AcquireTime = &metav1.MicroTime{Time: time.Now()}
		lease, err = c.client.CoordinationV1().Leases(c.namespace).Update(context.TODO(), lease, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	} else {
		return NewErrorDrainIsLocked()
	}

	err = c.drainNode(node)
	if err != nil {
		return err
	}

	_, err = c.client.CoordinationV1().Leases(c.namespace).Patch(context.TODO(), lease.GetName(), types.MergePatchType, []byte("{\"spec\":{\"holderIdentity\":\"done\"}}"), metav1.PatchOptions{})
	return err
}

// Drain a node of all pods, skipping daemonsets
func (c *Client) drainNode(node string) error {
	_, err := c.client.CoreV1().Nodes().Patch(context.TODO(), node, types.MergePatchType, nodeUnschedulablePatch(true), metav1.PatchOptions{})
	if err != nil {
		return err
	}

	pods, err := c.client.CoreV1().Pods(v1.NamespaceAll).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fields.SelectorFromSet(fields.Set{"spec.nodeName": node}).String(),
	})
	if err != nil {
		return err
	}

	var returnError error
	for _, pod := range pods.Items {
		// Skip mirror pods
		if _, ok := pod.ObjectMeta.Annotations[v1.MirrorPodAnnotationKey]; ok {
			continue
		}
		// Skip daemonsets
		controller := metav1.GetControllerOf(&pod)
		if controller != nil && controller.Kind == "DaemonSet" {
			continue
		}

		err = c.client.PolicyV1().Evictions(pod.GetNamespace()).Evict(context.TODO(), &policyv1.Eviction{
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
	}

	return returnError
}

// Find the node in the cluster with the matching machine id
func (c *Client) FindNodeByZincatiID(zincatiID string) (string, error) {
	nodes, err := c.client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
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
	_, err := c.client.CoreV1().Nodes().Patch(context.TODO(), node, types.MergePatchType, nodeUnschedulablePatch(false), metav1.PatchOptions{})
	if err != nil {
		return err
	}
	err = c.client.CoordinationV1().Leases(c.namespace).Delete(context.TODO(), drainLeaseName(node), metav1.DeleteOptions{})
	if errors.IsNotFound(err) {
		return nil
	} else {
		return err
	}
}

// Check if a node has been drained
func (c *Client) IsDrained(node string) (bool, error) {
	lease, err := c.client.CoordinationV1().Leases(c.namespace).Get(context.TODO(), drainLeaseName(node), metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return *lease.Spec.HolderIdentity == "done", nil
}

package kubernetes

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/heathcliff26/fleetlock/pkg/k8s/utils"
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/types"

	coordv1 "k8s.io/api/coordination/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	v1 "k8s.io/client-go/kubernetes/typed/coordination/v1"
)

const keyformat = "fleetlock-reservation-%s-"

type KubernetesBackend struct {
	client    v1.CoordinationV1Interface
	namespace string
}

type KubernetesConfig struct {
	Kubeconfig string `yaml:"-"`
	Namespace  string `yaml:"namespace,omitempty"`
}

func NewKubernetesBackend(cfg KubernetesConfig) (*KubernetesBackend, error) {
	client, err := utils.CreateNewClientset(cfg.Kubeconfig)
	if err != nil {
		return nil, err
	}

	ns := cfg.Namespace
	if ns == "" {
		ns, err = utils.GetNamespace()
		if err != nil {
			return nil, err
		}
	}

	return &KubernetesBackend{
		client:    client.CoordinationV1(),
		namespace: ns,
	}, nil
}

// Create a test client with a fake kubernetes clientset
func NewKubernetesBackendWithFakeClient(namespace string) (*KubernetesBackend, *fake.Clientset) {
	fakeclient := fake.NewSimpleClientset()
	return &KubernetesBackend{
		client: fakeclient.CoordinationV1(),
	}, fakeclient
}

// Reserve a lock for the given group.
// Returns true if the lock is successfully reserved, even if the lock is already held by the specific id
func (k *KubernetesBackend) Reserve(group string, id string) error {
	// Kubernetes names do not allow uppercase
	group = strings.ToLower(group)
	leases, err := k.getLeasesForGroup(group)
	if err != nil {
		return err
	}

	names := make([]string, 0)
	for _, lease := range leases {
		if *lease.Spec.HolderIdentity == id {
			return nil
		}
		names = append(names, lease.GetName())
	}

	i := 0
	key := fmt.Sprintf(keyformat, group)

	name := key + strconv.Itoa(i)
	for slices.Contains(names, name) {
		i++
		name = key + strconv.Itoa(i)
	}

	lease := &coordv1.Lease{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: k.namespace,
			Name:      name,
		},
		Spec: coordv1.LeaseSpec{
			HolderIdentity: &id,
		},
	}
	_, err = k.client.Leases(k.namespace).Create(context.TODO(), lease, metav1.CreateOptions{})

	return err
}

// Returns the current number of locks for the given group
func (k *KubernetesBackend) GetLocks(group string) (int, error) {
	leases, err := k.getLeasesForGroup(group)
	return len(leases), err
}

// Release the lock currently held by the id.
// Does not fail when no lock is held.
func (k *KubernetesBackend) Release(group string, id string) error {
	leases, err := k.getLeasesForGroup(group)
	if err != nil {
		return err
	}

	for _, lease := range leases {
		if *lease.Spec.HolderIdentity == id {
			return k.client.Leases(k.namespace).Delete(context.TODO(), lease.GetName(), metav1.DeleteOptions{})
		}
	}

	return nil
}

// Return all locks older than x
func (k *KubernetesBackend) GetStaleLocks(ts time.Duration) ([]types.Lock, error) {
	panic("not implemented") // TODO: Implement
}

// Check if a given id already has a lock for this group
func (k *KubernetesBackend) HasLock(group string, id string) (bool, error) {
	leases, err := k.getLeasesForGroup(group)
	if err != nil {
		return false, err
	}

	for _, lease := range leases {
		if *lease.Spec.HolderIdentity == id {
			return true, nil
		}
	}
	return false, nil
}

// Calls all necessary finalization if necessary
func (k *KubernetesBackend) Close() error {
	return nil
}

func (k *KubernetesBackend) getLeasesForGroup(group string) ([]coordv1.Lease, error) {
	// Kubernetes names do not allow uppercase
	group = strings.ToLower(group)
	leases, err := k.client.Leases(k.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	matcher, err := regexp.Compile(fmt.Sprintf(keyformat, group) + "\\d+$")
	if err != nil {
		return nil, err
	}

	result := make([]coordv1.Lease, 0)

	for _, lease := range leases.Items {
		if matcher.MatchString(lease.GetName()) {
			result = append(result, lease)
		}
	}
	return result, nil
}

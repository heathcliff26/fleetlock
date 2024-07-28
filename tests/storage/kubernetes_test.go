package storage

import (
	"context"
	"testing"

	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/kubernetes"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKubernetesBackend(t *testing.T) {
	nsName := "fleetlock"
	storage, client := kubernetes.NewKubernetesBackendWithFakeClient(nsName)
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: nsName,
		},
	}
	_, _ = client.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})

	RunLockManagerTestsuiteWithStorage(t, storage)
}

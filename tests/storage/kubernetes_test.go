package storage

import (
	"testing"

	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/kubernetes"
)

func TestKubernetesBackend(t *testing.T) {
	nsName := "fleetlock"
	storage, _ := kubernetes.NewKubernetesBackendWithFakeClient(nsName)

	RunLockManagerTestsuiteWithStorage(t, storage)
}

package storage

import (
	"testing"

	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/memory"
)

func TestMemoryBackend(t *testing.T) {
	testGroups := GetGroups()
	names := make([]string, len(testGroups))
	i := 0
	for k := range testGroups {
		names[i] = k
		i++
	}

	storage := memory.NewMemoryBackend(names)

	RunLockManagerTestsuiteWithStorage(t, storage)
}

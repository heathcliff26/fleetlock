package storage

import (
	"os"
	"testing"

	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/sql"
)

func TestSQLiteBackend(t *testing.T) {
	cfg := sql.SQLiteConfig{
		File: "test.db",
	}
	storage, err := sql.NewSQLiteBackend(&cfg)
	if err != nil {
		t.Fatalf("Failed to create storage backend: %v", err)
	}
	t.Cleanup(func() {
		err = os.Remove("test.db")
		if err != nil {
			t.Logf("Failed to cleanup sqlite database file: %v", err)
		}
	})

	RunLockManagerTestsuiteWithStorage(t, storage)
}

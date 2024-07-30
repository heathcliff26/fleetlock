package storage

import (
	"testing"

	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/sql"
)

func TestSQLiteBackend(t *testing.T) {
	cfg := sql.SQLiteConfig{
		File: "file:test.db?mode=memory",
	}
	storage, err := sql.NewSQLiteBackend(cfg)
	if err != nil {
		t.Fatalf("Failed to create storage backend: %v", err)
	}

	RunLockManagerTestsuiteWithStorage(t, storage)
}

package storage

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/redis"
)

func TestRedisBackend(t *testing.T) {
	mr := miniredis.RunT(t)

	cfg := redis.RedisConfig{
		Addr: mr.Addr(),
	}

	storage, err := redis.NewRedisBackend(&cfg)
	if err != nil {
		t.Fatalf("Failed to create storage backend: %v", err)
	}

	RunLockManagerTestsuiteWithStorage(t, storage)
}

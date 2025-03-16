package storage

import (
	"testing"
	"time"

	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/mongodb"
	"github.com/heathcliff26/fleetlock/tests/utils"
)

func TestMongoDBBackend(t *testing.T) {
	if !utils.HasContainerRuntimer() {
		t.Skip("Missing Container Runtime")
	}
	t.Parallel()

	// The latest tag on the etcd image is not being updated
	err := utils.ExecCRI("run", "--replace", "--name", "fleetlock-mongodb", "-d", "-p", "27017:27017", "docker.io/library/mongo:latest")
	if err != nil {
		t.Fatalf("Failed to start test db: %v\n", err)
	}
	t.Cleanup(func() {
		_ = utils.ExecCRI("stop", "fleetlock-mongodb")
		_ = utils.ExecCRI("rm", "fleetlock-mongodb")
	})

	cfg := mongodb.MongoDBConfig{
		URL:      "mongodb://localhost:27017/",
		Database: mongodb.DEFAULT_DATABASE,
	}

	var storage *mongodb.MongoDBBackend
	for i := 0; i < 20; {
		storage, err = mongodb.NewMongoDBBackend(cfg)
		if err == nil {
			break
		}
		<-time.After(time.Second)
		i++
	}
	if err != nil {
		t.Fatalf("Failed to create storage backend: %v", err)
	}

	RunLockManagerTestsuiteWithStorage(t, storage)
}

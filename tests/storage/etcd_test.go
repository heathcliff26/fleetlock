package storage

import (
	"testing"
	"time"

	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/etcd"
	"github.com/heathcliff26/fleetlock/tests/utils"
)

func TestEtcdBackend(t *testing.T) {
	if !utils.HasContainerRuntimer() {
		t.Skip("Missing Container Runtime")
	}
	t.Parallel()

	// The latest tag on the etcd image is not being updated
	err := utils.ExecCRI("run", "--name", "fleetlock-etcd-db", "-d", "-p", "2379:2379", "-p", "2380:2380",
		"quay.io/coreos/etcd:v3.5.15",
		"etcd",
		"--listen-client-urls", "http://0.0.0.0:2379",
		"--advertise-client-urls", "http://localhost:2379",
	)
	if err != nil {
		t.Fatalf("Failed to start test db: %v\n", err)
	}
	t.Cleanup(func() {
		_ = utils.ExecCRI("stop", "fleetlock-etcd-db")
		_ = utils.ExecCRI("rm", "fleetlock-etcd-db")
	})

	<-time.After(time.Second * 5)

	cfg := etcd.EtcdConfig{
		Endpoints: []string{"http://localhost:2379"},
	}
	storage, err := etcd.NewEtcdBackend(cfg)
	if err != nil {
		cmd := utils.GetCommand("logs", "fleetlock-etcd-db")
		out, _ := cmd.Output()
		t.Log("logs from etcd:\n" + string(out))

		cmd = utils.GetCommand("ps", "-a")
		out, _ = cmd.Output()
		t.Log("Output of ps -a:\n" + string(out))

		t.Fatalf("Failed to create storage backend: %v", err)
	}

	RunLockManagerTestsuiteWithStorage(t, storage)
}

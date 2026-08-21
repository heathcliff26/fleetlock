package storage

import (
	"log"
	"testing"
	"time"

	"go.etcd.io/etcd/server/v3/embed"

	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/etcd"
	"github.com/heathcliff26/fleetlock/tests/utils"
)

func TestEtcdBackend(t *testing.T) {
	serverCfg := embed.NewConfig()
	serverCfg.Dir = t.TempDir()
	serverCfg.LogLevel = "fatal"
	e, err := embed.StartEtcd(serverCfg)
	if err != nil {
		t.Fatalf("Failed to start etcd: %v", err)
	}
	defer e.Close()

	select {
	case <-e.Server.ReadyNotify():
		log.Printf("Server is ready!")
	case <-time.After(60 * time.Second):
		e.Server.Stop()
		t.Fatalf("Timed out waiting for etcd to be ready: %v", <-e.Err())
	}

	cfg := etcd.EtcdConfig{}
	for _, url := range e.Config().ListenClientUrls {
		cfg.Endpoints = append(cfg.Endpoints, url.String())
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

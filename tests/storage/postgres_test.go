package storage

import (
	"strings"
	"testing"
	"time"

	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/sql"
	"github.com/heathcliff26/fleetlock/tests/utils"
)

func TestPostgresBackend(t *testing.T) {
	if !utils.HasContainerRuntimer() {
		t.Skip("Missing Container Runtime")
	}
	t.Parallel()

	err := utils.ExecCRI("run", "--name", "fleetlock-cockroach-db", "-d", "--rm", "-p", "26257:26257",
		"--env", "COCKROACH_DATABASE=fleetlock", "--env", "COCKROACH_USER=postgres", "--env", "COCKROACH_PASSWORD=password",
		"docker.io/cockroachdb/cockroach:latest",
		"start-single-node", "--http-addr=localhost:8080", "--accept-sql-without-tls",
	)
	if err != nil {
		t.Fatalf("Failed to start test db: %v\n", err)
	}
	t.Cleanup(func() {
		_ = utils.ExecCRI("stop", "fleetlock-cockroach-db")
	})

	var storage *sql.SQLBackend
	cfg := sql.PostgresConfig{
		Address:  "localhost:26257",
		Username: "postgres",
		Password: "password",
		Database: "fleetlock",
	}
	for i := 0; i < 10; {
		storage, err = sql.NewPostgresBackend(cfg)
		if err == nil || (!strings.Contains(err.Error(), "failed to open postgres database") && !strings.Contains(err.Error(), "failed to ping postgres database")) {
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

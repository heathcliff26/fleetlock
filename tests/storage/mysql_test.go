package storage

import (
	"reflect"
	"testing"
	"time"

	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/sql"
	"github.com/heathcliff26/fleetlock/tests/utils"
)

func TestMySQLBackend(t *testing.T) {
	if !utils.HasContainerRuntimer() {
		t.Skip("Missing Container Runtime")
	}

	err := utils.ExecCRI("run", "--name", "fleetlock-mysql-db", "-d", "--rm", "-p", "3306:3306",
		"--env", "MYSQL_ROOT_PASSWORD=password", "--env", "MYSQL_DATABASE=fleetlock",
		"docker.io/library/mysql:latest",
	)
	if err != nil {
		t.Fatalf("Failed to start test db: %v\n", err)
	}
	t.Cleanup(func() {
		_ = utils.ExecCRI("stop", "fleetlock-mysql-db")
	})

	var storage *sql.SQLBackend
	cfg := sql.MySQLConfig{
		Address:  "tcp(localhost:3306)",
		Username: "root",
		Password: "password",
		Database: "fleetlock",
	}
	for i := 0; i < 20; {
		storage, err = sql.NewMySQLBackend(&cfg)
		if err == nil || (reflect.TypeOf(err).String() != "*errors.errorString" && err.Error() != "driver: bad connection") {
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

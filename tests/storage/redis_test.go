package storage

import (
	"net"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/redis"
	"github.com/heathcliff26/fleetlock/tests/utils"
)

func TestRedisBackend(t *testing.T) {
	mr := miniredis.RunT(t)

	cfg := redis.RedisConfig{
		Addr: mr.Addr(),
	}

	storage, err := redis.NewRedisBackend(cfg)
	if err != nil {
		t.Fatalf("Failed to create storage backend: %v", err)
	}

	RunLockManagerTestsuiteWithStorage(t, storage)
}

func TestRedisLoadbalancerBackend(t *testing.T) {
	mr1 := miniredis.RunT(t)
	mr2 := miniredis.RunT(t)

	cfg := redis.RedisConfig{
		Addrs: []string{mr1.Addr(), mr2.Addr()},
	}

	storage, err := redis.NewRedisBackend(cfg)
	if err != nil {
		t.Fatalf("Failed to create storage backend: %v", err)
	}

	RunLockManagerTestsuiteWithStorage(t, storage)
}

func TestRedisSentinelBackend(t *testing.T) {
	if !utils.HasContainerRuntimer() {
		t.Skip("Missing Container Runtime")
	}

	err := utils.ExecCRI("run", "--name", "fleetlock-redis-sentinel-db", "-d", "--net", "host",
		"docker.io/valkey/valkey:latest",
		"--port", "6379",
	)
	if err != nil {
		t.Fatalf("Failed to start test db: %v\n", err)
	}
	t.Cleanup(func() {
		_ = utils.ExecCRI("stop", "fleetlock-redis-sentinel-db")
		_ = utils.ExecCRI("rm", "fleetlock-redis-sentinel-db")
	})

	err = utils.ExecCRI("run", "--name", "fleetlock-redis-sentinel-sentinel", "-d", "--net", "host",
		"-v", "./testdata/redis-sentinel.conf:/config/sentinel.conf", "--userns=keep-id",
		"docker.io/valkey/valkey:latest",
		"/config/sentinel.conf", "--sentinel",
	)
	if err != nil {
		t.Fatalf("Failed to start test sentinel: %v\n", err)
	}
	t.Cleanup(func() {
		_ = utils.ExecCRI("stop", "fleetlock-redis-sentinel-sentinel")
		_ = utils.ExecCRI("rm", "fleetlock-redis-sentinel-sentinel")
	})

	for i := 0; i < 10; {
		conn, err := net.Dial("tcp", "localhost:26379")
		if err == nil {
			conn.Close()
			break
		}
		<-time.After(time.Second)
		i++
	}

	cfg := redis.RedisConfig{
		Sentinel: redis.RedisSentinelConfig{
			Enabled:    true,
			MasterName: "redis-sentinel-backend",
			Addresses:  []string{"localhost:26379"},
		},
	}

	storage, err := redis.NewRedisBackend(cfg)
	if err != nil {
		cmd := utils.GetCommand("logs", "fleetlock-redis-sentinel-sentinel")
		out, _ := cmd.Output()
		t.Log("logs from sentinel:\n" + string(out))
		cmd = utils.GetCommand("logs", "fleetlock-redis-sentinel-db")
		out, _ = cmd.Output()
		t.Log("logs from db:\n" + string(out))

		cmd = utils.GetCommand("ps", "-a")
		out, _ = cmd.Output()
		t.Log("Output of ps -a:\n" + string(out))

		t.Fatalf("Failed to create storage backend: %v", err)
	}

	RunLockManagerTestsuiteWithStorage(t, storage)
}

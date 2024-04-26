package redis

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/heathcliff26/fleetlock/tests/utils"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestLoadbalancer(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		mr1 := miniredis.RunT(t)
		mr2 := miniredis.RunT(t)

		addrs := []string{mr1.Addr(), mr2.Addr()}

		client, lb := NewRedisClientWithLoadbalancer(addrs, &redis.Options{})
		t.Cleanup(func() {
			lb.Close()
			client.Close()
		})

		assert := assert.New(t)

		res, err := client.Ping(context.Background()).Result()
		if !assert.Nil(err, "Can reach client") || !assert.Equal("PONG", res, "Can reach client") {
			t.FailNow()
		}
	})
	t.Run("Failover", func(t *testing.T) {
		if !utils.HasContainerRuntimer() {
			t.Skip("Missing Container Runtime")
		}

		err := utils.ExecCRI("run", "--name", "fleetlock-redis-loadbalancer-failover-1", "-d", "--rm", "--net", "host", "docker.io/eqalpha/keydb:latest", "--port", "6379", "--active-replica", "yes", "--replicaof", "localhost", "6380")
		if err != nil {
			t.Fatalf("Failed to start test db: %v\n", err)
		}
		t.Cleanup(func() {
			_ = utils.ExecCRI("stop", "fleetlock-redis-loadbalancer-failover-1")
		})
		err = utils.ExecCRI("run", "--name", "fleetlock-redis-loadbalancer-failover-2", "-d", "--rm", "--net", "host", "docker.io/eqalpha/keydb:latest", "--port", "6380", "--active-replica", "yes", "--replicaof", "localhost", "6379")
		if err != nil {
			t.Fatalf("Failed to start test db: %v\n", err)
		}
		t.Cleanup(func() {
			_ = utils.ExecCRI("stop", "fleetlock-redis-loadbalancer-failover-2")
		})

		addrs := []string{"localhost:6379", "localhost:6380"}
		client, lb := NewRedisClientWithLoadbalancer(addrs, &redis.Options{})
		t.Cleanup(func() {
			lb.Close()
			client.Close()
		})

		assert := assert.New(t)

		assert.Equal(0, lb.selected, "Should have currently the first client selected")

		err = utils.ExecCRI("stop", "fleetlock-redis-loadbalancer-failover-1")
		if err != nil {
			t.Fatalf("Failed to stop keydb instance: %v\n", err)
		}
		lb.HealthCheck()

		if !assert.Equal(1, lb.selected, "Should have failed over") {
			t.FailNow()
		}
		res, err := client.Ping(context.Background()).Result()
		if !assert.Nil(err, "Should have failed over") || !assert.Equal("PONG", res, "Should have failed over") {
			t.FailNow()
		}
	})
}

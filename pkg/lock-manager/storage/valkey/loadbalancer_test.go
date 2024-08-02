package valkey

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/heathcliff26/fleetlock/tests/utils"
	"github.com/stretchr/testify/assert"
	"github.com/valkey-io/valkey-go"
)

func TestLoadbalancer(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		mr1 := miniredis.RunT(t)
		mr2 := miniredis.RunT(t)

		opt := valkey.ClientOption{
			InitAddress:  []string{mr1.Addr(), mr2.Addr()},
			DisableCache: true,
		}

		assert := assert.New(t)

		client, lb, err := NewValkeyLoadbalancer(opt)
		if !assert.NoError(err, "Should not return an error") || !assert.NotNil(client, "Should return a client") || !assert.NotNil(lb, "Should return a loadbalancer") {
			t.FailNow()
		}
		t.Cleanup(func() {
			lb.Close()
			client.Close()
		})

		res, err := client.Do(context.Background(), client.B().Ping().Build()).ToString()

		assert.Nil(err, "Can reach client")
		assert.Equal("PONG", res, "Can reach client")
	})
	t.Run("Failover", func(t *testing.T) {
		if !utils.HasContainerRuntimer() {
			t.Skip("Missing Container Runtime")
		}

		err := utils.ExecCRI("run", "--name", "fleetlock-valkey-loadbalancer-failover-1", "-d", "--net", "host", "docker.io/eqalpha/keydb:latest", "--port", "6379", "--active-replica", "yes", "--replicaof", "localhost", "6380")
		if err != nil {
			t.Fatalf("Failed to start test db: %v\n", err)
		}
		t.Cleanup(func() {
			_ = utils.ExecCRI("stop", "fleetlock-valkey-loadbalancer-failover-1")
			_ = utils.ExecCRI("rm", "fleetlock-valkey-loadbalancer-failover-1")
		})
		err = utils.ExecCRI("run", "--name", "fleetlock-valkey-loadbalancer-failover-2", "-d", "--net", "host", "docker.io/eqalpha/keydb:latest", "--port", "6380", "--active-replica", "yes", "--replicaof", "localhost", "6379")
		if err != nil {
			t.Fatalf("Failed to start test db: %v\n", err)
		}
		t.Cleanup(func() {
			_ = utils.ExecCRI("stop", "fleetlock-valkey-loadbalancer-failover-2")
			_ = utils.ExecCRI("rm", "fleetlock-valkey-loadbalancer-failover-2")
		})

		assert := assert.New(t)

		opt := valkey.ClientOption{
			InitAddress:  []string{"localhost:6379", "localhost:6380"},
			DisableCache: true,
		}
		client, lb, err := NewValkeyLoadbalancer(opt)
		if !assert.NoError(err, "Should not return an error") || !assert.NotNil(client, "Should return a client") || !assert.NotNil(lb, "Should return a loadbalancer") {
			t.FailNow()
		}
		t.Cleanup(func() {
			lb.Close()
			client.Close()
		})

		assert.Equal(0, lb.selected, "Should have currently the first client selected")

		err = utils.ExecCRI("stop", "fleetlock-valkey-loadbalancer-failover-1")
		if err != nil {
			t.Fatalf("Failed to stop keydb instance: %v\n", err)
		}

		lb.HealthCheck()
		if !assert.Equal(1, lb.selected, "Should have failed over") {
			t.FailNow()
		}

		res, err := client.Do(context.Background(), client.B().Ping().Build()).ToString()

		assert.NoError(err, "Should have failed over")
		assert.Equal("PONG", res, "Should have failed over")
	})
	t.Run("DeadlockCheck", func(t *testing.T) {
		mr1 := miniredis.RunT(t)
		mr2 := miniredis.RunT(t)

		opt := valkey.ClientOption{
			InitAddress:  []string{mr1.Addr(), mr2.Addr()},
			DisableCache: true,
		}

		assert := assert.New(t)

		client, lb, err := NewValkeyLoadbalancer(opt)
		if !assert.NoError(err, "Should not return an error") || !assert.NotNil(client, "Should return a client") || !assert.NotNil(lb, "Should return a loadbalancer") {
			t.FailNow()
		}
		t.Cleanup(func() {
			lb.Close()
			client.Close()
		})

		// Ensure no failover happens automatically
		lb.cancel()

		assert.Equal(0, lb.selected, "Should have currently the first client selected")

		mr1.Close()

		done := make(chan struct{}, 1)

		go func() {
			lb.HealthCheck()
			done <- struct{}{}
			close(done)
		}()
		go func() {
			_, err = client.Do(context.Background(), client.B().Ping().Build()).ToString()

			assert.Error(err, "Call should fail")
		}()

		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Fatal("Timed out waiting for the failover to finished")
		}
	})
}

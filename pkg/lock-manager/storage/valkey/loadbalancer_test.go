package valkey

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		require := require.New(t)

		client, lb, err := NewValkeyLoadbalancer(opt)
		require.NoError(err, "Should not return an error")
		require.NotNil(client, "Should return a client")
		require.NotNil(lb, "Should return a loadbalancer")
		t.Cleanup(func() {
			lb.Close()
			client.Close()
		})

		res, err := client.Do(t.Context(), client.B().Ping().Build()).ToString()

		assert.Nil(err, "Can reach client")
		assert.Equal("PONG", res, "Can reach client")
	})
	t.Run("Failover", func(t *testing.T) {
		t.Parallel()

		assert := assert.New(t)
		require := require.New(t)

		mr1 := miniredis.RunT(t)
		mr2 := miniredis.RunT(t)

		opt := valkey.ClientOption{
			InitAddress:  []string{mr1.Addr(), mr2.Addr()},
			DisableCache: true,
		}
		client, lb, err := NewValkeyLoadbalancer(opt)
		require.NoError(err, "Should not return an error")
		require.NotNil(client, "Should return a client")
		require.NotNil(lb, "Should return a loadbalancer")
		t.Cleanup(func() {
			lb.Close()
			client.Close()
		})

		lb.HealthCheck()

		assert.Equal(0, lb.selected, "Should have currently the first client selected")

		mr1.Close()

		lb.HealthCheck()
		require.Equal(1, lb.selected, "Should have failed over")

		res, err := client.Do(t.Context(), client.B().Ping().Build()).ToString()

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
		require := require.New(t)

		client, lb, err := NewValkeyLoadbalancer(opt)
		require.NoError(err, "Should not return an error")
		require.NotNil(client, "Should return a client")
		require.NotNil(lb, "Should return a loadbalancer")
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
			_, err = client.Do(t.Context(), client.B().Ping().Build()).ToString()

			assert.Error(err, "Call should fail")
		}()

		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Fatal("Timed out waiting for the failover to finished")
		}
	})
}

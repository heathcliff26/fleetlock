package redis

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
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
}

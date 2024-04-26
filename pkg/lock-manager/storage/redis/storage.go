package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/heathcliff26/fleetlock/pkg/lock-manager/types"
	"github.com/redis/go-redis/v9"
)

const keyformat = "group:%s,id:%s"

type RedisBackend struct {
	client *redis.Client
	lb     *loadbalancer
}

type RedisConfig struct {
	Addr     string              `yaml:"address,omitempty"`
	Addrs    []string            `yaml:"addresses,omitempty"`
	Username string              `yaml:"username,omitempty"`
	Password string              `yaml:"password,omitempty"`
	DB       int                 `yaml:"db,omitempty"`
	Sentinel RedisSentinelConfig `yaml:"sentinel,omitempty"`
}

type RedisSentinelConfig struct {
	Enabled    bool     `yaml:"enabled,omitempty"`
	MasterName string   `yaml:"master,omitempty"`
	Addresses  []string `yaml:"addresses,omitempty"`
	Username   string   `yaml:"username,omitempty"`
	Password   string   `yaml:"password,omitempty"`
}

func NewRedisBackend(cfg *RedisConfig) (*RedisBackend, error) {
	var client *redis.Client
	var lb *loadbalancer
	switch {
	case cfg.Sentinel.Enabled:
		client = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:       cfg.Sentinel.MasterName,
			SentinelAddrs:    cfg.Sentinel.Addresses,
			SentinelUsername: cfg.Sentinel.Username,
			SentinelPassword: cfg.Sentinel.Password,
			Username:         cfg.Username,
			Password:         cfg.Password,
			DB:               cfg.DB,
		})
	case len(cfg.Addrs) > 0:
		opt := redis.Options{
			Username: cfg.Username,
			Password: cfg.Password,
			DB:       cfg.DB,
		}
		client, lb = NewRedisClientWithLoadbalancer(cfg.Addrs, &opt)
	default:
		client = redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,
			Username: cfg.Username,
			Password: cfg.Password,
			DB:       cfg.DB,
		})
	}

	err := client.Ping(context.Background()).Err()
	if err != nil {
		return nil, err
	}

	return &RedisBackend{
		client: client,
		lb:     lb,
	}, nil
}

// Reserve a lock for the given group.
// Returns true if the lock is successfully reserved, even if the lock is already held by the specific id
func (r *RedisBackend) Reserve(group string, id string) error {
	key := fmt.Sprintf(keyformat, group, id)
	ctx := context.Background()

	ok, err := r.client.SetNX(ctx, key, time.Now(), 0).Result()
	if err != nil {
		return err
	}

	if ok {
		err := r.client.SAdd(ctx, group, key).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

// Returns the current number of locks for the given group
func (r *RedisBackend) GetLocks(group string) (int, error) {
	result := r.client.SCard(context.Background(), group)
	if err := result.Err(); err != nil {
		return 0, err
	}
	return int(result.Val()), nil
}

// Release the lock currently held by the id.
// Does not fail when no lock is held.
func (r *RedisBackend) Release(group string, id string) error {
	key := fmt.Sprintf(keyformat, group, id)
	ctx := context.Background()

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return err
	}

	err = r.client.SRem(ctx, group, key).Err()
	return err
}

// Return all locks older than x
func (r *RedisBackend) GetStaleLocks(ts time.Duration) ([]types.Lock, error) {
	panic("not implemented") // TODO: Implement
}

// Check if a given id already has a lock for this group
func (r *RedisBackend) HasLock(group string, id string) (bool, error) {
	key := fmt.Sprintf(keyformat, group, id)
	ctx := context.Background()

	count, err := r.client.Exists(ctx, key).Result()
	return count == 1 || err != nil, err
}

// Calls all necessary finalization if necessary
func (r *RedisBackend) Close() error {
	if r.lb != nil {
		r.lb.Close()
	}
	return r.client.Close()
}

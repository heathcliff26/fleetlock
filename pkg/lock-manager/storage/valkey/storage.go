package valkey

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/heathcliff26/fleetlock/pkg/lock-manager/types"
	"github.com/valkey-io/valkey-go"
)

const keyformat = "group:%s,id:%s"

type ValkeyBackend struct {
	client valkey.Client
	lb     *loadbalancer
}

type ValkeyConfig struct {
	Addrs    []string             `json:"addresses,omitempty"`
	Username string               `json:"username,omitempty"`
	Password string               `json:"password,omitempty"`
	DB       int                  `json:"db,omitempty"`
	TLS      bool                 `json:"tls,omitempty"`
	Sentinel ValkeySentinelConfig `json:"sentinel,omitempty"`
}

type ValkeySentinelConfig struct {
	Enabled    bool     `json:"enabled,omitempty"`
	MasterName string   `json:"master,omitempty"`
	Addresses  []string `json:"addresses,omitempty"`
	Username   string   `json:"username,omitempty"`
	Password   string   `json:"password,omitempty"`
}

func NewValkeyBackend(cfg ValkeyConfig) (*ValkeyBackend, error) {
	var client valkey.Client
	var lb *loadbalancer
	var tlsConfig *tls.Config

	if cfg.TLS {
		tlsConfig = &tls.Config{}
	}

	opt := valkey.ClientOption{
		InitAddress: cfg.Addrs,
		Username:    cfg.Username,
		Password:    cfg.Password,
		SelectDB:    cfg.DB,
		TLSConfig:   tlsConfig,

		DisableCache: true,
	}

	var err error
	switch {
	case cfg.Sentinel.Enabled:
		opt.Sentinel = valkey.SentinelOption{
			MasterSet: cfg.Sentinel.MasterName,
			Username:  cfg.Sentinel.Username,
			Password:  cfg.Sentinel.Password,
		}
		opt.InitAddress = cfg.Sentinel.Addresses

		client, err = valkey.NewClient(opt)
	case len(cfg.Addrs) > 1:
		client, lb, err = NewValkeyLoadbalancer(opt)
	default:
		client, err = valkey.NewClient(opt)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to valkey server: %v", err)
	}

	return &ValkeyBackend{
		client: client,
		lb:     lb,
	}, nil
}

// Reserve a lock for the given group.
// Returns true if the lock is successfully reserved, even if the lock is already held by the specific id
func (r *ValkeyBackend) Reserve(group string, id string) error {
	key := fmt.Sprintf(keyformat, group, id)
	ctx := context.Background()

	cmdSetNX := r.client.B().Setnx().Key(key).Value(time.Now().String()).Build()
	cmdSAdd := r.client.B().Sadd().Key(group).Member(key).Build()

	ok, err := r.client.Do(ctx, cmdSetNX).AsBool()
	if err != nil {
		return fmt.Errorf("failed to create key: %w", err)
	}

	if ok {
		err := r.client.Do(ctx, cmdSAdd).Error()
		if err != nil {
			return fmt.Errorf("failed to add key to group list: %w", err)
		}
	}

	return nil
}

// Returns the current number of locks for the given group
func (r *ValkeyBackend) GetLocks(group string) (int, error) {
	cmdSCard := r.client.B().Scard().Key(group).Build()
	result, err := r.client.Do(context.Background(), cmdSCard).AsInt64()
	if err != nil {
		return 0, fmt.Errorf("failed to get locks from database: %w", err)
	}
	return int(result), nil
}

// Release the lock currently held by the id.
// Does not fail when no lock is held.
func (r *ValkeyBackend) Release(group string, id string) error {
	key := fmt.Sprintf(keyformat, group, id)
	ctx := context.Background()

	cmdDel := r.client.B().Del().Key(key).Build()
	cmdSRem := r.client.B().Srem().Key(group).Member(key).Build()

	err := r.client.Do(ctx, cmdDel).Error()
	if err != nil {
		return fmt.Errorf("failed to delete key in database: %w", err)
	}

	err = r.client.Do(ctx, cmdSRem).Error()
	if err != nil {
		return fmt.Errorf("failed to remove key from group: %w", err)
	}
	return nil
}

// Return all locks older than x
func (r *ValkeyBackend) GetStaleLocks(ts time.Duration) ([]types.Lock, error) {
	panic("not implemented") // TODO: Implement
}

// Check if a given id already has a lock for this group
func (r *ValkeyBackend) HasLock(group string, id string) (bool, error) {
	key := fmt.Sprintf(keyformat, group, id)
	ctx := context.Background()

	cmdExists := r.client.B().Exists().Key(key).Build()
	count, err := r.client.Do(ctx, cmdExists).AsInt64()
	if err != nil {
		return false, fmt.Errorf("failed to count keys in group: %w", err)
	}
	return count == 1, nil
}

// Calls all necessary finalization if necessary
func (r *ValkeyBackend) Close() error {
	if r.lb != nil {
		r.lb.Close()
	}
	r.client.Close()
	return nil
}

package etcd

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/heathcliff26/fleetlock/pkg/lock-manager/types"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const keyformat = "com.github.heathcliff26.fleetlock/group/%s/id/%s"

const timeout = 200 * time.Millisecond

type EtcdBackend struct {
	client *clientv3.Client
}

type EtcdConfig struct {
	Endpoints []string `yaml:"endpoints,omitempty"`
	Username  string   `yaml:"username,omitempty"`
	Password  string   `yaml:"password,omitempty"`
	CertFile  string   `yaml:"cert,omitempty"`
	KeyFile   string   `yaml:"key,omitempty"`
}

func NewEtcdBackend(cfg *EtcdConfig) (*EtcdBackend, error) {
	var tls *tls.Config
	if cfg.CertFile != "" && cfg.KeyFile != "" {
		var err error
		tls, err = transport.TLSInfo{
			CertFile: cfg.CertFile,
			KeyFile:  cfg.KeyFile,
		}.ClientConfig()
		if err != nil {
			return nil, err
		}
	}
	c, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Endpoints,
		Username:    cfg.Username,
		Password:    cfg.Password,
		DialTimeout: time.Second,
		TLS:         tls,
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_, err = c.MemberList(ctx)
	if err != nil {
		return nil, err
	}

	return &EtcdBackend{
		client: c,
	}, nil
}

// Reserve a lock for the given group.
// Returns true if the lock is successfully reserved, even if the lock is already held by the specific id
func (e *EtcdBackend) Reserve(group string, id string) error {
	key := fmt.Sprintf(keyformat, group, id)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := e.client.Txn(ctx).If(
		clientv3.Compare(clientv3.Version(key), "=", 0),
	).Then(
		clientv3.OpPut(key, time.Now().String()),
	).Commit()

	if err != nil {
		return err
	}

	return nil
}

// Returns the current number of locks for the given group
func (e *EtcdBackend) GetLocks(group string) (int, error) {
	key := fmt.Sprintf(keyformat, group, "")
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := e.client.Get(ctx, key, clientv3.WithPrefix(), clientv3.WithCountOnly())
	if err != nil {
		return 0, err
	}
	return int(res.Count), nil
}

// Release the lock currently held by the id.
// Does not fail when no lock is held.
func (e *EtcdBackend) Release(group string, id string) error {
	key := fmt.Sprintf(keyformat, group, id)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := e.client.Delete(ctx, key)
	return err
}

// Return all locks older than x
func (e *EtcdBackend) GetStaleLocks(ts time.Duration) ([]types.Lock, error) {
	panic("not implemented") // TODO: Implement
}

// Check if a given id already has a lock for this group
func (e *EtcdBackend) HasLock(group string, id string) (bool, error) {
	key := fmt.Sprintf(keyformat, group, id)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := e.client.Get(ctx, key, clientv3.WithCountOnly())
	if err != nil || res == nil {
		return false, err
	}
	return res.Count == 1, nil
}

// Calls all necessary finalization if necessary
func (e *EtcdBackend) Close() error {
	return e.client.Close()
}

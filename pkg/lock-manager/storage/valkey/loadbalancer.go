package valkey

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/valkey-io/valkey-go"
)

const loadbalancerHealtchCheckPeriod = time.Second * 10

type loadbalancer struct {
	// List of addresses for valkey endpoints
	addrs []string
	// Options for connecting to valkey
	options valkey.ClientOption
	// Context to cancle health check
	ctx    context.Context
	cancel context.CancelFunc

	client   valkey.Client
	selected int
	rwlock   sync.RWMutex
}

// Create a new valkey client with loadbalanced connections
func NewValkeyLoadbalancer(opt valkey.ClientOption) (valkey.Client, *loadbalancer, error) {
	opt.ForceSingleClient = true

	ctx, cancel := context.WithCancel(context.Background())
	lb := &loadbalancer{
		addrs:   opt.InitAddress,
		options: opt,
		ctx:     ctx,
		cancel:  cancel,
	}

	opt.DialFn = lb.DialFn

	client, err := valkey.NewClient(opt)
	if err != nil {
		return nil, nil, err
	}

	lb.client = client
	lb.PeriodicHealthCheck()

	return client, lb, nil
}

// Determine the first healthy master node
func (lb *loadbalancer) HealthCheck() {
	found := false

	for i, addr := range lb.addrs {
		opt := lb.options
		opt.InitAddress = []string{addr}

		client, err := valkey.NewClient(opt)
		if err != nil {
			slog.Debug("Endpoint down", slog.String("addr", addr), "err", err)
			continue
		}
		defer client.Close()

		cmdInfo := client.B().Info().Build()
		res, err := client.Do(context.Background(), cmdInfo).ToString()
		if err != nil {
			slog.Error("Failed to get endpoint info", slog.String("addr", addr), "err", err)
			continue
		}
		s := strings.Split(res, "\r\n")
		if slices.Contains(s, "role:master") || slices.Contains(s, "role:active-replica") {
			lb.rwlock.Lock()
			defer lb.rwlock.Unlock()

			if lb.selected != i {
				lb.selected = i
				slog.Info("Failed over to new database", slog.String("addr", addr))
				// valkey keeps a connection. Try to ping it to ensure it gets terminated and the next try will be a new connection
				_ = lb.client.Do(context.Background(), client.B().Ping().Build())
			}

			found = true
			break
		}
	}

	if found {
		slog.Info("Could not connect to any valkey database, all connections are down")
	}
}

// Starts go-routine that periodically runs a healthcheck in the background
func (lb *loadbalancer) PeriodicHealthCheck() {
	go lb.periodicHealthCheck()
}

func (lb *loadbalancer) periodicHealthCheck() {
	for {
		lb.HealthCheck()

		select {
		case <-lb.ctx.Done():
			return
		case <-time.After(loadbalancerHealtchCheckPeriod):
		}
	}
}

func (lb *loadbalancer) DialFn(_ string, dialer *net.Dialer, cfg *tls.Config) (conn net.Conn, err error) {
	lb.rwlock.RLock()
	defer lb.rwlock.RUnlock()
	dst := lb.addrs[lb.selected]

	if cfg != nil {
		return tls.DialWithDialer(dialer, "tcp", dst, cfg)
	}
	return dialer.Dial("tcp", dst)
}

func (lb *loadbalancer) Close() {
	lb.cancel()
}

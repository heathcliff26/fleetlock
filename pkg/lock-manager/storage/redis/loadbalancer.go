package redis

import (
	"context"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const loadbalancerHealtchCheckPeriod = time.Second * 10

type loadbalancer struct {
	// List of addresses for redis endpoints
	addrs []string
	// Options for connecting to redis
	options redis.Options
	// Context to cancle health check
	ctx    context.Context
	cancel context.CancelFunc

	selected int
	rwlock   sync.RWMutex
}

// Create a new redis client with loadbalanced connections
func NewRedisClientWithLoadbalancer(addrs []string, opt *redis.Options) (*redis.Client, *loadbalancer) {
	lb := NewRedisLoadbalancer(addrs, *opt)
	lb.PeriodicHealthCheck()
	opt.Dialer = lb.Dial
	return redis.NewClient(opt), lb
}

// Return a new loadbalancer for redis
func NewRedisLoadbalancer(addrs []string, opt redis.Options) *loadbalancer {
	ctx, cancel := context.WithCancel(context.Background())

	if opt.Network == "" {
		opt.Network = "tcp"
	}

	return &loadbalancer{
		addrs:   addrs,
		options: opt,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Determine the first healthy master node
func (lb *loadbalancer) HealthCheck() {
	for i, addr := range lb.addrs {
		opt := lb.options
		opt.Addr = addr
		client := redis.NewClient(&opt)
		defer client.Close()

		res, err := client.Ping(context.Background()).Result()
		if err != nil || res != "PONG" {
			continue
		}

		res, err = client.Info(context.Background(), "replication").Result()
		if err == nil || strings.Contains(res, "role:master") {
			lb.rwlock.Lock()
			defer lb.rwlock.Unlock()
			lb.selected = i
			break
		}
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

func (lb *loadbalancer) Dial(ctx context.Context, _ string, _ string) (net.Conn, error) {
	lb.rwlock.RLock()
	defer lb.rwlock.RUnlock()
	return net.Dial(lb.options.Network, lb.addrs[lb.selected])
}

func (lb *loadbalancer) Close() {
	lb.cancel()
}

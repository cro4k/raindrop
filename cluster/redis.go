package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/cro4k/raindrop/v1/core"
	"github.com/redis/go-redis/v9"
)

type RedisClusterService struct {
	client  redis.UniversalClient
	prefix  string
	timeout time.Duration
}

type RedisClusterServiceOption func(s *RedisClusterService)

func WithPrefix(prefix string) RedisClusterServiceOption {
	return func(s *RedisClusterService) {
		s.prefix = prefix
	}
}
func WithTimeout(timeout time.Duration) RedisClusterServiceOption {
	return func(s *RedisClusterService) {
		s.timeout = timeout
	}
}

func (s *RedisClusterService) Register(ctx context.Context, id, host string, cc core.Healthy) error {
	go s.ping(id, host, cc)
	return s.register(ctx, id, host)
}

func (s *RedisClusterService) Deregister(ctx context.Context, id string) error {
	return s.deregister(ctx, id)
}

func (s *RedisClusterService) Discover(ctx context.Context, id string) (string, error) {
	key := fmt.Sprintf("%s%s", s.prefix, id)
	return s.client.Get(ctx, key).Result()
}

func (s *RedisClusterService) register(ctx context.Context, id, host string) error {
	key := fmt.Sprintf("%s%s", s.prefix, id)
	return s.client.Set(ctx, key, host, s.timeout).Err()
}

func (s *RedisClusterService) deregister(ctx context.Context, id string) error {
	key := fmt.Sprintf("%s%s", s.prefix, id)
	return s.client.Del(ctx, key).Err()
}

func (s *RedisClusterService) ping(id, host string, cc core.Healthy) {
	defer s.deregister(context.Background(), id)
	timeout := s.timeout
	if timeout > time.Second {
		timeout = timeout - time.Second // make sure refresh the alive time before the redis key is expired
	}
	tick := time.NewTicker(timeout)
	for {
		<-tick.C
		if !cc.IsAlive() {
			_ = s.register(context.Background(), id, host)
			return
		}
	}
}

// NewRedisClusterService is a simple implementation of GRPCService.
// We don't suggest use it in produce environment, you'd better make your own implementation.
func NewRedisClusterService(client redis.UniversalClient, options ...RedisClusterServiceOption) GRPCService {
	s := &RedisClusterService{
		client:  client,
		timeout: 30 * time.Second,
	}
	for _, opt := range options {
		opt(s)
	}
	return s
}

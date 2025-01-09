package registry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/cro4k/raindrop/core"
	"github.com/redis/go-redis/v9"
)

type RedisRegistry struct {
	client  redis.UniversalClient
	prefix  string
	timeout time.Duration
}

type RedisRegistryOption func(*RedisRegistry)

func WithPrefix(prefix string) RedisRegistryOption {
	return func(r *RedisRegistry) {
		r.prefix = prefix
	}
}

func WithTimeout(timeout time.Duration) RedisRegistryOption {
	return func(r *RedisRegistry) {
		r.timeout = timeout
	}
}

func NewRedisRegistry(client redis.UniversalClient, options ...RedisRegistryOption) *RedisRegistry {
	re := &RedisRegistry{
		client:  client,
		prefix:  "RAINDROP_CLIENTS:",
		timeout: 30 * time.Second,
	}
	for _, option := range options {
		option(re)
	}
	return re
}

func (s *RedisRegistry) Register(ctx context.Context, id, host string, cc core.Healthy) error {
	go s.ping(id, host, cc)
	return s.register(ctx, id, host)
}

func (s *RedisRegistry) Deregister(ctx context.Context, id string) error {
	return s.deregister(ctx, id)
}

func (s *RedisRegistry) Discover(ctx context.Context, id string) (string, error) {
	key := fmt.Sprintf("%s%s", s.prefix, id)
	val, err := s.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", core.ErrClientConnectionNotFound
	}
	return val, err
}

func (s *RedisRegistry) register(ctx context.Context, id, host string) error {
	key := fmt.Sprintf("%s%s", s.prefix, id)
	return s.client.Set(ctx, key, host, s.timeout).Err()
}

func (s *RedisRegistry) deregister(ctx context.Context, id string) error {
	key := fmt.Sprintf("%s%s", s.prefix, id)
	return s.client.Del(ctx, key).Err()
}

func (s *RedisRegistry) ping(id, host string, cc core.Healthy) {
	defer func(s *RedisRegistry, ctx context.Context, id string) {
		err := s.deregister(ctx, id)
		if err != nil {
			slog.ErrorContext(ctx, "redis error", "error", err)
		}
	}(s, context.Background(), id)
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

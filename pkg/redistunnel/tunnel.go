package redistunnel

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	KEY = "RAINDROP"

	DefaultRenewDuration = 5 * time.Minute
	DefaultExpiration    = 6 * time.Minute
)

type TunnelConnector struct {
	client redis.UniversalClient
	rg     RunnerGroup

	prefix        string
	renewDuration time.Duration
	expiration    time.Duration
}

func NewTunnelConnector(client redis.UniversalClient, options ...TunnelConnectorOption) *TunnelConnector {
	ts := &TunnelConnector{client: client}
	for _, opt := range options {
		opt(ts)
	}
	if ts.renewDuration == 0 {
		ts.renewDuration = DefaultRenewDuration
	}
	if ts.expiration == 0 {
		ts.expiration = DefaultExpiration
	}
	return ts
}

type TunnelConnectorOption func(*TunnelConnector)

func WithPrefix(prefix string) TunnelConnectorOption {
	return func(s *TunnelConnector) {
		s.prefix = prefix
	}
}

func WithRenewDuration(duration time.Duration) TunnelConnectorOption {
	return func(s *TunnelConnector) {
		s.renewDuration = duration
	}
}

func WithExpiration(expiration time.Duration) TunnelConnectorOption {
	return func(s *TunnelConnector) {
		s.expiration = expiration
	}
}

func (s *TunnelConnector) key(id string) string {
	return fmt.Sprintf("%s%s:%s", s.prefix, KEY, id)
}

func (s *TunnelConnector) Discover(ctx context.Context, id string) (string, error) {
	key := s.key(id)
	return s.client.Get(ctx, key).Result()
}

func (s *TunnelConnector) Register(ctx context.Context, id, host string, renew func(id string) bool) (err error) {
	err = s.set(ctx, id, host)
	rr := &renewRunner{
		s:        s,
		id:       id,
		host:     host,
		renew:    renew,
		duration: s.renewDuration,
	}
	s.rg.Run(rr)
	return err
}

func (s *TunnelConnector) set(ctx context.Context, id, host string) error {
	key := s.key(id)
	return s.client.Set(ctx, key, host, s.expiration).Err()
}

type renewRunner struct {
	id    string
	host  string
	renew func(id string) bool

	duration time.Duration
	s        *TunnelConnector
}

func (r *renewRunner) Run() {
	tick := time.NewTicker(r.duration)
	defer tick.Stop()
	for {
		<-tick.C
		if !r.renew(r.id) {
			return
		}
		_ = r.s.set(context.Background(), r.id, r.host)
	}
}

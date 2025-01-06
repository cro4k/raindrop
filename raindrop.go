package raindrop

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cro4k/raindrop/v1/core"
	"golang.org/x/sync/errgroup"
)

type Raindrop struct {
	pub      MessagePublisher
	sub      MessageSubscriber
	resolver MessageResolver
	server   *core.Server
}

type Option interface {
	apply(*Raindrop)
}

type OptionFunc func(*Raindrop)

func (f OptionFunc) apply(r *Raindrop) {
	f(r)
}

type Options struct {
	MessagePublisher  MessagePublisher
	MessageSubscriber MessageSubscriber
	MessageResolver   MessageResolver
	Server            *core.Server
}

func (c *Options) apply(r *Raindrop) {
	r.pub = c.MessagePublisher
	r.sub = c.MessageSubscriber
	r.resolver = c.MessageResolver
	r.server = c.Server
}

func applyOptions(raindrop *Raindrop, options ...Option) {
	for _, option := range options {
		option.apply(raindrop)
	}
}

func WithMessagePublisher(pub MessagePublisher) OptionFunc {
	return func(r *Raindrop) {
		r.pub = pub
	}
}

func WithMessageSubscriber(sub MessageSubscriber) OptionFunc {
	return func(r *Raindrop) {
		r.sub = sub
	}
}

func WithMessageResolver(resolver MessageResolver) OptionFunc {
	return func(r *Raindrop) {
		r.resolver = resolver
	}
}

func WithServer(server *core.Server) OptionFunc {
	return func(r *Raindrop) {
		r.server = server
	}
}

func NewRaindrop(options ...Option) *Raindrop {
	raindrop := &Raindrop{}
	applyOptions(raindrop, options...)
	return raindrop
}

func (r *Raindrop) serve(ctx context.Context) error {
	return r.sub.Subscribe(ctx, func(ctx context.Context, m *RawMessage) error {
		destinations, err := r.resolver.Resolve(ctx, m)
		if err != nil {
			return nil
		}
		for _, dst := range destinations {
			err = errors.Join(err, r.server.WriteTo(ctx, dst, m.Data))
		}
		return err
	})
}

func (r *Raindrop) Send(ctx context.Context, data []byte) error {
	if r.pub == nil {
		return fmt.Errorf("message publisher is not set")
	}
	return r.pub.Publish(ctx, &RawMessage{Data: data, Timestamp: time.Now()})
}

func (r *Raindrop) Start(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)

	if r.sub != nil {
		group.Go(func() error {
			return r.serve(ctx)
		})
	}

	group.Go(func() error {
		return r.server.Start(ctx)
	})

	return group.Wait()
}

func (r *Raindrop) Stop(ctx context.Context) error {
	_ = r.sub.Unsubscribe(ctx)
	return r.server.Stop(ctx)
}

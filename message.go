package raindrop

import (
	"context"
	"sync"
	"time"
)

type (
	RawMessage struct {
		ID        string
		Data      []byte
		Timestamp time.Time
	}

	MessageResolver interface {
		Resolve(ctx context.Context, msg *RawMessage) (destinations []string, err error)
	}

	MessageResolveFunc func(ctx context.Context, msg *RawMessage) (destinations []string, err error)

	MessagePublisher interface {
		Publish(ctx context.Context, m *RawMessage) error
	}

	MessageSubscriber interface {
		Subscribe(ctx context.Context, f func(ctx context.Context, m *RawMessage) error) error
		Unsubscribe(ctx context.Context) error
	}

	InMemoryMessageQueue struct {
		mu     sync.RWMutex
		queues map[chan *RawMessage]struct{}
		cancel context.CancelFunc
	}
)

func (f MessageResolveFunc) Resolve(ctx context.Context, msg *RawMessage) (destinations []string, err error) {
	return f(ctx, msg)
}

// NewInMemoryMessageQueue
// Deprecated: you'd better make your own implementation.
// In actually, an in-memory MQ is nonsensical.
func NewInMemoryMessageQueue() *InMemoryMessageQueue {
	return &InMemoryMessageQueue{queues: make(map[chan *RawMessage]struct{})}
}

func (mq *InMemoryMessageQueue) Publish(ctx context.Context, m *RawMessage) error {
	mq.mu.RLock()
	defer mq.mu.RUnlock()
	for q := range mq.queues {
		select {
		case q <- m:
		case <-time.After(time.Second):
			continue // WARNING message will lose
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (mq *InMemoryMessageQueue) Subscribe(ctx context.Context, f func(ctx context.Context, m *RawMessage) error) error {
	ctx, mq.cancel = context.WithCancel(ctx)
	ch := make(chan *RawMessage)
	defer close(ch)
	mq.register(ch)
	defer mq.deregister(ch)
	for {
		select {
		case msg, ok := <-ch:
			if ok {
				_ = f(ctx, msg)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (mq *InMemoryMessageQueue) Unsubscribe(ctx context.Context) error {
	if mq.cancel != nil {
		mq.cancel()
	}
	return nil
}

func (mq *InMemoryMessageQueue) register(ch chan *RawMessage) {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	mq.queues[ch] = struct{}{}
}

func (mq *InMemoryMessageQueue) deregister(ch chan *RawMessage) {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	delete(mq.queues, ch)
}

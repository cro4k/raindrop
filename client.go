package raindrop

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type Client struct {
	conn    Conn
	ping    chan struct{}
	isAlive int32
	once    sync.Once

	timeout   time.Duration
	onError   func(context.Context, error)
	onMessage func(context.Context, []byte) error
	onClose   func(context.Context)
	onConnect func(context.Context)
}

type ClientOption func(*Client)

func WithClientTimeout(timeout time.Duration) ClientOption {
	return func(client *Client) {
		client.timeout = timeout
	}
}

func WithClientOnErrorHandler(onError func(context.Context, error)) ClientOption {
	return func(client *Client) {
		client.onError = onError
	}
}

func WithClientOnMessageHandler(onMessage func(context.Context, []byte) error) ClientOption {
	return func(client *Client) {
		client.onMessage = onMessage
	}
}

func WithClientOnCloseHandler(onClose func(context.Context)) ClientOption {
	return func(client *Client) {
		client.onClose = onClose
	}
}

func NewClient(conn Conn, opts ...ClientOption) *Client {
	c := &Client{
		conn:    conn,
		timeout: 15 * time.Second,
		ping:    make(chan struct{}, 1),
		isAlive: 1,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func WithClientOnConnect(onConnect func(context.Context)) ClientOption {
	return func(client *Client) {
		client.onConnect = onConnect
	}
}

func (c *Client) Write(ctx context.Context, message []byte) error {
	if atomic.LoadInt32(&c.isAlive) != 1 {
		return ErrConnectionClosed
	}
	return c.conn.Write(ctx, message)
}

func (c *Client) Close() (err error) {
	return c.doClose(context.Background())
}

func (c *Client) IsAlive() bool {
	return atomic.LoadInt32(&c.isAlive) == 1
}

func (c *Client) Run(ctx context.Context) (err error) {
	defer func() {
		_ = c.doClose(ctx)
	}()

	go func() {
		_ = c.accept(ctx)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case _, ok := <-c.ping:
			if !ok {
				return nil
			}
		case <-time.After(c.timeout):
			return ErrConnectionTimeout
		}
	}
}

func (c *Client) doClose(ctx context.Context) (err error) {
	c.once.Do(func() {
		if c.onClose != nil {
			c.onClose(ctx)
		}
		atomic.StoreInt32(&c.isAlive, 0)
		err = c.conn.Close()
	})
	<-c.ping
	return err
}

func (c *Client) accept(ctx context.Context) error {
	defer close(c.ping)

	if c.onConnect != nil {
		c.onConnect(ctx)
	}

	for {
		message, err := c.conn.Read(ctx)
		if err != nil {
			atomic.StoreInt32(&c.isAlive, 0)
			if c.onError != nil {
				c.onError(ctx, err)
			}
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case c.ping <- struct{}{}:
		}

		if c.onMessage != nil {
			if err = c.onMessage(ctx, message); err != nil {
				return err
			}
		}
	}
}

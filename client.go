package raindrop

import (
	"context"
	"sync"
	"time"
)

type Client struct {
	Conn
	ping         chan struct{}
	handler      func(ctx context.Context, data []byte) error
	errorHandler func(ctx context.Context, err error)
	closeHandler func(ctx context.Context)
	timeout      time.Duration

	once sync.Once
}

type ClientOption func(*Client)

// WithTimeout set the client connection timout.
// If the server has not received a new message after the 'timeout' duration, the connection will be forcefully closed.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		if timeout > 0 {
			c.timeout = timeout
		}
	}
}

// WithHandler set the message handler
func WithHandler(h func(ctx context.Context, data []byte) error) ClientOption {
	return func(c *Client) {
		c.handler = h
	}
}

// WithErrorHandler set the error handler
func WithErrorHandler(h func(context.Context, error)) ClientOption {
	return func(c *Client) {
		c.errorHandler = h
	}
}

// WithCloseHandler set the callback when the connection is closed.
func WithCloseHandler(callback func(context.Context)) ClientOption {
	return func(c *Client) {
		c.closeHandler = callback
	}
}

// NewClient create a new client with a Conn
func NewClient(conn Conn, options ...ClientOption) *Client {
	c := &Client{
		Conn:    conn,
		ping:    make(chan struct{}),
		timeout: time.Minute,
	}
	for _, option := range options {
		option(c)
	}
	return c
}

// Run is start to accept new messages
func (c *Client) Run(ctx context.Context) {
	defer c.Close()

	go c.accept(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-c.ping:
			if !ok {
				return
			}
		case <-time.After(c.timeout):
		}
	}
}

// Close the connection
func (c *Client) Close() (err error) {
	c.once.Do(func() {
		err = c.Conn.Close()
	})
	<-c.ping
	return err
}

func (c *Client) accept(ctx context.Context) {
	defer close(c.ping)
	defer c.onClose(ctx)
	for {
		data, err := c.Read(context.Background())
		if err != nil {
			c.onError(ctx, err)
			return
		}
		c.onMessage(ctx, data)
		c.ping <- struct{}{}
	}
}

// IsAlive check the connection's health roughly.
// The connection will be treated as alive if it does not time out and is not closed.
func (c *Client) IsAlive() bool {
	select {
	case _, ok := <-c.ping:
		// no matter the connection is time out or closed, the ping channel will be closed
		return ok
	default:
		// otherwise the connection is alive
		return true
	}
}

func (c *Client) onMessage(ctx context.Context, data []byte) {
	if c.errorHandler != nil {
		c.onError(ctx, c.handler(ctx, data))
	}
}

func (c *Client) onError(ctx context.Context, err error) {
	if err != nil && c.errorHandler != nil {
		c.errorHandler(ctx, err)
	}
}
func (c *Client) onClose(ctx context.Context) {
	if c.closeHandler != nil {
		c.closeHandler(ctx)
	}
}

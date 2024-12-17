package websocket

import (
	"context"

	"github.com/coder/websocket"
)

type conn struct {
	id string
	md map[string][]string
	*websocket.Conn

	done chan struct{}
}

func (c *conn) ID() string {
	return c.id
}
func (c *conn) Read(ctx context.Context) ([]byte, error) {
	_, data, err := c.Conn.Read(ctx)
	return data, err
}

func (c *conn) Close() error {
	close(c.done)
	return c.Conn.Close(websocket.StatusInternalError, "")
}

func (c *conn) Write(ctx context.Context, data []byte) error {
	return c.Conn.Write(ctx, websocket.MessageBinary, data)
}

func (c *conn) Metadata() map[string][]string {
	return c.md
}

func (c *conn) Done() <-chan struct{} {
	return c.done
}

func newConn(id string, c *websocket.Conn) *conn {
	return &conn{
		id:   id,
		md:   make(map[string][]string),
		Conn: c,
		done: make(chan struct{}),
	}
}

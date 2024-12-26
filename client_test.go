package raindrop

import (
	"context"
	"io"
	"testing"
	"time"
)

type mockConn struct {
	id string
	ch chan []byte
}

func (c *mockConn) ID() string {
	return c.id
}

func (c *mockConn) Read(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case b, ok := <-c.ch:
		if !ok {
			return nil, io.EOF
		}
		return b, nil
	}
}

func (c *mockConn) Close() error {
	close(c.ch)
	return nil
}

func (c *mockConn) Metadata() map[string][]string {
	return make(map[string][]string)
}

func (c *mockConn) Write(ctx context.Context, data []byte) error {
	return nil
}

func (c *mockConn) send(data []byte) {
	c.ch <- data
}

func newMockConn(id string, bufSize int) *mockConn {
	return &mockConn{
		id: id,
		ch: make(chan []byte, bufSize),
	}
}

func TestClientWithHandler(t *testing.T) {
	id := "1"
	conn := newMockConn(id, 0)

	message := "hello"

	expected := make(chan []byte, 1)
	c := NewClient(conn, WithClientOnMessageHandler(func(ctx context.Context, data []byte) error {
		expected <- data
		return nil
	}))
	defer c.Close()

	go c.Run(context.Background())

	conn.send([]byte(message))

	select {
	case data := <-expected:
		t.Log(string(data))
		if string(data) != message {
			t.Fail()
		}
	case <-time.After(time.Second * 5):
		t.Fail()
	}
}

func TestClientWithTimeout(t *testing.T) {
	id := "1"
	conn := newMockConn(id, 0)
	c := NewClient(conn, WithClientTimeout(3*time.Second))
	t.Log(c.Run(context.Background()))
	if c.IsAlive() {
		t.Fail()
	}
}

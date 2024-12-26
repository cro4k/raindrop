package websocket

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/cro4k/raindrop"
	"github.com/google/uuid"
)

const (
	DefaultMaxWaiting = 5 * time.Second
	XClientIDHeader   = "X-Client-Id"
)

type Listener struct {
	ch chan *conn

	auth func(r *http.Request) (string, error)

	maxWaiting time.Duration
}

var _ raindrop.Listener = (*Listener)(nil)

type ListenerOption func(*Listener)

func WithAuthFunc(auth func(r *http.Request) (string, error)) ListenerOption {
	return func(l *Listener) {
		l.auth = auth
	}
}

func WithMaxWaiting(maxWaiting time.Duration) ListenerOption {
	return func(l *Listener) {
		l.maxWaiting = maxWaiting
	}
}

func DefaultAuth(r *http.Request) (string, error) {
	if id := r.Header.Get(XClientIDHeader); id != "" {
		return id, nil
	}
	return uuid.New().String(), nil
}

func NewListener(options ...ListenerOption) *Listener {
	l := &Listener{
		ch: make(chan *conn),
	}
	for _, option := range options {
		option(l)
	}
	if l.auth == nil {
		l.auth = DefaultAuth
	}
	if l.maxWaiting == 0 {
		l.maxWaiting = DefaultMaxWaiting
	}
	return l
}

func (l *Listener) Accept(ctx context.Context) (raindrop.Conn, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case c, ok := <-l.ch:
		if !ok {
			return nil, io.EOF
		}
		return c, nil
	}
}

// ServeHTTP
// WARNING: goroutine leap
// The current goroutine will never exit if the conn.Close() is not invoked.
func (l *Listener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, err := l.auth(r)
	if err != nil {
		http.Error(w, "auth error", http.StatusUnauthorized)
		return
	}
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		http.Error(w, "accept error", http.StatusBadRequest)
		return
	}
	cc := newConn(id, c)

	select {
	case l.ch <- cc:
	case <-time.After(l.maxWaiting):
		_ = cc.Close()
		http.Error(w, "too many connections", http.StatusRequestTimeout)
		return
	}

	<-cc.Done() // goroutine leap waring
}

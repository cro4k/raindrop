package protocol

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/cro4k/raindrop/core"
	"github.com/google/uuid"
)

const (
	DefaultMaxWaiting = 5 * time.Second
	XClientIDHeader   = "X-Client-Id"
	WebsocketListenOn = ":8010"
)

type websocketConn struct {
	id   string
	conn *websocket.Conn
	done chan struct{}
}

func (c *websocketConn) Read(ctx context.Context) ([]byte, error) {
	_, data, err := c.conn.Read(ctx)
	return data, err
}

func (c *websocketConn) Write(ctx context.Context, data []byte) error {
	return c.conn.Write(ctx, websocket.MessageBinary, data)
}

func (c *websocketConn) Close() error {
	close(c.done)
	return c.conn.Close(websocket.StatusInternalError, "")
}

func (c *websocketConn) Done() <-chan struct{} {
	return c.done
}

func newWebsocketConn(id string, conn *websocket.Conn) *websocketConn {
	return &websocketConn{id: id, conn: conn, done: make(chan struct{})}
}

type WebsocketListener struct {
	auth func(r *http.Request) (string, error)

	srv *http.Server

	maxWaiting time.Duration

	handler func(id string, conn core.Conn) error
}

func (l *WebsocketListener) Close() error {
	return l.srv.Shutdown(context.Background())
}

func (l *WebsocketListener) Serve(ctx context.Context, f func(id string, conn core.Conn) error) error {
	l.handler = f
	l.srv.Handler = l
	slog.InfoContext(ctx, "websocket server is listening on "+l.srv.Addr)
	return l.srv.ListenAndServe()
}

// ServeHTTP
// WARNING: goroutine leap
// The current goroutine will never exit if the conn.Close() is not invoked.
func (l *WebsocketListener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	cc := newWebsocketConn(id, c)

	if err := l.handler(id, cc); err != nil {
		// TODO
		return
	}
	<-cc.Done() // goroutine leap waring
}

type WebsocketListenerOption func(*WebsocketListener)

func WithAuthFunc(auth func(r *http.Request) (string, error)) WebsocketListenerOption {
	return func(l *WebsocketListener) {
		l.auth = auth
	}
}

func WithMaxWaiting(maxWaiting time.Duration) WebsocketListenerOption {
	return func(l *WebsocketListener) {
		l.maxWaiting = maxWaiting
	}
}

func WithListenOn(listenOn string) WebsocketListenerOption {
	return func(listener *WebsocketListener) {
		listener.srv.Addr = listenOn
	}
}

func DefaultAuth(r *http.Request) (string, error) {
	if id := r.Header.Get(XClientIDHeader); id != "" {
		return id, nil
	}
	return uuid.New().String(), nil
}

func NewWebsocketListener(options ...WebsocketListenerOption) *WebsocketListener {
	l := &WebsocketListener{
		srv: &http.Server{Addr: WebsocketListenOn},
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

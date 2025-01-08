package protocol

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/coder/websocket"
	"github.com/cro4k/raindrop/core"
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
	auth    func(r *http.Request) (string, error)
	handler func(id string, conn core.Conn) error
}

func NewWebsocketListener(auth func(r *http.Request) (string, error)) *WebsocketListener {
	return &WebsocketListener{auth: auth, handler: func(id string, conn core.Conn) error {
		return conn.Close()
	}}
}

func (wl *WebsocketListener) Serve(ctx context.Context, h func(id string, conn core.Conn) error) error {
	wl.handler = h
	return nil
}

func (wl *WebsocketListener) Close() error {
	return nil
}

func (wl *WebsocketListener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, err := wl.auth(r)
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
	if err := wl.handler(id, cc); err != nil {
		// TODO
		return
	}
	<-cc.Done()
}

type WebsocketServer struct {
	auth func(r *http.Request) (string, error)

	srv *http.Server
}

func (ws *WebsocketServer) Serve(ctx context.Context, h func(id string, conn core.Conn) error) error {
	ws.srv.Handler = &WebsocketListener{
		auth:    ws.auth,
		handler: h,
	}
	slog.InfoContext(ctx, "websocket server is listening on"+ws.srv.Addr)
	return ws.srv.ListenAndServe()
}

func (ws *WebsocketServer) Close() error {
	return ws.srv.Shutdown(context.Background())
}

func NewWebsocketServer(addr string, auth func(r *http.Request) (string, error)) *WebsocketServer {
	return &WebsocketServer{
		auth: auth,
		srv: &http.Server{
			Addr: addr,
		},
	}
}

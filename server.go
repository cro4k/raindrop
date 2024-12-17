package raindrop

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/cro4k/raindrop/tunnel"
)

type Server struct {
	host     string
	listener Listener
	tunnel   TunnelConnector
	factory  TunnelWriterFactory

	clients sync.Map

	counter        int32
	limitation     int32
	messageHandler func(context.Context, string, []byte) error
	errorHandler   func(context.Context, string, error)

	runners RunnerGroup

	ctx    context.Context
	cancel context.CancelFunc
}

type ServerOption func(s *Server)

// WithTunnelConnector set the TunnelConnector to enable distributed mode.
// If the TunnelConnector is set, the server host is required.
func WithTunnelConnector(tunnel TunnelConnector) ServerOption {
	return func(s *Server) {
		s.tunnel = tunnel
	}
}

// WithServerHost set the server host.
// If the TunnelConnector is set, the server host is required.
func WithServerHost(host string) ServerOption {
	return func(s *Server) {
		s.host = host
	}
}

// WithServerLimitation set the client connection count limitation
func WithServerLimitation(limitation int32) ServerOption {
	return func(s *Server) {
		s.limitation = limitation
	}
}

func WithTunnelWriterFactory(factory TunnelWriterFactory) ServerOption {
	return func(s *Server) {
		s.factory = factory
	}
}

func WithMessageHandler(handler func(context.Context, string, []byte) error) ServerOption {
	return func(s *Server) {
		s.messageHandler = handler
	}
}

func WithServerErrorHandler(handler func(context.Context, string, error)) ServerOption {
	return func(s *Server) {
		s.errorHandler = handler
	}
}

// NewServer create a new server
func NewServer(listener Listener, options ...ServerOption) *Server {
	s := &Server{
		listener: listener,
	}
	for _, option := range options {
		option(s)
	}
	if s.tunnel != nil && s.factory == nil {
		s.factory = NewTunnelWriterFactory(false)
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	return s
}

func (s *Server) Send(ctx context.Context, id string, data []byte) error {
	client, ok := s.clients.Load(id)
	if ok {
		return client.(*Client).Write(ctx, data)
	}
	if s.tunnel == nil || s.factory == nil {
		return ErrConnectionNotFound
	}

	remote, err := s.tunnel.Discover(ctx, id)
	if err != nil {
		return err
	}
	w, err := s.factory.Build(ctx, id, remote)
	if err != nil {
		return err
	}
	return w.Write(ctx, data)
}

func (s *Server) Serve(ctx context.Context) error {
	for {
		conn, err := s.listener.Accept(ctx)
		if err != nil {
			return err
		}
		s.handleConn(conn)
	}
}

func (s *Server) Close() (err error) {
	s.cancel()
	s.runners.Wait()
	return err
}

func (s *Server) handleConn(conn Conn) {
	id := conn.ID()
	if old, ok := s.clients.LoadAndDelete(id); ok {
		atomic.AddInt32(&s.counter, -1)
		// shutdown old connection
		// TODO notify client before close
		_ = old.(*Client).Close()
	}

	// create new connection
	client := NewClient(
		conn,
		WithHandler(s.onClientMessage(id)),
		WithErrorHandler(s.onClientError(id)),
		WithCloseHandler(s.onClientClose(id)),
	)

	if s.tunnel != nil {
		if err := s.tunnel.Register(context.Background(), id, s.host, s.onClientRenew()); err != nil {
			client.Close()
			// TODO
			return
		}
	}

	if s.limitation > 0 && atomic.LoadInt32(&s.counter) >= s.limitation {
		// TODO notify client
		client.Close()
		return
	}
	atomic.AddInt32(&s.counter, 1)
	s.clients.Store(id, client)
	s.runners.Run(s.ctx, client)
}

func (s *Server) onClientMessage(id string) func(ctx context.Context, data []byte) error {
	return func(ctx context.Context, data []byte) error {
		if s.messageHandler != nil {
			return s.messageHandler(ctx, id, data)
		}
		return nil
	}
}

func (s *Server) onClientError(id string) func(context.Context, error) {
	return func(ctx context.Context, err error) {
		if s.errorHandler != nil {
			s.errorHandler(ctx, id, err)
		}
	}
}

func (s *Server) onClientClose(id string) func(context.Context) {
	return func(ctx context.Context) {
		s.clients.Delete(id)
		atomic.AddInt32(&s.counter, -1)
	}
}

func (s *Server) onClientRenew() func(id string) bool {
	return func(id string) bool {
		client, ok := s.clients.Load(id)
		if !ok {
			return false
		}
		return client.(*Client).IsAlive()
	}
}

func (s *Server) RaindropTunnelServer() tunnel.RaindropTunnelServer {
	return &tunnelServer{Server: s}
}

type tunnelServer struct {
	*Server
	tunnel.UnimplementedRaindropTunnelServer
}

func (s *tunnelServer) SendMessage(ctx context.Context, in *tunnel.SendMessageRequest) (*tunnel.SendMessageResponse, error) {
	client, ok := s.clients.Load(in.To)
	if !ok {
		return nil, ErrConnectionNotFound
	}
	err := client.(*Client).Write(ctx, in.Data)
	if err != nil {
		return nil, err
	}
	return &tunnel.SendMessageResponse{}, nil
}

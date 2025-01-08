package core

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"
)

const (
	DefaultClientConnTimeout = 15 * time.Second
)

type Listener interface {
	Serve(ctx context.Context, f func(id string, conn Conn) error) error
	Close() error
}

type Writer interface {
	WriteTo(ctx context.Context, to string, data []byte) error
}

type Server struct {
	listener Listener

	clients sync.Map

	innerCtx context.Context
	cancel   context.CancelFunc

	*options
}

type options struct {
	onClientConnected    func(ctx context.Context, id string, cb Writer)
	onClientMessage      func(ctx context.Context, id string, data []byte, cb Writer)
	onClientDisconnected func(ctx context.Context, id string, cb Writer)
	clientTimeout        time.Duration

	serverIdentity  any
	registryService RegistryService
}

type Option func(*options)

func WithOnClientConnected(onClientConnected func(ctx context.Context, id string, cb Writer)) Option {
	return func(o *options) {
		o.onClientConnected = onClientConnected
	}
}

func WithOnClientMessage(onClientMessage func(ctx context.Context, id string, data []byte, cb Writer)) Option {
	return func(o *options) {
		o.onClientMessage = onClientMessage
	}
}

func WithOnClientDisconnect(onClientDisconnected func(ctx context.Context, id string, cb Writer)) Option {
	return func(o *options) {
		o.onClientDisconnected = onClientDisconnected
	}
}

func WithClientTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.clientTimeout = timeout
	}
}

func WithRegistryService(identity any, service RegistryService) Option {
	return func(o *options) {
		o.serverIdentity = identity
		o.registryService = service
	}
}

func applyOptions(opts ...Option) *options {
	opt := &options{clientTimeout: DefaultClientConnTimeout}
	for _, o := range opts {
		o(opt)
	}
	return opt
}

func NewServer(listener Listener, opts ...Option) *Server {
	return &Server{
		listener: listener,
		options:  applyOptions(opts...),
	}
}

func (s *Server) WriteTo(ctx context.Context, to string, data []byte) error {
	cc, ok := s.clients.Load(to)
	if ok {
		return cc.(*clientConn).Write(ctx, data)
	}

	if s.registryService == nil {
		return ErrClientConnectionNotFound
	}

	w, err := s.registryService.Discover(ctx, to)
	if err != nil {
		return err
	}
	return w.WriteTo(ctx, to, data)
}

func (s *Server) Start(ctx context.Context) error {
	s.innerCtx, s.cancel = context.WithCancel(ctx)
	return s.listener.Serve(ctx, func(id string, conn Conn) error {
		select {
		case <-s.innerCtx.Done():
			return s.innerCtx.Err()
		default:
		}
		cc := newClientConn(id, conn, s.options, s)
		go s.serve(ctx, id, cc)
		return nil
	})
}

func (s *Server) Stop(ctx context.Context) error {
	err := s.listener.Close()
	if s.cancel == nil {
		return errors.Join(err, errors.New("the server has not been started"))
	}
	s.cancel()
	return err
}

func (s *Server) serve(ctx context.Context, id string, cc *clientConn) {
	old, loaded := s.clients.Swap(id, cc)
	if loaded {
		_ = old.(*clientConn).Close()
	}
	defer s.clients.Delete(id)
	if s.registryService != nil {
		defer s.registryService.Deregister(ctx, id)
		if err := s.registryService.Register(ctx, id, s.serverIdentity, cc); err != nil {
			slog.ErrorContext(ctx, "register client conn failed", slog.String("id", id),
				slog.String("error", err.Error()))
			return
		}
	}
	cc.Run(ctx)
}

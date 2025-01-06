package cluster

import (
	"context"
	"errors"
	"log/slog"
	"net"

	"github.com/cro4k/raindrop/v1/core"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	StatusCodeClientConnectionNotFound codes.Code = 10001
)

type GRPCServer struct {
	cc      core.Writer
	version string

	listenOn string
	l        net.Listener
	options  []grpc.ServerOption

	UnimplementedServiceServer
}

func NewGRPCServer(cc core.Writer, listenOn, version string, options ...grpc.ServerOption) *GRPCServer {
	return &GRPCServer{cc: cc, listenOn: listenOn, version: version, options: options}
}

func (s *GRPCServer) Start(ctx context.Context) (err error) {
	s.l, err = net.Listen("tcp", s.listenOn)
	if err != nil {
		return err
	}
	srv := grpc.NewServer(s.options...)
	RegisterServiceServer(srv, s)
	slog.InfoContext(ctx, "grpc server is listening on "+s.listenOn)
	return srv.Serve(s.l)
}

func (s *GRPCServer) Stop(ctx context.Context) error {
	return s.l.Close()
}

func (s *GRPCServer) SendMessage(ctx context.Context, in *SendMessageRequest) (*SendMessageResponse, error) {
	err := s.cc.WriteTo(ctx, in.To, in.Data)
	if errors.Is(err, core.ErrClientConnectionNotFound) {
		return nil, status.Error(StatusCodeClientConnectionNotFound, err.Error())
	}
	if err != nil {
		return nil, err
	}
	return &SendMessageResponse{}, nil
}

func (s *GRPCServer) GetVersion(ctx context.Context, in *GetVersionRequest) (*GetVersionResponse, error) {
	return &GetVersionResponse{Version: s.version}, nil
}

type Client struct {
	cc *grpc.ClientConn
	c  ServiceClient
}

func (c *Client) WriteTo(ctx context.Context, to string, data []byte) error {
	_, err := c.c.SendMessage(ctx, &SendMessageRequest{To: to, Data: data})
	if err == nil {
		return nil
	}
	if st, ok := status.FromError(err); ok && st.Code() == StatusCodeClientConnectionNotFound {
		return core.ErrClientConnectionNotFound
	}
	return err
}

func (c *Client) Close() error {
	return c.cc.Close()
}

func NewClient(cc *grpc.ClientConn) *Client {
	return &Client{
		cc: cc,
		c:  NewServiceClient(cc),
	}
}

func CreateClient(target string, options ...grpc.DialOption) (*Client, error) {
	cc, err := grpc.NewClient(target, options...)
	if err != nil {
		return nil, err
	}
	return NewClient(cc), nil
}

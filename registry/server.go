package registry

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"

	"github.com/cro4k/raindrop/core"
	"github.com/cro4k/raindrop/registry/connector"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type GRPCRegistryServer struct {
	w       core.Writer
	version string
	connector.UnimplementedConnectorServiceServer
}

func NewGRPCRegistryServer(w core.Writer, version string) *GRPCRegistryServer {
	return &GRPCRegistryServer{w: w, version: version}
}

func (s *GRPCRegistryServer) SendMessage(ctx context.Context, req *connector.SendMessageRequest) (*connector.SendMessageResponse, error) {
	err := s.w.WriteTo(ctx, req.To, req.Data)
	if errors.Is(err, core.ErrClientConnectionNotFound) {
		return nil, status.Error(StatusCodeClientConnectionNotFound, err.Error())
	}
	if err != nil {
		return nil, err
	}
	return &connector.SendMessageResponse{}, nil
}

func (s *GRPCRegistryServer) GetVersion(ctx context.Context, req *connector.GetVersionRequest) (*connector.GetVersionResponse, error) {
	return &connector.GetVersionResponse{Version: s.version}, nil
}

func StartGRPCRegistryServer(listenOn string, s *GRPCRegistryServer, options ...grpc.ServerOption) (io.Closer, error) {
	l, err := net.Listen("tcp", listenOn)
	if err != nil {
		return nil, err
	}
	srv := grpc.NewServer(options...)
	connector.RegisterConnectorServiceServer(srv, s)
	slog.InfoContext(context.Background(), "grpc server is listening on "+listenOn)
	go srv.Serve(l)
	return l, err
}

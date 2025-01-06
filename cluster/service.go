package cluster

import (
	"context"
	"sync"

	"github.com/cro4k/raindrop/core"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

type (
	GRPCService interface {
		Register(ctx context.Context, id, host string, healthy core.Healthy) error
		Deregister(ctx context.Context, id string) error
		Discover(ctx context.Context, id string) (string, error)
	}

	grpcServiceWrapper struct {
		GRPCService

		options []grpc.DialOption

		cache sync.Map
	}
)

func FromGRPCService(s GRPCService, options ...grpc.DialOption) core.ClusterService {
	return &grpcServiceWrapper{GRPCService: s, options: options}
}

func (s *grpcServiceWrapper) Register(ctx context.Context, id string, host any, cc core.Healthy) error {
	return s.GRPCService.Register(ctx, id, host.(string), cc)
}

func (s *grpcServiceWrapper) Deregister(ctx context.Context, id string) error {
	return s.GRPCService.Deregister(ctx, id)
}

func (s *grpcServiceWrapper) Discover(ctx context.Context, id string) (core.Writer, error) {
	host, err := s.GRPCService.Discover(ctx, id)
	if err != nil {
		return nil, err
	}

	val, ok := s.cache.Load(host)
	if ok {
		client := val.(*Client)
		state := client.cc.GetState()
		if state <= connectivity.Ready {
			return client, nil
		} else {
			_ = client.Close()
		}
	}

	c, err := CreateClient(host, s.options...)
	if err != nil {
		return nil, err
	}
	s.cache.Store(host, c)
	return c, nil
}

package registry

import (
	"context"
	"sync"

	"github.com/cro4k/raindrop/core"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

type (
	GRPCRegistryService interface {
		Register(ctx context.Context, id, host string, healthy core.Healthy) error
		Deregister(ctx context.Context, id string) error
		Discover(ctx context.Context, id string) (string, error)
	}

	grpcServiceWrapper struct {
		GRPCRegistryService

		options []grpc.DialOption

		cache sync.Map
	}
)

func FromGRPCRegistryService(s GRPCRegistryService, options ...grpc.DialOption) core.RegistryService {
	return &grpcServiceWrapper{GRPCRegistryService: s, options: options}
}

func (s *grpcServiceWrapper) Register(ctx context.Context, id string, host any, cc core.Healthy) error {
	return s.GRPCRegistryService.Register(ctx, id, host.(string), cc)
}

func (s *grpcServiceWrapper) Deregister(ctx context.Context, id string) error {
	return s.GRPCRegistryService.Deregister(ctx, id)
}

func (s *grpcServiceWrapper) Discover(ctx context.Context, id string) (core.Writer, error) {
	host, err := s.GRPCRegistryService.Discover(ctx, id)
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

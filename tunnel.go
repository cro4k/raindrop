package raindrop

import (
	"context"
	"sync"

	"github.com/cro4k/raindrop/tunnel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type (
	TunnelDiscovery interface {
		Discover(ctx context.Context, id string) (string, error)
	}

	TunnelRegistry interface {
		Register(ctx context.Context, id string, host string, renew func(id string) bool) error
	}

	TunnelConnector interface {
		TunnelDiscovery
		TunnelRegistry
	}

	TunnelWriterFactory interface {
		Build(ctx context.Context, id string, host string) (Writer, error)
	}

	tunnelWriter struct {
		to     string
		client tunnel.RaindropTunnelClient
	}

	_tunnelWriterFactory struct {
		clients  sync.Map
		insecure bool
	}
)

func (w *tunnelWriter) Write(ctx context.Context, data []byte) error {
	_, err := w.client.SendMessage(ctx, &tunnel.SendMessageRequest{
		To:   w.to,
		Data: data,
	})
	return err
}

func (f *_tunnelWriterFactory) Build(ctx context.Context, id string, host string) (Writer, error) {
	client, ok := f.clients.Load(host)
	if ok {
		c := client.(tunnel.RaindropTunnelClient)
		return &tunnelWriter{to: id, client: c}, nil
	}
	var options []grpc.DialOption
	if f.insecure {
		options = append(options, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	cc, err := grpc.NewClient(host, options...)
	if err != nil {
		return nil, err
	}
	c := tunnel.NewRaindropTunnelClient(cc)
	f.clients.Store(host, c)
	return &tunnelWriter{to: id, client: c}, nil
}

func NewTunnelWriterFactory(insecure bool) TunnelWriterFactory {
	return &_tunnelWriterFactory{insecure: insecure}
}

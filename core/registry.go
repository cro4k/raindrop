package core

import "context"

type RegistryService interface {
	Register(ctx context.Context, id string, serverIdentity any, healthy Healthy) error
	Deregister(ctx context.Context, id string) error
	Discover(ctx context.Context, id string) (Writer, error)
}

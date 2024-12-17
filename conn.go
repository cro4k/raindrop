package raindrop

import "context"

type (
	Writer interface {
		Write(ctx context.Context, data []byte) error
	}

	Conn interface {
		ID() string
		Read(ctx context.Context) ([]byte, error)
		Close() error
		Metadata() map[string][]string
		Writer
	}

	Listener interface {
		// Accept new connection
		Accept(ctx context.Context) (conn Conn, err error)
	}
)

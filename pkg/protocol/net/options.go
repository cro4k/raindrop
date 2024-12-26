package net

const (
	DefaultMaxMessageSize    = 1 << 22
	DefaultMessageBufferSize = 1 << 12
)

type options struct {
	maxMessageSize    int
	messageBufferSize int
	auth              func(data []byte) (string, error)
	connBufferSize    int
}

type Option func(o *options)

func WithMaxMessageSize(size int) Option {
	return func(o *options) {
		o.maxMessageSize = size
	}
}

func WithMessageBufferSize(size int) Option {
	return func(o *options) {
		o.messageBufferSize = size
	}
}

func WithAuth(auth func(data []byte) (string, error)) Option {
	return func(o *options) {
		o.auth = auth
	}
}

func WithConnBufferSize(size int) Option {
	return func(o *options) {
		o.connBufferSize = size
	}
}

func applyOptions(opts ...Option) *options {
	opt := &options{
		maxMessageSize:    DefaultMaxMessageSize,
		messageBufferSize: DefaultMessageBufferSize,
	}
	for _, o := range opts {
		o(opt)
	}
	return opt
}

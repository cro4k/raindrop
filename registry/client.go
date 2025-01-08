package registry

import (
	"context"

	"github.com/cro4k/raindrop/core"
	"github.com/cro4k/raindrop/registry/connector"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	StatusCodeClientConnectionNotFound codes.Code = 10001
)

type Client struct {
	cc *grpc.ClientConn
	c  connector.ConnectorServiceClient
}

func (c *Client) WriteTo(ctx context.Context, to string, data []byte) error {
	_, err := c.c.SendMessage(ctx, &connector.SendMessageRequest{To: to, Data: data})
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
		c:  connector.NewConnectorServiceClient(cc),
	}
}

func CreateClient(target string, options ...grpc.DialOption) (*Client, error) {
	cc, err := grpc.NewClient(target, options...)
	if err != nil {
		return nil, err
	}
	return NewClient(cc), nil
}

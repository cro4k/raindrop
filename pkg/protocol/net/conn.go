package net

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
)

type conn struct {
	conn net.Conn
	id   string
	md   map[string][]string

	*options
}

func (c *conn) ID() string {
	return c.id
}

func (c *conn) Read(ctx context.Context) ([]byte, error) {
	head := make([]byte, 4)
	n, err := c.conn.Read(head)
	if err != nil {
		return nil, err
	}
	if n != 4 {
		return nil, errors.New("read header size error")
	}
	size := binary.BigEndian.Uint32(head)
	if int(size) > c.maxMessageSize {
		return nil, errors.New("read header size error")
	}
	message := make([]byte, size)
	_, err = io.ReadFull(c.conn, message)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func (c *conn) Close() error {
	return c.conn.Close()
}

func (c *conn) Metadata() map[string][]string {
	return c.md
}

func (c *conn) Write(ctx context.Context, data []byte) error {
	if len(data) > c.maxMessageSize {
		return errors.New("data too large")
	}
	head := make([]byte, 4)
	binary.BigEndian.PutUint32(head, uint32(len(data)))
	_, err := c.conn.Write(append(head, data...))
	return err
}

func newConn(id string, c net.Conn, opt *options) *conn {
	return &conn{
		conn:    c,
		id:      id,
		md:      make(map[string][]string),
		options: opt,
	}
}

package net

import (
	"context"
	"errors"
	"io"
	"net"

	"github.com/cro4k/raindrop"
)

var errListenerBroken = errors.New("listener broken")

type Listener struct {
	l   net.Listener
	opt *options

	ch chan raindrop.Conn
}

func (l *Listener) Accept(ctx context.Context) (raindrop.Conn, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case c, ok := <-l.ch:
		if !ok {
			return nil, io.EOF
		}
		return c, nil
	}
}

func (l *Listener) Serve(ctx context.Context) error {
	defer close(l.ch)
	for {
		cc, err := l.accept(ctx)
		if errors.Is(err, errListenerBroken) {
			return err
		}
		if err != nil {
			continue
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case l.ch <- cc:
		}
	}
}

func (l *Listener) Close() error {
	return l.l.Close()
}

func (l *Listener) accept(ctx context.Context) (rc raindrop.Conn, err error) {
	c, err := l.l.Accept()
	if err != nil {
		return nil, errors.Join(errListenerBroken, err)
	}
	defer func() {
		if err != nil {
			c.Close()
		}
	}()
	cc := newConn("", c, l.opt)
	firstMessage, err := cc.Read(ctx)
	if err != nil {
		return nil, err
	}
	var id string
	if l.opt.auth != nil {
		id, err = l.opt.auth(firstMessage)
		if err != nil {
			return nil, err
		}
	} else {
		id = c.RemoteAddr().String()
	}
	cc.id = id
	return cc, nil
}

func NewListener(network, addr string, opts ...Option) (*Listener, error) {
	l, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}
	opt := applyOptions(opts...)
	return &Listener{
		l:   l,
		opt: opt,
		ch:  make(chan raindrop.Conn, opt.connBufferSize),
	}, nil
}

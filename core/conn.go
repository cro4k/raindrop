package core

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type (
	Conn interface {
		Read(ctx context.Context) ([]byte, error)
		Write(ctx context.Context, data []byte) error
		Close() error
	}

	Healthy interface {
		IsAlive() bool
	}
)

type clientConn struct {
	Conn

	timeout time.Duration

	once sync.Once

	ping chan struct{}

	onClientConnected    func(ctx context.Context)
	onClientMessage      func(ctx context.Context, data []byte)
	onClientDisconnected func(ctx context.Context)

	isAlive int32
}

func (cc *clientConn) onceClose(ctx context.Context) (err error) {
	cc.once.Do(func() {
		cc.setAlive(false)
		if cc.onClientDisconnected != nil {
			cc.onClientDisconnected(ctx)
		}
		err = cc.Conn.Close()
	})
	<-cc.ping
	return err
}

func (cc *clientConn) IsAlive() bool {
	return atomic.LoadInt32(&cc.isAlive) == 1
}

func (cc *clientConn) setAlive(alive bool) {
	if alive {
		atomic.StoreInt32(&cc.isAlive, 1)
	} else {
		atomic.StoreInt32(&cc.isAlive, 0)
	}
}

func (cc *clientConn) Close() (err error) {
	return cc.onceClose(context.Background())
}

func (cc *clientConn) receive(ctx context.Context) {
	defer cc.onceClose(ctx)
	defer close(cc.ping)

	cc.onClientConnected(ctx)
	cc.setAlive(true)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		data, err := cc.Read(ctx)
		if err != nil {
			return
		}
		cc.ping <- struct{}{}
		if cc.onClientMessage != nil {
			cc.onClientMessage(ctx, data)
		}
	}
}

func (cc *clientConn) Run(ctx context.Context) {
	defer cc.onceClose(ctx)

	go cc.receive(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-cc.ping:
			if !ok {
				return
			}
		case <-time.After(cc.timeout):
			return
		}
	}
}

func newClientConn(id string, conn Conn, opt *options, cb Writer) *clientConn {
	return &clientConn{
		Conn:    conn,
		timeout: opt.clientTimeout,
		ping:    make(chan struct{}),
		onClientConnected: func(ctx context.Context) {
			if opt.onClientConnected != nil {
				opt.onClientConnected(ctx, id, cb)
			}
		},
		onClientMessage: func(ctx context.Context, data []byte) {
			if opt.onClientMessage != nil {
				opt.onClientMessage(ctx, id, data, cb)
			}
		},
		onClientDisconnected: func(ctx context.Context) {
			if opt.onClientDisconnected != nil {
				opt.onClientDisconnected(ctx, id, cb)
			}
		},
	}
}

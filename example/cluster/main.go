package main

import (
	"context"
	"encoding/json"
	"flag"
	"log/slog"
	"os"
	"os/signal"

	"github.com/cro4k/raindrop"
	"github.com/cro4k/raindrop/cluster"
	"github.com/cro4k/raindrop/core"
	"github.com/cro4k/raindrop/example/messages"
	"github.com/cro4k/raindrop/protocol"
	"github.com/redis/go-redis/v9"
)

var (
	httpListenOn string
	grpcListenOn string
	redisServer  string
)

func init() {
	flag.StringVar(&httpListenOn, "http", "127.0.0.1:8001", "Listen on host:port")
	flag.StringVar(&grpcListenOn, "grpc", "127.0.0.1:9001", "Listen on host:port")
	flag.StringVar(&redisServer, "redis-server", "127.0.0.1:6379", "Redis server address")
	flag.Parse()
}

// This example is shown send message from clint-to-server side by websocket connection, so the MQ is skipped.
// In suggested way, you can start a http server and send message by http api, and use Raindrop.Send() to
// distribute message, then the message is distributed by MQ. And the websocket connection should be only used for
// server-to-client side communicate
func main() {
	// the multi-services registry and discovery
	clusterService := cluster.NewRedisClusterService(
		redis.NewUniversalClient(&redis.UniversalOptions{Addrs: []string{redisServer}}),
	)

	// hold the client connections, and write the message to target client.
	srv := core.NewServer(
		protocol.NewWebsocketListener(protocol.WithListenOn(httpListenOn)),
		core.WithOnClientMessage(onClientMessage),
		core.WithClusterService(grpcListenOn, cluster.FromGRPCService(clusterService)),
	)

	// the connector for multi-servers in the cluster
	grpcServer := cluster.NewGRPCServer(srv, grpcListenOn, "0.0.1")

	// distribute the messages over clients
	r := raindrop.NewRaindrop(&raindrop.Options{
		MessagePublisher:  nil,
		MessageSubscriber: nil,
		MessageResolver:   raindrop.MessageResolveFunc(messageResolver),
		Server:            srv,
	})

	ctx := context.Background()

	// MQ is used when send message by this function, and we suggested send message to client in this way.
	// r.Send(ctx, data)

	go grpcServer.Start(ctx)
	go r.Start(ctx)
	defer grpcServer.Stop(ctx)
	defer r.Stop(ctx)

	s := make(chan os.Signal)
	signal.Notify(s, os.Interrupt)
	<-s
}

func onClientMessage(ctx context.Context, id string, data []byte, cb core.Writer) {
	m := new(messages.Message)
	if err := json.Unmarshal(data, m); err != nil {
		slog.ErrorContext(ctx, "cannot unmarshal message", "error", err)
		return
	}
	slog.InfoContext(ctx, "IN",
		slog.String("FROM", m.From),
		slog.String("TO", m.To),
		slog.String("CONTENT", m.Content),
	)
	if m.To == "" {
		return
	}
	_ = cb.WriteTo(ctx, m.To, data)
}

func messageResolver(ctx context.Context, msg *raindrop.RawMessage) (destinations []string, err error) {
	m := new(messages.Message)
	if err := json.Unmarshal(msg.Data, m); err != nil {
		return nil, err
	}
	return []string{m.To}, nil
}

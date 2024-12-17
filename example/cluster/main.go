package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/cro4k/raindrop"
	"github.com/cro4k/raindrop/example/messages"
	"github.com/cro4k/raindrop/pkg/protocol/websocket"
	"github.com/cro4k/raindrop/pkg/redistunnel"
	"github.com/cro4k/raindrop/tunnel"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

var httpListenOn string
var grpcListenOn string

func init() {
	flag.StringVar(&httpListenOn, "http", ":8080", "Listen on host:port")
	flag.StringVar(&grpcListenOn, "grpc", ":9001", "grpc listen on")
	flag.Parse()
}

func main() {
	// The websocket listener, or you can replace it with any implemented listener.
	listener := websocket.NewListener()

	client := redis.NewUniversalClient(&redis.UniversalOptions{Addrs: []string{"localhost:6379"}})
	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Fatal(err)
	}

	var s *raindrop.Server
	s = raindrop.NewServer(
		listener,
		raindrop.WithServerHost("localhost"+grpcListenOn),                    // set server host
		raindrop.WithTunnelConnector(redistunnel.NewTunnelConnector(client)), // set tunnel connector
		raindrop.WithTunnelWriterFactory(raindrop.NewTunnelWriterFactory(true)),
		raindrop.WithMessageHandler(func(ctx context.Context, id string, bytes []byte) error {
			fmt.Println(id, string(bytes))
			m := messages.Message{}
			if err := json.Unmarshal(bytes, &m); err != nil {
				return err
			}
			m.From = id
			return s.Send(context.Background(), m.To, m.JSON())
		}),
	)

	srv := grpc.NewServer()
	tunnel.RegisterRaindropTunnelServer(srv, s.RaindropTunnelServer())

	tcpListener, err := net.Listen("tcp", grpcListenOn)
	if err != nil {
		log.Fatal(err)
	}

	// start grpc server, it's used to receive message which is forwarded by another server.
	go srv.Serve(tcpListener)
	// start raindrop
	go s.Serve(context.Background())
	// start http server to accept websocket connection
	go http.ListenAndServe(httpListenOn, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(websocket.XClientIDHeader)
		if id == "" {
			id = uuid.New().String()
			r.Header.Set(websocket.XClientIDHeader, id)
		}
		fmt.Println("id:", id)
		listener.ServeHTTP(w, r)
	}))

	fmt.Println("websocket server listen on ", httpListenOn)
	fmt.Println("grpc server listen on ", grpcListenOn)
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt)
	<-signals
}

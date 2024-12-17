package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/cro4k/raindrop"
	"github.com/cro4k/raindrop/example/messages"
	"github.com/cro4k/raindrop/pkg/protocol/websocket"
	"github.com/google/uuid"
)

var listenOn string

func init() {
	flag.StringVar(&listenOn, "listen", ":8080", "Listen on host:port")
	flag.Parse()
}

func main() {
	// The websocket listener, or you can replace it with any implemented listener.
	listener := websocket.NewListener()

	var s *raindrop.Server
	s = raindrop.NewServer(
		listener,
		raindrop.WithMessageHandler(func(ctx context.Context, id string, bytes []byte) error {
			fmt.Println(id, string(bytes))
			m := messages.Message{}
			if err := json.Unmarshal(bytes, &m); err != nil {
				return err
			}
			m.From = id
			return s.Send(ctx, m.To, m.JSON())
		}),
	)

	// start raindrop
	go s.Serve(context.Background())
	// start http server to accept websocket connection
	go http.ListenAndServe(listenOn, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Rewrite the http header to overwrite client id, this part is just for example.
		// You should write your own authentication in your application.
		id := r.Header.Get(websocket.XClientIDHeader)
		if id == "" {
			id = uuid.New().String()
			r.Header.Set(websocket.XClientIDHeader, id)
		}
		fmt.Println("id:", id)

		listener.ServeHTTP(w, r)
	}))
	fmt.Println("websocket server listen on ", listenOn)

	// wait for exit
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt)
	<-signals
}

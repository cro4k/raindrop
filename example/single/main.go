package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"

	"github.com/cro4k/raindrop/core"
	"github.com/cro4k/raindrop/example/messages"
	"github.com/cro4k/raindrop/protocol"
)

// This example is shown send message from clint-to-server side by websocket connection, so the MQ is skipped.
// In suggested way, you can start a http server and send message by http api, and use Raindrop.Send() to
// distribute message, then the message is distributed by MQ. And the websocket connection should be only used for
// server-to-client side communicate
func main() {

	srv := core.NewServer(
		protocol.NewWebsocketListener(
			protocol.WithListenOn(":8000"), // If not set, the server will listen on :8010 in default.
		),
		core.WithOnClientMessage(onClientMessage),
	)

	// mq := raindrop.NewInMemoryMessageQueue()

	r := raindrop.NewRaindrop(&raindrop.Options{
		// MQ is not used in this example, because the message is sent in callback, the MQ is skipped.
		// MessagePublisher:  mq,
		// MessageSubscriber: mq,
		MessageResolver: raindrop.MessageResolveFunc(messageResolver),
		Server:          srv,
	})
	ctx := context.Background()

	// MQ is used when send message by this function, and we suggested send message to client in this way.
	// r.Send(ctx, data)

	go r.Start(ctx)
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

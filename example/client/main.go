package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/coder/websocket"
	"github.com/cro4k/raindrop/example/messages"
	"github.com/google/uuid"
)

var (
	server   string
	clientID string
)

func init() {
	flag.StringVar(&server, "server", "ws://localhost:8080", "server address")
	flag.StringVar(&clientID, "id", "", "client id")
	flag.Parse()
	if !strings.HasPrefix(server, "ws") {
		server = "ws://" + server
	}
	if clientID == "" {
		clientID = uuid.NewString()
	}
}

func main() {
	fmt.Println(clientID)
	conn, _, err := websocket.Dial(context.Background(), server, &websocket.DialOptions{
		HTTPHeader: http.Header{"X-Client-Id": []string{clientID}},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(0, "")

	go read(conn)

	fmt.Printf("connect %s successfully: %s\n", server, clientID)

	re := bufio.NewReader(os.Stdin)
	for {
		b, err := re.ReadBytes('\n')
		if err != nil {
			log.Fatal(err)
		}
		text := strings.TrimSpace(string(b))
		if text == "" {
			continue
		}
		if text == "exit" {
			return
		}
		n := strings.IndexByte(text, ' ')
		if n <= 0 {
			fmt.Println("[invalid command]")
			continue
		}

		to := text[:n]
		message := strings.TrimSpace(text[n+1:])

		m := messages.Message{
			To:      to,
			From:    clientID,
			Content: message,
		}
		err = conn.Write(context.Background(), websocket.MessageBinary, m.JSON())
		if err != nil {
			log.Println(err)
		}
	}
}

func read(conn *websocket.Conn) {
	for {
		_, data, err := conn.Read(context.Background())
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println(">>:", string(data))
	}
}

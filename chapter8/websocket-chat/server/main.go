package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/net/websocket"

	"chat"
)

func main() {
	dispatch := make(chan chat.Message)
	connectCh := make(chan chan chat.Message)
	disconnectCh := make(chan chan chat.Message)
	go func() {
		clients := make(map[chan chat.Message]struct{})
		for {
			select {
			case c := <-connectCh:
				clients[c] = struct{}{}
			case c := <-disconnectCh:
				delete(clients, c)
			case msg := <-dispatch:
				for c := range clients {
					select {
					case c <- msg:
					default:
						delete(clients, c)
						close(c)
					}
				}
			}
		}
	}()
	// Create a websocket server
	http.Handle("/chat", websocket.Handler(func(conn *websocket.Conn) {
		client := conn.RemoteAddr().String()
		inputCh := make(chan chat.Message, 10)
		connectCh <- inputCh
		defer func() {
			disconnectCh <- inputCh
		}()
		decoder := json.NewDecoder(conn)
		data, decodeErrCh := chat.DecodeToChan(func(msg *chat.Message) error {
			err := decoder.Decode(msg)
			msg.From = client
			return err
		})
		encodeErrCh := chat.EncodeFromChan(inputCh, func(msg chat.Message) ([]byte, error) {
			return json.Marshal(msg)
		}, conn)
		for {
			select {
			case msg, ok := <-data:
				if !ok {
					return
				}
				dispatch <- msg
			case <-decodeErrCh:
				return
			case <-encodeErrCh:
				return
			}
		}
	}))

	fmt.Println("Server started at :10001")
	if err := http.ListenAndServe(":10001", nil); err != nil {
		panic(err)
	}
}

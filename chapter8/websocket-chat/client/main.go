package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/net/websocket"

	"chat"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Run client with server address")
	}

	// Create a client
	cli, err := websocket.Dial("ws://"+os.Args[1]+"/chat", "", "http://"+os.Args[1])
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	decoder := json.NewDecoder(cli)
	rcvCh, rcvErrCh := chat.DecodeToChan(func(msg *chat.Message) error {
		return decoder.Decode(msg)
	})
	sendCh := make(chan chat.Message)
	sendErrCh := chat.EncodeFromChan(sendCh, func(msg chat.Message) ([]byte, error) {
		return json.Marshal(msg)
	}, cli)
	done := make(chan struct{})
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := scanner.Text()
			select {
			case <-done:
				return
			default:
			}
			sendCh <- chat.Message{
				Message: text,
			}
		}
	}()
	for {
		select {
		case msg, ok := <-rcvCh:
			if !ok {
				close(done)
				return
			}
			fmt.Println(msg)
		case <-sendErrCh:
			return
		case <-rcvErrCh:
			return
		}
	}
}

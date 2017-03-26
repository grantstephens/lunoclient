package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

// var connection = websocket.Conn{}
var market = marketStruct{}

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	connection := connect()

	done := make(chan struct{})
	update := make(chan marketUpdate, 100)
	go receiveUpdate(update)

	go func() {
		defer connection.Close()
		defer close(done)
		for {
			upwind := marketUpdate{}
			_, message, err := connection.ReadMessage()
			if err != nil {
				log.Fatal("send", err)
			}
			if string(message) == "\"\"" {
				log.Println("Blank keep alive message")
			} else {
				// log.Print(string(message))
				err := json.Unmarshal(message, &upwind)
				if err != nil {
					log.Fatal("Could not unmarshall")
				}
				update <- upwind
			}
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-interrupt:
			log.Println("interrupt")
			// To cleanly close a connection, a client should send a close
			// frame and wait for the server to close the connection.
			err := connection.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			connection.Close()
			return
		}
	}
}

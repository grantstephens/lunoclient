package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

type tran struct {
	ID     string  `json:"id"`
	Price  float64 `json:"price,string"`
	Volume float64 `json:"volume,string"`
}

type marketStruct struct {
	Sequence int64 `json:"sequence,string"`
	AsksM    map[string]tran
	BidsM    map[string]tran
	Asks     []tran `json:"asks"`
	Bids     []tran `json:"bids"`
}

type tranUpdate struct {
	Base    string `json:"base"`
	Counter string `json:"counter"`
	OrderID string `json:"order_id"`
}

type marketUpdate struct {
	Sequence     int64        `json:"sequence,string"`
	TradeUpdates []tranUpdate `json:"trade_updates"`
	CreateUpdate struct {
		OrderID string `json:"order_id"`
		Type    string `json:"type"`
		Price   string `json:"price"`
		Volume  string `json:"volume"`
	} `json:"create_update"`
	DeleteUpdate struct {
		OrderID string `json:"order_id"`
	} `json:"delete_update"`
	Timestamp int64 `json:"timestamp"`
}

var market = marketStruct{}
var conn = websocket.Conn{}

func receiveUpdate(update chan marketUpdate) {
	for updateMsg := range update {
		if market.Sequence+1 == updateMsg.Sequence {
			log.Println("Msg:", updateMsg, "Queue:", len(update))
			// log.Println(market.Sequence)
			market.Sequence = updateMsg.Sequence //Temp
		} else {
			log.Fatal("Out of Sequence")
		}
		// time.Sleep(4 * 1e9)
		// up := <-update
		// fmt.Println("Update Message: ", up)
	}
}

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	conn = *connect()

	done := make(chan struct{})
	update := make(chan marketUpdate, 100)
	go receiveUpdate(update)

	go func() {
		defer conn.Close()
		defer close(done)
		for {
			upwind := marketUpdate{}
			_, message, err := conn.ReadMessage()
			// err3 := c.ReadJSON(&upwind)
			if err != nil {
				log.Fatal("send", err)
			}
			// fmt.Println(upwind)
			if string(message) == "\"\"" {
				log.Println("Blank keep alive message")
			} else {
				// fmt.Println(string(message))
				err := json.Unmarshal(message, &upwind)
				if err != nil {
					log.Fatal("Could not unmarshall")
				}
				// fmt.Println(upwind)
				update <- upwind
				// log.Println(len(update))
			}
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			err := conn.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")
			// To cleanly close a connection, a client should send a close
			// frame and wait for the server to close the connection.
			err := conn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			conn.Close()
			return
		}
	}
}

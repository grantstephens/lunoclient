package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "ws.bitx.co", "http service address")
var endpoint = "/api/1/stream/XBTZAR"
var auth struct {
	Key    string `json:"api_key_id"`
	Secret string `json:"api_key_secret"`
}

func getAuthStr() []byte {
	authFile, err := os.Open("auth.json")
	defer authFile.Close()
	if err != nil {
		log.Fatal("opening config file", err.Error())
	}
	jsonParser := json.NewDecoder(authFile)
	if err = jsonParser.Decode(&auth); err != nil {
		log.Fatal("parsing config file", err.Error())
	}
	authStr, _ := json.Marshal(auth)
	return authStr
}

func doAuth(conn *websocket.Conn) {
	// if authStr {
	authStr := getAuthStr()
	// }
	err := conn.WriteMessage(websocket.TextMessage, authStr)
	if err != nil {
		log.Fatal("Auth string could not be sent:", err)
	}
	log.Println("Authenticated")
}

func connect() *websocket.Conn {
	u := url.URL{Scheme: "wss", Host: *addr, Path: endpoint}
	log.Printf("connecting to %s", u.String())
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer conn.Close()
	doAuth(conn)
	market.getMarket(conn)
	return conn
}

func (market *marketStruct) getMarket(conn *websocket.Conn) {
	err := conn.ReadJSON(&market)
	if err != nil {
		log.Fatal("Could not read market:", err)
	}
	log.Println("Market Window Received")

	market.AsksM = make(map[string]tran)
	for _, ask := range market.Asks {
		market.AsksM[ask.ID] = ask
	}
	market.BidsM = make(map[string]tran)
	for _, bid := range market.Bids {
		market.BidsM[bid.ID] = bid
	}
}

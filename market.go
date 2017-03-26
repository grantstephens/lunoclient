package main

import (
	"log"
	"math"
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
	Base    float64 `json:"base,string"`
	Counter float64 `json:"counter,string"`
	OrderID string  `json:"order_id"`
}

type createUpdate struct {
	OrderID string  `json:"order_id"`
	Type    string  `json:"type"`
	Price   float64 `json:"price,string"`
	Volume  float64 `json:"volume,string"`
}

type deleteUpdate struct {
	OrderID string `json:"order_id"`
}

type marketUpdate struct {
	Sequence     int64        `json:"sequence,string"`
	TradeUpdates []tranUpdate `json:"trade_updates"`
	CreateUpdate createUpdate `json:"create_update"`
	DeleteUpdate deleteUpdate `json:"delete_update"`
	Timestamp    int64        `json:"timestamp"`
}

// var market = marketStruct{}
var nilCreateUpdate = createUpdate{}
var nilDeleteUpdate = deleteUpdate{}
var order = tran{}

func receiveUpdate(update chan marketUpdate) {
	for updateMsg := range update {
		if market.Sequence+1 == updateMsg.Sequence {
			market.processUpdate(&updateMsg)
		} else {
			log.Fatal("Out of Sequence")
		}
	}
}

func (market *marketStruct) processUpdate(updateMsg *marketUpdate) {
	// log.Println(updateMsg)
	if len(updateMsg.TradeUpdates) != 0 {
		for _, updateTrans := range updateMsg.TradeUpdates {
			market.processTrade(&updateTrans)
		}
	}
	if updateMsg.CreateUpdate != nilCreateUpdate {
		market.processCreate(&updateMsg.CreateUpdate)
	}
	if updateMsg.DeleteUpdate != nilDeleteUpdate {
		_, okAsk := market.AsksM[updateMsg.DeleteUpdate.OrderID]
		_, okBid := market.BidsM[updateMsg.DeleteUpdate.OrderID]
		if okAsk {
			delete(market.AsksM, updateMsg.DeleteUpdate.OrderID)
		} else if okBid {
			delete(market.BidsM, updateMsg.DeleteUpdate.OrderID)
		} else {
			log.Fatal("ID not found while trying to delete.")
		}
	}
	market.Sequence = updateMsg.Sequence
	maxBid := 0.0
	for id := range market.BidsM {
		maxBid = math.Max(market.BidsM[id].Price, maxBid)
	}
	minAsk := 10e12
	for id := range market.AsksM {
		minAsk = math.Min(market.AsksM[id].Price, minAsk)
	}
	log.Println(maxBid, minAsk, minAsk-maxBid)
}

func (market *marketStruct) processTrade(updateTrans *tranUpdate) {
	// log.Println("Update", updateTrans)
	// log.Println("Market Ask", market.AsksM[updateTrans.OrderID])
	// log.Println("Market Bid", market.BidsM[updateTrans.OrderID])
	// log.Println(updateTrans)
	_, okAsk := market.AsksM[updateTrans.OrderID]
	_, okBid := market.BidsM[updateTrans.OrderID]
	if okAsk {
		askUpdate := tran{
			ID:     updateTrans.OrderID,
			Price:  market.AsksM[updateTrans.OrderID].Price,
			Volume: market.AsksM[updateTrans.OrderID].Volume - updateTrans.Base,
		}
		market.AsksM[updateTrans.OrderID] = askUpdate
		if market.AsksM[updateTrans.OrderID].Volume == 0 {
			delete(market.AsksM, updateTrans.OrderID)
		} else if market.AsksM[updateTrans.OrderID].Volume < 0 {
			log.Fatal("Not enough volume")
		}
	} else if okBid {
		bidUpdate := tran{
			ID:     updateTrans.OrderID,
			Price:  market.BidsM[updateTrans.OrderID].Price,
			Volume: market.BidsM[updateTrans.OrderID].Volume - updateTrans.Base,
		}
		market.BidsM[updateTrans.OrderID] = bidUpdate
		if market.BidsM[updateTrans.OrderID].Volume == 0 {
			delete(market.BidsM, updateTrans.OrderID)
		} else if market.BidsM[updateTrans.OrderID].Volume < 0 {
			log.Fatal("Not enough volume")
		}
	} else {
		log.Fatal("ID not found while trying to update.")
	}
}

func (market *marketStruct) processCreate(updateCreate *createUpdate) {
	order = tran{
		ID:     updateCreate.OrderID,
		Price:  updateCreate.Price,
		Volume: updateCreate.Volume,
	}
	if updateCreate.Type == "ASK" {
		market.AsksM[updateCreate.OrderID] = order
	} else {
		market.BidsM[updateCreate.OrderID] = order
	}
}

package binance

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/crypto_pickle/internal/orderbook"
	"github.com/crypto_pickle/internal/utils"
)

type RawOrderBook struct {
	LastUpdateId int64      `json:"lastUpdateId"`
	Bids         [][]string `json:"bids"`
	Asks         [][]string `json:"asks"`
}

func (rawOB *RawOrderBook) ToOrderBook() orderbook.OrderBook {
	newOB := orderbook.OrderBook{
		Time: 0,
		Bids: make(orderbook.DepthLevel),
		Asks: make(orderbook.DepthLevel),
	}

	for _, bid := range rawOB.Bids {
		newOB.Bids[utils.StringToFloat(bid[0])] = utils.StringToFloat(bid[1])
	}

	for _, ask := range rawOB.Asks {
		newOB.Asks[utils.StringToFloat(ask[0])] = utils.StringToFloat(ask[1])
	}

	return newOB
}

func calculateOrderBookWeight(limit int32) int32 {
	if limit > 1000 {
		return 50
	} else if limit > 500 {
		return 10
	} else if limit > 100 {
		return 5
	} else {
		return 1
	}
}

func (client *BinanceClient) GetOrderBook(symbol string, limit int32) *RawOrderBook {
	endpoint := fmt.Sprintf("v3/depth?symbol=%s&limit=%d", symbol, limit)
	bytes := client.makeAPIRequest(endpoint, calculateOrderBookWeight(limit))

	rawOB := new(RawOrderBook)
	err := json.Unmarshal(bytes, &rawOB)
	if err != nil {
		log.Fatal("Error: ", err)
	}

	return rawOB
}

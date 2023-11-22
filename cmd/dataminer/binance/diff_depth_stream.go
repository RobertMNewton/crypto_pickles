package binance

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"

	"github.com/crypto_pickle/internal/orderbook"
	"github.com/crypto_pickle/internal/utils"
)

type RawDepthDiff struct {
	EventType     string     `json:"e"`
	EventTime     int64      `json:"E"`
	Symbol        string     `json:"s"`
	FirstUpdateId int64      `json:"U"`
	LastUpdateId  int64      `json:"u"`
	Bids          [][]string `json:"b"`
	Asks          [][]string `json:"a"`
}

func (rawDiff RawDepthDiff) ToDepthDiff() orderbook.DepthDiff {
	newDiff := orderbook.DepthDiff{
		Time: rawDiff.EventTime,
		Bids: make(orderbook.DepthLevel),
		Asks: make(orderbook.DepthLevel),
	}

	for _, bid := range rawDiff.Bids {
		newDiff.Bids[utils.StringToFloat(bid[0])] = utils.StringToFloat(bid[1])
	}

	for _, ask := range rawDiff.Asks {
		newDiff.Asks[utils.StringToFloat(ask[0])] = utils.StringToFloat(ask[1])
	}

	return newDiff
}

func FromJsonBytes(bytes []byte) *RawDepthDiff {
	rawDiff := new(RawDepthDiff)
	err := json.Unmarshal(bytes, &rawDiff)
	if err != nil {
		log.Fatal("Error: ", err)
	}

	return rawDiff
}

func (client *BinanceClient) SubscribeDepthDiffStream(symbol string) (chan RawDepthDiff, chan struct{}) {
	diffStream, done := make(chan RawDepthDiff, 10), make(chan struct{})

	conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("wss://stream.binance.com:9443/ws/%s@depth@100ms", symbol), nil)
	if err != nil {
		log.Fatal("Encountered Error: ", err)
	}

	go func() {
		defer conn.Close()

		for {
			select {
			case <-done:
				close(diffStream)
				close(done)

				return
			default:
				_, message, err := conn.ReadMessage()
				if err != nil {
					log.Fatal()
				}

				diffStream <- *FromJsonBytes(message)
			}
		}
	}()

	return diffStream, done
}

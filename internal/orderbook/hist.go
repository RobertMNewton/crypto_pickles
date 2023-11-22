package orderbook

import (
	"encoding/json"
	"log"

	"github.com/kelindar/binary"
	"github.com/vmihailenco/msgpack/v5"
)

type OrderBookHistory struct {
	Symbol  string      `json:"Symbol"`
	Start   OrderBook   `json:"Start"`
	History []DepthDiff `json:"History"`
}

func (hist *OrderBookHistory) GetStartTime() int64 {
	return hist.History[0].Time
}

func (hist *OrderBookHistory) GetEndTime() int64 {
	return hist.History[len(hist.History)-1].Time
}

func HistFromJson(data []byte) OrderBookHistory {
	var hist OrderBookHistory

	err := json.Unmarshal(data, &hist)
	if err != nil {
		log.Printf("failed to decode msg pack encoding: %s", err)
	}

	log.Printf("%v", hist)

	return hist
}

func HistToJson(hist OrderBookHistory) []byte {
	bytes, err := json.Marshal(hist)
	if err != nil {
		log.Fatal(err)
	}

	return bytes
}

func HistFromMsgPack(data []byte) OrderBookHistory {
	var hist OrderBookHistory

	err := msgpack.Unmarshal(data, &hist)
	if err != nil {
		log.Printf("failed to decode msg pack encoding: %s", err)
	}

	return hist
}

func HistToMsgPack(hist OrderBookHistory) []byte {
	bytes, err := msgpack.Marshal(hist)
	if err != nil {
		log.Fatal(err)
	}

	return bytes
}

func HistToBytes(hist OrderBookHistory) []byte {
	bytes, err := binary.Marshal(hist)
	if err != nil {
		log.Fatal(err)
	}

	return bytes
}

func HistFromBytes(data []byte) OrderBookHistory {
	var hist OrderBookHistory

	err := binary.Unmarshal(data, &hist)
	if err != nil {
		log.Printf("failed to decode msg pack encoding: %s", err)
	}

	log.Printf("%v", hist)

	return hist
}

func (hist OrderBookHistory) ToSmallArray(sort bool) []OrderBookSmall {
	obs := make([]OrderBookSmall, len(hist.History)+1)

	currentOB := hist.Start
	obs[0] = currentOB.ToOrderBookSmall()
	if sort {
		obs[0].SortAndCut(5000)
	}

	for i, diff := range hist.History {
		currentOB.ApplyDepthDiff(diff)
		obs[i+1] = currentOB.ToOrderBookSmall()
		if sort {
			obs[i+1].SortAndCut(5000)
		}
	}

	return obs
}

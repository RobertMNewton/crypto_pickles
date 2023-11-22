package orderbook

import (
	"encoding/json"
	"strconv"
)

type DepthLevel map[float32]float32

type OrderBook struct {
	Time int64
	Bids DepthLevel
	Asks DepthLevel
}

type DepthDiff struct {
	Time int64
	Bids DepthLevel
	Asks DepthLevel
}

func (ob *OrderBook) ApplyDepthDiff(diff DepthDiff) OrderBook {
	ob.Time = diff.Time

	for price, volume := range diff.Bids {
		if volume == 0 {
			delete(ob.Bids, price)
		} else {
			ob.Bids[price] = volume
		}
	}

	for price, volume := range diff.Asks {
		if volume == 0 {
			delete(ob.Asks, price)
		} else {
			ob.Asks[price] = volume
		}
	}

	return *ob
}

func (dl DepthLevel) MarshalJSON() ([]byte, error) {
	dlString := make(map[string]string)
	for key, value := range dl {
		dlString[strconv.FormatFloat(float64(key), 'E', -1, 32)] = strconv.FormatFloat(float64(value), 'E', -1, 32)
	}

	return json.Marshal(dlString)
}

func (dl DepthLevel) UnmarshalJSON(data []byte) error {
	var dlString map[string]string
	err := json.Unmarshal(data, &dlString)

	for key, value := range dlString {
		key_f, _ := strconv.ParseFloat(key, 32)
		value_f, _ := strconv.ParseFloat(value, 32)

		dl[float32(key_f)] = float32(value_f)
	}

	return err
}

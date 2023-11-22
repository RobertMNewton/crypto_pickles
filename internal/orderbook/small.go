package orderbook

import (
	"sort"
)

type OrderBookSmall struct {
	Time int64
	Bids PriceLevelArray
	Asks PriceLevelArray
}

type OrderBookSmallArray []OrderBookSmall

func (ob *OrderBookSmall) SortAndCut(limit int) {
	sort.Sort(ob.Bids)
	sort.Sort(ob.Asks)

	if len(ob.Bids) > limit {
		ob.Bids = ob.Bids[(len(ob.Bids))-limit:]
	}

	if len(ob.Asks) > limit {
		ob.Asks = ob.Asks[:limit]
	}
}

func (ob OrderBookSmall) Cut(limit int) OrderBookSmall {
	if len(ob.Bids) > limit {
		ob.Bids = ob.Bids[(len(ob.Bids))-limit:]
	}

	if len(ob.Asks) > limit {
		ob.Asks = ob.Asks[:limit]
	}

	return ob
}

func (ob OrderBook) ToOrderBookSmall() OrderBookSmall {
	obs := OrderBookSmall{
		Time: ob.Time,
		Bids: make(PriceLevelArray, len(ob.Bids)),
		Asks: make(PriceLevelArray, len(ob.Asks)),
	}

	ctr := 0
	for price, volume := range ob.Bids {
		obs.Bids[ctr] = PriceLevel{price, volume}
		ctr++
	}

	ctr = 0
	for price, volume := range ob.Asks {
		obs.Asks[ctr] = PriceLevel{price, volume}
		ctr++
	}

	return obs
}

func (oba OrderBookSmallArray) GetTimeIndex(t int64) int {
	for i := 0; i < len(oba)-1; i++ {
		if oba[i].Time <= t && oba[i+1].Time >= t {
			return i
		}
	}
	return -1
}

func (oba OrderBookSmallArray) Cut(limit, freq int) OrderBookSmallArray {
	iter := 10 / freq
	res := make(OrderBookSmallArray, len(oba)/iter)

	for i := 0; i < len(res); i++ {
		res[i] = oba[i*iter].Cut(limit)
	}

	return res
}

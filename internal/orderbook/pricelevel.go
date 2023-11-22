package orderbook

type PriceLevel [2]float32
type PriceLevelArray []PriceLevel

func (pla PriceLevelArray) Len() int {
	return len(pla)
}

func (pla PriceLevelArray) Swap(i, j int) {
	pla[i], pla[j] = pla[j], pla[i]
}

func (pla PriceLevelArray) Less(i, j int) bool {
	return pla[i][0] < pla[j][0]
}

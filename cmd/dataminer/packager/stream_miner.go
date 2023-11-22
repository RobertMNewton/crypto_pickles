package packager

import (
	"log"
	"strings"
	"time"

	"github.com/crypto_pickle/internal/orderbook"
)

var (
	ORDERBOOK_FRAMES  = 10 * 60 * 5
	CHANGEOVER_FRAMES = 10 * 5
	WS_RESET          = (time.Hour * 23) + (time.Minute * 55)
)

func Configure(obFrames int, cFrames int) {
	ORDERBOOK_FRAMES = obFrames
	CHANGEOVER_FRAMES = cFrames
}

func (packager *Packager) StartStreamMiner(symbol string, depth int32) {
	ticker := time.NewTicker(WS_RESET)

	go func() {
		counter := 0
		history := make([]orderbook.DepthDiff, ORDERBOOK_FRAMES)

		diffStream, done := packager.binance_client.SubscribeDepthDiffStream(strings.ToLower(symbol))
		currentOrderBook := packager.binance_client.GetOrderBook(strings.ToUpper(symbol), depth)

		diff := <-diffStream
		for diff.FirstUpdateId < currentOrderBook.LastUpdateId {
			diff = <-diffStream
		}

		for {
			history[counter] = diff.ToDepthDiff()
			counter += 1

			diff = <-diffStream

			if counter == ORDERBOOK_FRAMES-CHANGEOVER_FRAMES {
				newOrderBook := packager.binance_client.GetOrderBook(strings.ToUpper(symbol), depth)
				newHistory := make([]orderbook.DepthDiff, ORDERBOOK_FRAMES)

				for {
					if diff.FirstUpdateId < newOrderBook.LastUpdateId && counter < len(history) {
						history[counter] = diff.ToDepthDiff()
						counter += 1

						diff = <-diffStream
					} else {
						newHistory[0] = diff.ToDepthDiff()

						ob := currentOrderBook.ToOrderBook()

						packager.histChan <- orderbook.OrderBookHistory{
							Symbol:  symbol,
							Start:   ob.ApplyDepthDiff(history[0]),
							History: history[1:counter],
						}

						currentOrderBook = newOrderBook
						history = newHistory

						counter = 1

						break
					}
				}

				select {
				case <-ticker.C:
					done <- struct{}{}
					diffStream, done = packager.binance_client.SubscribeDepthDiffStream(symbol)

					log.Printf("Reset Peer Connection for symbol %s \n", symbol)
				default:
					continue
				}
			}
		}
	}()
}

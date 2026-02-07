package packager

import (
	"fmt"
	"log"
	"os"

	"github.com/crypto_pickle/cmd/dataminer/binance"
	"github.com/crypto_pickle/internal/orderbook"
	"github.com/crypto_pickle/internal/s3_client"
)

type Packager struct {
	histChan       chan orderbook.OrderBookHistory
	s3_client      *s3_client.S3Client
	local          string
	format         string
	binance_client *binance.BinanceClient
}

func New(bufferLength int, s3 *s3_client.S3Client, local string, format string, binance *binance.BinanceClient) Packager {
	return Packager{
		histChan:       make(chan orderbook.OrderBookHistory, bufferLength),
		s3_client:      s3,
		local:          local,
		format:         format,
		binance_client: binance,
	}
}

func (packager *Packager) Start() {
	go func() {
		for {
			newHist := <-packager.histChan

			var bytes []byte
			if packager.format == "json" {
				bytes = orderbook.HistToJson(newHist)
			}

			name := fmt.Sprintf("%s/%d-%d.%s", newHist.Symbol, newHist.GetStartTime(), newHist.GetEndTime(), packager.format)

			if packager.s3_client != nil {
				go func() {
					packager.s3_client.UploadData("datapickles", name, bytes)
				}()
			}

			if len(packager.local) > 0 {
				go func() {
					if _, err := os.Stat(packager.local); os.IsNotExist(err) {
						if err := os.MkdirAll(packager.local, os.ModePerm); err != nil {
							log.Fatalf("Failed to create directory %s. Got error %s \n", packager.local, err)
						}
					}

					if _, err := os.Stat(packager.local + "/" + newHist.Symbol); os.IsNotExist(err) {
						if err := os.MkdirAll(packager.local+"/"+newHist.Symbol, os.ModePerm); err != nil {
							log.Fatalf("Failed to create sub directory %s. Got error %s \n", packager.local+"/"+newHist.Symbol, err)
						}
					}

					err := os.WriteFile(packager.local+"/"+name, bytes, os.ModePerm)
					if err != nil {
						log.Fatalf("Failed to save file: %s \n", err)
					}
				}()
			}
		}
	}()
}

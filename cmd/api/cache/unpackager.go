package cache

import (
	"github.com/crypto_pickle/internal/orderbook"
	"github.com/crypto_pickle/internal/s3_client"
)

func DownloadOrderBooks(client *s3_client.S3Client, symbol string, key string, format string) []orderbook.OrderBookSmall {
	bytes := client.DownloadData("datapickles", symbol+"/"+key)

	var hist orderbook.OrderBookHistory
	if format == "msgpack" {
		hist = orderbook.HistFromMsgPack(bytes)
	} else if format == "json" {
		hist = orderbook.HistFromMsgPack(bytes)
	} else if format == "bin" {
		hist = orderbook.HistFromBytes(bytes)
	}

	return hist.ToSmallArray(true)
}

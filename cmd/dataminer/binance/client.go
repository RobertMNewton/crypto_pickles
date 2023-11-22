package binance

import (
	"io"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

const (
	API_WEIGHT_LIMIT    = 1200
	API_REQUEST_TIMEOUT = time.Minute

	CONNECTION_LIMIT        = 300
	MAKE_CONNECTION_LIMIT   = 300
	MAKE_CONNECTION_TIMEOUT = 5 * time.Minute
)

type BinanceClient struct {
	api_weight int32
	api_timer  chan struct{}

	connections []Connection
}

func NewClient() BinanceClient {
	return BinanceClient{
		api_weight: API_WEIGHT_LIMIT,
		api_timer:  make(chan struct{}),
	}
}

func (client *BinanceClient) makeAPIRequest(endpoint string, weight int32) []byte {
	for weight > atomic.LoadInt32(&client.api_weight) {
		<-client.api_timer
	}

	atomic.AddInt32(&client.api_weight, -weight)
	go func() {
		<-time.NewTimer(API_REQUEST_TIMEOUT).C
		client.api_timer <- struct{}{}
		atomic.AddInt32(&client.api_weight, weight)
	}()

	resp, err := http.Get("https://api.binance.com/api/" + endpoint)
	if err != nil {
		log.Fatal("Error: ", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error: ", err)
	}

	return bodyBytes
}

func (client *BinanceClient) subscribeStream(streamName string, handler handlerFunc) {
	if client.connections == nil {
		client.connections = make([]Connection, 1)

		client.connections[0] = NewConnection()
		client.connections[0].StartReader()
	} else if len(client.connections) == CONNECTION_LIMIT {
		log.Fatal("Connection Limit Reached!")
	}

	for i := range client.connections {
		res := client.connections[i].SubscribeStream(streamName, handler)
		if res {
			break
		} else if i == len(client.connections)-1 {
			client.connections = append(
				client.connections,
				NewConnection(),
			)

			client.connections[i+1].StartReader()
			client.connections[i+1].SubscribeStream(streamName, handler)
		}
	}
}

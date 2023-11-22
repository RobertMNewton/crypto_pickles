package binance


import (
	"log"
	"sync/atomic"
	"time"
	
	"github.com/gorilla/websocket"
)


const (
	STREAM_LIMIT = 1024

	CONNECTION_REQUEST_LIMIT = 5
	CONNECTION_REQUEST_TIMEOUT = time.Second
)


var messageId int32 = 0


type handlerFunc func([]byte)

type Connection struct {
	conn *websocket.Conn
	streams int32
	requests int32
	timer chan struct{}
}

type Message struct {
	Method string `json:"method"`
	Params []string `json:"params"`
	Id int32 `json:"id"`
}


func getNextMessageId() int32 {
	atomic.AddInt32(&messageId, 1)
	return atomic.LoadInt32(&messageId)
}

func NewConnection() Connection {
	conn, _, err := websocket.DefaultDialer.Dial("wss://stream.binance.com:9443/ws/", nil)
	if err != nil {
		log.Fatal("Encountered Error: ", err)
	}

	return Connection{
		conn: conn,
		streams: 0,
		requests: 0,
		timer: make(chan struct{}),
	}
}

func (connection *Connection) StartReader() {
	go func() {
		for {}
	}()
}

func (connection *Connection) Close() {
	connection.conn.Close()
}

func (connection *Connection) SubscribeStream(streamName string, handler handlerFunc) bool {
	if atomic.LoadInt32(&connection.streams) == STREAM_LIMIT {
		return false
	} else {
		atomic.AddInt32(&connection.streams, 1)
	}

	atomic.AddInt32(&connection.requests, 1)
	go func() {
		<-time.NewTimer(CONNECTION_REQUEST_TIMEOUT).C
		atomic.AddInt32(&connection.requests, -1)
	}()

	err := connection.conn.WriteJSON(Message{Method: "SUBSCRIBE", Params: []string{streamName}, Id: getNextMessageId()})
	if err != nil {
		log.Fatal(err)
	}

	return true
}

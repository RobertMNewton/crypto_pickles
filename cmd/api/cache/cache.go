package cache

import (
	"container/list"
	"errors"
	"sync"
	"time"

	"github.com/crypto_pickle/cmd/api/utils"
	"github.com/crypto_pickle/internal/orderbook"
	"github.com/crypto_pickle/internal/s3_client"
)

type Cache struct {
	client *s3_client.S3Client
	index  Index
	lru    *Lru
	symbol string
	mut    sync.Mutex
}

func NewCache(client *s3_client.S3Client, symbol string, size int) *Cache {
	return &Cache{
		client: client,
		index:  NewIndex(client, symbol),
		lru:    NewLru(size),
		symbol: symbol,
	}
}

func (c *Cache) updateIndex() {
	new_index := NewIndex(c.client, c.symbol)

	c.mut.Lock()
	defer c.mut.Unlock()

	TransferElements(&c.index, &new_index)
	c.index = new_index
}

func (c *Cache) ScheduleUpdateIndex(duration time.Duration) {
	go func() {
		for {
			<-time.NewTimer(duration).C
			c.updateIndex()
		}
	}()
}

func (c *Cache) download(e *IndexElement) {
	newData := DownloadOrderBooks(c.client, c.symbol, e.key, e.format)

	c.lru.Insert(e.key, newData)
	e.downloaded = true
}

func (c *Cache) ClearOnce() {
	c.mut.Lock()
	defer c.mut.Unlock()

	dropKeys := c.lru.Clear()

	for i := 0; i < len(c.index); i += 1 {
		for k := 0; k < len(dropKeys); k += 1 {
			if c.index[i].key == dropKeys[k] {
				c.index[i].downloaded = false
			}
		}
	}
}

func (c *Cache) ScheduleClear(duration time.Duration) {
	go func() {
		for {
			<-time.After(duration)
			c.ClearOnce()
		}
	}()
}

func (c *Cache) SelectWindow(t int, depth, freq int) []orderbook.OrderBookSmall {
	c.mut.Lock()
	defer c.mut.Unlock()

	e, _ := c.index.FindKey(t)
	if e == nil {
		return make([]orderbook.OrderBookSmall, 0)
	}

	if !e.downloaded {
		c.download(e)
	}

	return c.lru.Select(e.key, depth, freq)
}

func (c *Cache) SelectTime(t int, depth, freq int) orderbook.OrderBookSmall {
	c.mut.Lock()
	defer c.mut.Unlock()

	window := c.SelectWindow(t, depth, freq)
	i := orderbook.OrderBookSmallArray(window).GetTimeIndex(int64(t))

	return window[i]
}

func (c *Cache) GetAvailableTimes() (int, int) {
	return c.index.GetEarliestTime(), c.index.GetLatestTime()
}

func (c *Cache) Select(t1, t2 int, depth, freq int) ([]orderbook.OrderBookSmall, error) {
	selection, resLen := list.New(), 0

	c.mut.Lock()
	defer c.mut.Unlock()

	// collecting data from lru (or downloading it if necessary)

	e, i := c.index.FindKey(t1)
	if e == nil {
		return []orderbook.OrderBookSmall{}, errors.New("unable to find time: " + utils.UnixMilliToDateTimeString(t1))
	}

	if !e.downloaded {
		c.download(e)
	}

	winData := c.lru.Select(e.key, depth, freq)

	selection.PushFront(winData)
	resLen += len(winData)

	for e != nil && e.end < t2 {
		e, i = c.index.GetNext(i), i+1
		if e == nil {
			return []orderbook.OrderBookSmall{}, errors.New("unable to find time: " + utils.UnixMilliToDateTimeString(t2))
		}

		if !e.downloaded {
			c.download(e)
		}

		winData = c.lru.Select(e.key, depth, freq)

		selection.PushFront(winData)
		resLen += len(winData)
	}

	// convert selection to a single array

	res := make([]orderbook.OrderBookSmall, resLen)

	if selection.Len() == 1 {
		window := selection.Front().Value.([]orderbook.OrderBookSmall)
		i0, i1 := orderbook.OrderBookSmallArray(window).GetTimeIndex(int64(t1)), orderbook.OrderBookSmallArray(window).GetTimeIndex(int64(t2))

		res = window[i0 : i1+1]
	} else {
		ptr := 0
		for e := selection.Back(); e.Prev() != nil; e = e.Prev() {
			if e.Next() == nil { // The first window
				window := e.Value.([]orderbook.OrderBookSmall)
				i0 := orderbook.OrderBookSmallArray(window).GetTimeIndex(int64(t1))

				for ; ptr < len(window)-i0; ptr += 1 {
					res[ptr] = window[i0+ptr]
				}
			} else if e.Prev() == nil { // The last window
				window := e.Value.([]orderbook.OrderBookSmall)
				i1 := orderbook.OrderBookSmallArray(window).GetTimeIndex(int64(t1))

				for i := 0; i < i1; i += 1 {
					res[ptr] = window[i]
					ptr += 1
				}
			} else {
				window := e.Value.([]orderbook.OrderBookSmall)
				for i := 0; i < len(window); i += 1 {
					res[ptr] = window[i]
					ptr += 1
				}
			}
		}
	}

	return res, nil
}

func (c *Cache) GetInfo() []string {
	res := make([]string, 0, len(c.lru.cacheMap))
	for key := range c.lru.cacheMap {
		res = append(res, key)
	}

	return res
}

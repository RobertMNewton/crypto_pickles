package cache

import (
	"container/list"
	"log"
	"sync"
	"time"

	"github.com/crypto_pickle/internal/orderbook"
)

type cacheElement struct {
	key      string
	lastUsed int64

	data []orderbook.OrderBookSmall

	lruPtr *list.Element
}

type Lru struct {
	cacheMap map[string]*cacheElement
	lruList  *list.List

	mut  sync.Mutex
	size int
}

func NewLru(size int) *Lru {
	return &Lru{
		cacheMap: make(map[string]*cacheElement),
		lruList:  list.New(),
		size:     size,
	}
}

func (c *Lru) Insert(key string, data []orderbook.OrderBookSmall) {
	newCacheElement := &cacheElement{
		key:      key,
		lastUsed: time.Now().UnixMilli(),

		data: data,
	}

	c.mut.Lock()
	defer c.mut.Unlock()

	c.cacheMap[key] = newCacheElement
	newCacheElement.lruPtr = c.lruList.PushFront(newCacheElement)
}

func (c *Lru) Select(key string, depth, freq int) []orderbook.OrderBookSmall {
	c.mut.Lock()
	defer c.mut.Unlock()

	val, ok := c.cacheMap[key]
	if !ok {
		log.Fatalf("Key %s not found in cache}", key)
	}

	val.lastUsed = time.Now().UnixMilli()
	c.lruList.MoveToFront(c.cacheMap[key].lruPtr)

	return []orderbook.OrderBookSmall(orderbook.OrderBookSmallArray(c.cacheMap[key].data).Cut(depth, freq))
}

func (c *Lru) Clear() []string {
	c.mut.Lock()
	defer c.mut.Unlock()

	dropNum := c.lruList.Len() - c.size
	if dropNum > 0 {
		dropKeys := make([]string, dropNum)

		for i := 0; i < dropNum; i += 1 {
			dropEl := c.lruList.Remove(c.lruList.Back()).(*cacheElement)

			delete(c.cacheMap, dropEl.key)
			dropKeys[i] = dropEl.key
		}

		return dropKeys
	} else {
		return []string{}
	}
}

func (c *Lru) Query() bool {
	return c.lruList.Len() > c.size
}

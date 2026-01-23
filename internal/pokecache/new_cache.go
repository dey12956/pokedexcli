package pokecache

import (
	"sync"
	"time"
)

func NewCache(interval time.Duration) *Cache {
	c := &Cache{
		CacheMap: make(map[string]cacheEntry),
		mu:       &sync.RWMutex{},
	}
	go c.reapLoop(interval)
	return c
}

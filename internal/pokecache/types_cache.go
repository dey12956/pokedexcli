package pokecache

import (
	"sync"
	"time"
)

type Cache struct {
	CacheMap map[string]cacheEntry
	mu       *sync.RWMutex
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.CacheMap[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
}

func (c *Cache) Get(key string) (val []byte, exists bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, exists := c.CacheMap[key]
	val = entry.val
	return
}

func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		for key, entry := range c.CacheMap {
			if time.Since(entry.createdAt) > interval {
				delete(c.CacheMap, key)
			}
		}
		c.mu.Unlock()
	}
}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

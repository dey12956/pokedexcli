package pokecache

import (
	"sync"
	"time"
)

type Cache struct {
	CacheMap map[string]cacheEntry
	mu       *sync.RWMutex
	ttl      time.Duration

	reapInterval time.Duration
	maxEntries   int
	stopCh       chan struct{}
	doneCh       chan struct{}
	stopOnce     sync.Once
}

func (c *Cache) Add(key string, val []byte) {
	if c.ttl <= 0 {
		return
	}
	valCopy := make([]byte, len(val))
	copy(valCopy, val)
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.CacheMap[key] = cacheEntry{
		createdAt: now,
		expiresAt: now.Add(c.ttl),
		val:       valCopy,
	}
	c.evictIfNeededLocked()
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	entry, exists := c.CacheMap[key]
	if !exists {
		c.mu.RUnlock()
		return nil, false
	}
	now := time.Now()
	if now.After(entry.expiresAt) {
		c.mu.RUnlock()
		c.mu.Lock()
		entry, exists = c.CacheMap[key]
		if exists && now.After(entry.expiresAt) {
			delete(c.CacheMap, key)
		}
		c.mu.Unlock()
		return nil, false
	}
	valCopy := make([]byte, len(entry.val))
	copy(valCopy, entry.val)
	c.mu.RUnlock()
	return valCopy, true
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.CacheMap, key)
}

func (c *Cache) Close() {
	c.stopOnce.Do(func() {
		close(c.stopCh)
	})
	<-c.doneCh
}

func (c *Cache) evictIfNeededLocked() {
	if c.maxEntries <= 0 || len(c.CacheMap) <= c.maxEntries {
		return
	}
	var oldestKey string
	var oldestTime time.Time
	first := true
	for key, entry := range c.CacheMap {
		if first || entry.createdAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.createdAt
			first = false
		}
	}
	if !first {
		delete(c.CacheMap, oldestKey)
	}
}

func (c *Cache) evictExpiredLocked(now time.Time) {
	for key, entry := range c.CacheMap {
		if now.After(entry.expiresAt) {
			delete(c.CacheMap, key)
		}
	}
}

func (c *Cache) reapLoop() {
	ticker := time.NewTicker(c.reapInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			c.mu.Lock()
			c.evictExpiredLocked(now)
			c.mu.Unlock()
		case <-c.stopCh:
			close(c.doneCh)
			return
		}
	}
}

type cacheEntry struct {
	createdAt time.Time
	expiresAt time.Time
	val       []byte
}

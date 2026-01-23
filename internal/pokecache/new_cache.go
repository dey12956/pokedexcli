package pokecache

import (
	"sync"
	"time"
)

const defaultTTL = 5 * time.Minute
const defaultMaxEntries = 1024

type Option func(*Cache)

func WithMaxEntries(max int) Option {
	return func(c *Cache) {
		if max <= 0 {
			c.maxEntries = 0
			return
		}
		c.maxEntries = max
	}
}

func WithReapInterval(interval time.Duration) Option {
	return func(c *Cache) {
		if interval > 0 {
			c.reapInterval = interval
		}
	}
}

func NewCache(interval time.Duration, opts ...Option) *Cache {
	if interval <= 0 {
		interval = defaultTTL
	}
	c := &Cache{
		CacheMap:     make(map[string]cacheEntry),
		mu:           &sync.RWMutex{},
		ttl:          interval,
		reapInterval: interval,
		maxEntries:   defaultMaxEntries,
		stopCh:       make(chan struct{}),
		doneCh:       make(chan struct{}),
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.reapInterval <= 0 {
		c.reapInterval = c.ttl
	}
	go c.reapLoop()
	return c
}

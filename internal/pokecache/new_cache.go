package pokecache

import (
	"sync"
	"time"
)

const defaultTTL = 5 * time.Minute
const defaultMaxEntries = 1024
const defaultReapInterval = 5 * time.Second

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

func NewCache(reapInterval time.Duration, ttl time.Duration, opts ...Option) *Cache {
	if ttl <= 0 {
		ttl = defaultTTL
	}
	if reapInterval <= 0 {
		reapInterval = defaultReapInterval
	}
	c := &Cache{
		CacheMap:     make(map[string]cacheEntry),
		mu:           &sync.RWMutex{},
		ttl:          ttl,
		reapInterval: reapInterval,
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

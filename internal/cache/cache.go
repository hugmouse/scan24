package cache

import (
	"sync"
	"time"
)

type Cache[K comparable, V any] struct {
	items map[K]item[V]
	mu    sync.RWMutex
	ttl   time.Duration
}

type item[V any] struct {
	value      V
	expiration int64
}

func New[K comparable, V any](ttl time.Duration) *Cache[K, V] {
	c := &Cache[K, V]{
		items: make(map[K]item[V]),
		ttl:   ttl,
	}

	// Periodically clean up
	go c.cleanup()

	return c
}

func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = item[V]{
		value:      value,
		expiration: time.Now().Add(c.ttl).UnixNano(),
	}
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	it, ok := c.items[key]
	if !ok || time.Now().UnixNano() > it.expiration {
		return *new(V), false
	}

	return it.value, true
}

func (c *Cache[K, V]) cleanup() {
	for {
		time.Sleep(c.ttl)

		c.mu.Lock()
		for key, it := range c.items {
			if time.Now().UnixNano() > it.expiration {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

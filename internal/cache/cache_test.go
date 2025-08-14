package cache

import (
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	cache := New[string, string](time.Millisecond * 10)

	cache.Set("key1", "value1")
	value, found := cache.Get("key1")
	if !found || value != "value1" {
		t.Errorf("Expected to find key1 with value 'value1', but got value '%s' and found=%v", value, found)
	}

	// TTL expiration
	time.Sleep(time.Millisecond * 15)
	value, found = cache.Get("key1")
	if found {
		t.Errorf("Expected key1 to be expired, but it was found with value '%s'", value)
	}
}

func TestCache_Cleanup(t *testing.T) {
	cache := New[string, int](time.Millisecond * 10)
	cache.Set("key1", 123)
	cache.Set("key2", 456)

	time.Sleep(time.Millisecond * 15)

	cache.mu.RLock()
	if len(cache.items) > 0 {
		t.Errorf("Expected cache to be empty after cleanup, but it contains %d items", len(cache.items))
	}
	cache.mu.RUnlock()
}

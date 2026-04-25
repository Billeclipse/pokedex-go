package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	entries map[string]cacheEntry
	mu      sync.Mutex
}

func NewCache(interval time.Duration) *Cache {
	cache := &Cache{
		entries: make(map[string]cacheEntry),
	}

	go cache.reapLoop(interval)

	return cache
}

func (cache *Cache) Add(key string, val []byte) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	copiedVal := make([]byte, len(val))
	copy(copiedVal, val)

	cache.entries[key] = cacheEntry{
		createdAt: time.Now(),
		val:       copiedVal,
	}
}

func (cache *Cache) Get(key string) ([]byte, bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	entry, ok := cache.entries[key]
	if !ok {
		return nil, false
	}

	copiedVal := make([]byte, len(entry.val))
	copy(copiedVal, entry.val)

	return copiedVal, true
}

func (cache *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		cache.reap(time.Now(), interval)
	}
}

func (cache *Cache) reap(now time.Time, interval time.Duration) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	for key, entry := range cache.entries {
		if now.Sub(entry.createdAt) > interval {
			delete(cache.entries, key)
		}
	}
}

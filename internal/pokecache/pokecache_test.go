package pokecache

import (
	"bytes"
	"testing"
	"time"
)

func TestCacheAddGet(t *testing.T) {
	cache := NewCache(5 * time.Minute)
	key := "https://example.com/pokemon"
	val := []byte("cached response")

	cache.Add(key, val)

	actual, ok := cache.Get(key)
	if !ok {
		t.Fatalf("expected cache hit for %q", key)
	}

	if !bytes.Equal(actual, val) {
		t.Fatalf("expected %q, got %q", val, actual)
	}
}

func TestCacheReapLoop(t *testing.T) {
	interval := 10 * time.Millisecond
	cache := NewCache(interval)
	key := "https://example.com/expired"

	cache.Add(key, []byte("cached response"))

	deadline := time.After(200 * time.Millisecond)
	for {
		if _, ok := cache.Get(key); !ok {
			return
		}

		select {
		case <-deadline:
			t.Fatalf("expected cache entry to be reaped")
		default:
			time.Sleep(interval)
		}
	}
}

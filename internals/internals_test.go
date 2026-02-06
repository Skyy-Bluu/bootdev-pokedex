package internals

import (
	"fmt"
	"testing"
	"time"
)

func TestAddGet(t *testing.T) {
	const interval = 5 * time.Second
	cases := []struct {
		key string
		val []byte
	}{
		{
			key: "https://example.com",
			val: []byte("testdata"),
		},
		{
			key: "https://example.com/path",
			val: []byte("testdata2"),
		},
	}

	for i, e := range cases {
		t.Run(fmt.Sprintf("Testing case for %v", i), func(t *testing.T) {
			cache := NewCache(interval)
			cache.Add(e.key, e.val)
			val, ok := cache.Get(e.key)

			if !ok {
				t.Errorf("Expected to find key")
				return
			}
			if string(val) != string(e.val) {
				t.Errorf("Expected to find value")
				return
			}
		})
	}
}

func TestReapLoop(t *testing.T) {
	const baseTime = 5 * time.Millisecond
	const waitTime = baseTime + (5 * time.Millisecond)

	cache := NewCache(baseTime)
	cache.Add("https://example.com", []byte("testingdata"))

	_, ok := cache.Get("https://example.com")

	if !ok {
		t.Errorf("Expected to find key")
		return
	}

	time.Sleep(waitTime)

	_, ok = cache.Get("https://example.com")

	if ok {
		t.Errorf("Expected to not find key")
		return
	}
}

package internals

import (
	"fmt"
	"sync"
	"time"
)

type Cache struct {
	cacheMap map[string]cacheEntry
	mu       *sync.Mutex
}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

func NewCache(interval time.Duration) Cache {
	newCache := Cache{
		cacheMap: make(map[string]cacheEntry),
		mu:       &sync.Mutex{},
	}

	go newCache.reapLoop(interval)

	return newCache
}

func (c Cache) Add(key string, value []byte) {
	cacheEntry := cacheEntry{
		createdAt: time.Now(),
		val:       value,
	}

	c.mu.Lock()

	c.cacheMap[key] = cacheEntry

	fmt.Println("Added to cache: ", key)

	defer c.mu.Unlock() // maybe defer
}

func (c Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()

	entry, ok := c.cacheMap[key]

	defer c.mu.Unlock() //maybe defer

	return entry.val, ok
}

func (c Cache) reapLoop(interval time.Duration) {
	timer := time.NewTicker(interval)
	for {
		<-timer.C
		c.mu.Lock()
		//fmt.Println("Checking cache data...")
		for k := range c.cacheMap {
			if c.cacheMap[k].createdAt.Add(interval).Before(time.Now()) {
				fmt.Println("deleting from cache: ", k)
				delete(c.cacheMap, k)
			}
		}
		c.mu.Unlock()
	}
}

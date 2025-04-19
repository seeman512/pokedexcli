package pokecache

import (
	"fmt"
	"os"
	plogger "pokedexcli/internal/logger"
	"sync"
	"time"
)

var logger plogger.Logger = plogger.New(plogger.TRACE, os.Stderr, "CACHE: ")

type CacheEntry struct {
	CreatedAt time.Time
	Val       []byte
}

type Cache struct {
	Data map[string]CacheEntry
	mu   *sync.Mutex
}

func NewCache(interval time.Duration) Cache {
	c := Cache{
		Data: make(map[string]CacheEntry),
		mu:   &sync.Mutex{},
	}

	go c.reapLoop(interval)
	return c
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Data[key] = CacheEntry{
		CreatedAt: time.Now(),
		Val:       val,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.Data[key]
	if !ok {
		return nil, false
	}

	return entry.Val, true
}

func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		<-ticker.C
		now := time.Now()

		for k, v := range c.Data {
			if v.CreatedAt.Add(interval).After(now) {
				continue
			}

			c.mu.Lock()
			logger.Trace(fmt.Sprintf("Key %s expired; %v", k, v.CreatedAt))
			delete(c.Data, k)
			c.mu.Unlock()
		}
	}
}

package pricing

import (
	"sync"
	"time"
)

type Quote struct {
	Pair      string
	Price     int64
	ExpiresAt time.Time
	Source    string
}

type Cache struct {
	mu     sync.RWMutex
	quotes map[string]Quote
}

func NewCache() *Cache {
	return &Cache{
		quotes: make(map[string]Quote),
	}
}

func (c *Cache) Get(pair string) (Quote, bool) {
	c.mu.RLock()
	quote, ok := c.quotes[pair]
	c.mu.RUnlock()
	if !ok {
		return Quote{}, false
	}
	if time.Now().UTC().After(quote.ExpiresAt) {
		return Quote{}, false
	}
	return quote, true
}

func (c *Cache) Set(pair string, price int64, ttl time.Duration, source string) Quote {
	quote := Quote{
		Pair:      pair,
		Price:     price,
		ExpiresAt: time.Now().UTC().Add(ttl),
		Source:    source,
	}
	c.mu.Lock()
	c.quotes[pair] = quote
	c.mu.Unlock()
	return quote
}

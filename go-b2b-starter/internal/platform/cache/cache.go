// Package cache holds bounded, disposable provider metadata. PostgreSQL remains the only state store.
package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

type entry struct {
	value   string
	expires time.Time
}
type Cache struct {
	mu      sync.Mutex
	entries map[string]entry
}

func New() *Cache { return &Cache{entries: make(map[string]entry)} }
func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok || !time.Now().Before(e.expires) {
		delete(c.entries, key)
		return "", errors.New("cache miss")
	}
	return e.value, nil
}
func (c *Cache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) >= 256 {
		clear(c.entries)
	}
	c.entries[key] = entry{value, time.Now().Add(ttl)}
	return nil
}

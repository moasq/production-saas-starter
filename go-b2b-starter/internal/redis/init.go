package redis

import (
	"context"
	"log"
	"sync"
	"time"
)

func InitRedis() (Client, error) {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load Redis configuration: %v", err)
		return nil, err
	}

	client, err := newRedisClient(cfg)
	if err != nil {
		// Fall back to in-memory mock client for development
		log.Printf("Warning: Redis unavailable, using in-memory mock client: %v", err)
		return newMockClient(), nil
	}

	return client, nil
}

// mockClient is an in-memory implementation of Client for development
type mockClient struct {
	mu    sync.RWMutex
	store map[string]mockEntry
}

type mockEntry struct {
	value  string
	expiry time.Time
}

func newMockClient() *mockClient {
	return &mockClient{
		store: make(map[string]mockEntry),
	}
}

func (c *mockClient) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiry time.Time
	if ttl > 0 {
		expiry = time.Now().Add(ttl)
	}

	c.store[key] = mockEntry{
		value:  value.(string),
		expiry: expiry,
	}
	return nil
}

func (c *mockClient) Get(ctx context.Context, key string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.store[key]
	if !ok {
		return "", nil
	}

	if !entry.expiry.IsZero() && time.Now().After(entry.expiry) {
		delete(c.store, key)
		return "", nil
	}

	return entry.value, nil
}

func (c *mockClient) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, key)
	return nil
}

func (c *mockClient) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.store[key]
	if !ok {
		return false, nil
	}

	if !entry.expiry.IsZero() && time.Now().After(entry.expiry) {
		return false, nil
	}

	return true, nil
}

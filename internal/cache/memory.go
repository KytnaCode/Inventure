package cache

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

// MemoryConfig is the configuration for [Memory].
type MemoryConfig struct {
	ExpirationTime time.Duration
	CleanupTime    time.Duration
}

// Memory is an in-memory cache.
type Memory struct {
	c *gocache.Cache
}

// NewMemory creates a new in-memory cache.
func NewMemory(conf *MemoryConfig) *Memory {
	return &Memory{
		c: gocache.New(conf.ExpirationTime, conf.CleanupTime),
	}
}

// Get gets a key from cache, returns nil if not exists.
func (ca *Memory) Get(key string) any {
	v, ok := ca.c.Get(key)
	if !ok {
		return nil
	}

	return v
}

// Set sets a key value in the cache.
func (ca *Memory) Set(key string, value any) {
	ca.c.Set(key, value, 0)
}

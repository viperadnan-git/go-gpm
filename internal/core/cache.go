package core

import "sync"

// TokenCache defines the interface for token storage
type TokenCache interface {
	Get() (token string, expiry int64)
	Set(token string, expiry int64)
}

// MemoryTokenCache stores tokens in memory (thread-safe)
type MemoryTokenCache struct {
	mu     sync.RWMutex
	token  string
	expiry int64
}

// NewMemoryTokenCache creates a new in-memory token cache
func NewMemoryTokenCache() *MemoryTokenCache {
	return &MemoryTokenCache{}
}

// Get retrieves the cached token and expiry
func (c *MemoryTokenCache) Get() (string, int64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.token, c.expiry
}

// Set stores the token with its expiry timestamp
func (c *MemoryTokenCache) Set(token string, expiry int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = token
	c.expiry = expiry
}

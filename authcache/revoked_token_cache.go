package authcache

import (
	"sync"
	"time"
)

type revokedTokenCache struct {
	mu     sync.RWMutex
	tokens map[string]time.Time
}

var cache = &revokedTokenCache{
	tokens: make(map[string]time.Time),
}

func init() {
	go cache.cleanupLoop(10*time.Minute, time.Minute)
}

func Add(token string, expiresAt time.Time) {
	cache.mu.Lock()
	cache.tokens[token] = expiresAt
	cache.mu.Unlock()
}

func IsRevoked(token string) bool {
	cache.mu.RLock()
	expiresAt, exists := cache.tokens[token]
	cache.mu.RUnlock()
	if !exists {
		return false
	}
	if time.Now().After(expiresAt) {
		cache.mu.Lock()
		delete(cache.tokens, token)
		cache.mu.Unlock()
		return false
	}
	return true
}

func (c *revokedTokenCache) cleanupLoop(interval time.Duration, maxAge time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		cutoff := time.Now().Add(-maxAge)
		c.mu.Lock()
		for token, exp := range c.tokens {
			if exp.Before(cutoff) {
				delete(c.tokens, token)
			}
		}
		c.mu.Unlock()
	}
}

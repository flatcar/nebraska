package api

import "time"

// TestCacheManager provides test utilities for cache management in API tests
type TestCacheManager struct{}

// NewTestCacheManager creates a new test cache manager
func NewTestCacheManager() *TestCacheManager {
	return &TestCacheManager{}
}

// SetCacheLifespanForTest temporarily sets the cache lifespan for testing purposes.
// Returns the previous lifespan so it can be restored.
func (tcm *TestCacheManager) SetCacheLifespanForTest(lifespan time.Duration) time.Duration {
	oldLifespan := cachedGroupVersionCountLifespan
	cachedGroupVersionCountLifespan = lifespan
	return oldLifespan
}

// RestoreCacheLifespan restores the cache lifespan to the given value
func (tcm *TestCacheManager) RestoreCacheLifespan(lifespan time.Duration) {
	cachedGroupVersionCountLifespan = lifespan
}

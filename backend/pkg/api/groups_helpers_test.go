package api

import "time"

// TestCacheManager provides test utilities for cache management in API tests
type TestCacheManager struct {
	api *API
}

// NewTestCacheManager creates a new test cache manager for the given API instance
func NewTestCacheManager(api *API) *TestCacheManager {
	return &TestCacheManager{api: api}
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

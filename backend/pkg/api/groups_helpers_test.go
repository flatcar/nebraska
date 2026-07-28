package api

import (
	"time"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/dbreads"
)

// TestCacheManager provides test utilities for cache management in API tests
type TestCacheManager struct{}

// NewTestCacheManager creates a new test cache manager
func NewTestCacheManager() *TestCacheManager {
	return &TestCacheManager{}
}

// SetCacheLifespanForTest temporarily sets the cache lifespan for testing purposes.
// Returns the previous lifespan so it can be restored.
func (tcm *TestCacheManager) SetCacheLifespanForTest(lifespan time.Duration) time.Duration {
	oldLifespan := dbreads.CachedGroupVersionCountLifespan
	dbreads.CachedGroupVersionCountLifespan = lifespan
	return oldLifespan
}

// RestoreCacheLifespan restores the cache lifespan to the given value
func (tcm *TestCacheManager) RestoreCacheLifespan(lifespan time.Duration) {
	dbreads.CachedGroupVersionCountLifespan = lifespan
}

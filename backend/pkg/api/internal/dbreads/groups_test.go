package dbreads

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

// TestGroupVersionCountCacheIsBounded checks that the version count timeline
// cache stops growing once it is full. Entries are keyed by group ID, which is
// caller-supplied, so an unbounded cache would let a caller iterating over
// arbitrary group IDs grow it for the lifetime of the process.
func TestGroupVersionCountCacheIsBounded(t *testing.T) {
	cache := newGroupVersionCountCache()

	entries := maxCachedGroupVersionCountEntries * 2
	for i := 0; i < entries; i++ {
		key := groupDurationCacheKey{GroupID: fmt.Sprintf("group-%d", i), Duration: "1h"}
		cache.Add(key, groupVersionCountCache{
			data:     map[time.Time]types.VersionCountMap{},
			storedAt: time.Now(),
		})
	}

	assert.Equal(t, maxCachedGroupVersionCountEntries, cache.Len(),
		"cache grew past its bound after adding %d distinct group IDs", entries)
}

// TestGroupVersionCountCacheEvictsLeastRecentlyUsed checks that a key that
// keeps being read survives eviction while an untouched one does not.
func TestGroupVersionCountCacheEvictsLeastRecentlyUsed(t *testing.T) {
	cache := newGroupVersionCountCache()

	kept := groupDurationCacheKey{GroupID: "kept", Duration: "1h"}
	evicted := groupDurationCacheKey{GroupID: "evicted", Duration: "1h"}
	entry := groupVersionCountCache{data: map[time.Time]types.VersionCountMap{}, storedAt: time.Now()}

	cache.Add(evicted, entry)
	cache.Add(kept, entry)

	// Fill the cache, reading the key we expect to survive so it stays the
	// most recently used one.
	for i := 0; i < maxCachedGroupVersionCountEntries; i++ {
		_, ok := cache.Get(kept)
		require.True(t, ok, "kept key evicted after %d additions", i)
		cache.Add(groupDurationCacheKey{GroupID: fmt.Sprintf("filler-%d", i), Duration: "1h"}, entry)
	}

	_, ok := cache.Get(kept)
	assert.True(t, ok, "recently used key should have been kept")

	_, ok = cache.Get(evicted)
	assert.False(t, ok, "least recently used key should have been evicted")
}

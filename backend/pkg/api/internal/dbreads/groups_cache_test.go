package dbreads

import (
	"testing"
	"time"
)

func TestPruneExpiredGroupVersionCount(t *testing.T) {
	cachedGroupVersionCountLock.Lock()
	defer cachedGroupVersionCountLock.Unlock()

	previous := cachedGroupVersionCount
	t.Cleanup(func() {
		cachedGroupVersionCountLock.Lock()
		cachedGroupVersionCount = previous
		cachedGroupVersionCountLock.Unlock()
	})

	expired := groupDurationCacheKey{GroupID: "deleted-group", Duration: "1d"}
	fresh := groupDurationCacheKey{GroupID: "live-group", Duration: "1d"}
	now := time.Now()
	cachedGroupVersionCount = map[groupDurationCacheKey]groupVersionCountCache{
		expired: {storedAt: now.Add(-2 * CachedGroupVersionCountLifespan)},
		fresh:   {storedAt: now},
	}

	pruneExpiredGroupVersionCountLocked(now)

	if _, ok := cachedGroupVersionCount[expired]; ok {
		t.Fatal("expired group timeline was kept in the cache")
	}
	if _, ok := cachedGroupVersionCount[fresh]; !ok {
		t.Fatal("fresh group timeline was removed from the cache")
	}
}

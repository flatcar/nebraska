package memcache

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kinvolk/nebraska/backend/pkg/sessions"
)

func TestMain(m *testing.M) {
	if os.Getenv("NEBRASKA_SKIP_TESTS") != "" {
		return
	}

	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	cache := New(mockCopier{})
	assert.NotNil(t, cache)
	assert.IsType(t, &memCache{}, cache)
}

func TestStoreWithHarness(t *testing.T) {
	h := &sessions.TestHarness{
		T:           t,
		NewCache:    newTestCache,
		UseCountFor: useCountFor,
	}
	h.RunBasicSessionLifecycleTests()
	h.RunDeadCookiesTests()
}

func newTestCache() sessions.Cache {
	cache := newCache(mockCopier{})
	s := &testRandomStringer{}
	cache.randomString = s.randomString
	return cache
}

func useCountFor(cache sessions.Cache, id string) int {
	mcache := cache.(*memCache)
	mcache.sessionsLock.Lock()
	defer mcache.sessionsLock.Unlock()
	if info, ok := mcache.sessions[id]; ok {
		return int(info.uses)
	}
	return -1
}

type mockCopier struct{}

var _ ValuesCopier = mockCopier{}

func (mockCopier) Copy(to *sessions.ValuesType, from sessions.ValuesType) error {
	if to == nil {
		*to = make(sessions.ValuesType, len(from))
	}
	for k, v := range from {
		(*to)[k] = v
	}
	return nil
}

type testRandomStringer struct {
	idx int
}

func (s *testRandomStringer) randomString(n int) string {
	s.idx++
	return "id" + strconv.Itoa(s.idx)
}

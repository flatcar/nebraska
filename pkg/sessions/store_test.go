package sessions

import (
	"testing"
)

func TestStoreWithHarness(t *testing.T) {
	h := &TestHarness{
		T: t,
		NewCache: func() Cache {
			return NewMockCache()
		},
		UseCountFor: func(cache Cache, id string) int {
			mockCache := cache.(*MockCache)
			return mockCache.UseCountFor(id)
		},
	}
	h.RunBasicSessionLifecycleTests()
	h.RunDeadCookiesTests()
}

package sessions

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHarness contains a setup for running session tests.
type TestHarness struct {
	// T is just testing.T.
	T *testing.T
	// NewCache should return a Cache set in such a way, that the
	// new IDs for session should be in form of "idN", where N
	// grows from 1.
	NewCache func() Cache
	// UseCountFor should return a use count of a session with a
	// given ID, or -1 if the ID is not in cache.
	UseCountFor func(cache Cache, id string) int
}

// RunBasicSessionLifecycleTests tests creation and destruction of
// sessions, and the use count changes
func (h *TestHarness) RunBasicSessionLifecycleTests() {
	codec := NewMockCodec()
	codec.AddIDValueMapping("id1", "val1")
	cache := h.NewCache()
	store := NewStore(cache, codec)
	// new session
	request1 := newRequest()
	session1 := store.GetSessionUse(request1, "test")
	assert.NoError(h.T, session1.Save(httptest.NewRecorder()))
	assert.Equal(h.T, 1, h.UseCountFor(cache, "id1"))

	// session from cookie
	request2 := newRequestWithCookie("test", "val1")
	session2 := store.GetSessionUse(request2, "test")
	assert.Equal(h.T, 2, h.UseCountFor(cache, "id1"))

	// session from context
	request3 := newRequestWithCookie("test", "val1")
	request3 = request3.WithContext(ContextWithSession(request3.Context(), session2))
	session3 := store.GetSessionUse(request3, "test")
	assert.Equal(h.T, 3, h.UseCountFor(cache, "id1"))

	// order of PutSessionUse is not relevant - they can happen in
	// parallel
	store.PutSessionUse(session3)
	assert.Equal(h.T, 2, h.UseCountFor(cache, "id1"))

	store.PutSessionUse(session1)
	assert.Equal(h.T, 1, h.UseCountFor(cache, "id1"))

	store.PutSessionUse(session2)
	assert.Equal(h.T, 0, h.UseCountFor(cache, "id1"))
}

// RunDeadCookiesTests tests the behaviour of the cache for marked and
// destroyed cookies.
func (h *TestHarness) RunDeadCookiesTests() {
	codec := NewMockCodec()
	codec.AddIDValueMapping("id1", "val1", "id2", "val2")
	cache := h.NewCache()
	store := NewStore(cache, codec)

	// session from cookie
	request1 := newRequestWithCookie("test", "bogusvalue")
	session1 := store.GetSessionUse(request1, "test")
	assert.Equal(h.T, "", session1.ID())

	assert.NoError(h.T, session1.Save(httptest.NewRecorder()))
	assert.Equal(h.T, "id1", session1.ID())
	assert.Equal(h.T, 1, h.UseCountFor(cache, "id1"))
	assert.Equal(h.T, -1, h.UseCountFor(cache, "id2"))

	request2 := newRequestWithCookie("test", "val1")
	session2 := store.GetSessionUse(request2, "test")
	assert.Equal(h.T, "id1", session2.ID())
	assert.Equal(h.T, 2, h.UseCountFor(cache, "id1"))
	assert.Equal(h.T, -1, h.UseCountFor(cache, "id2"))
	session2.Mark()

	request3 := newRequestWithCookie("test", "val1")
	session3 := store.GetSessionUse(request3, "test")
	assert.Equal(h.T, "", session3.ID())
	assert.Equal(h.T, 2, h.UseCountFor(cache, "id1"))
	assert.Equal(h.T, -1, h.UseCountFor(cache, "id2"))

	assert.NoError(h.T, session3.Save(httptest.NewRecorder()))
	assert.Equal(h.T, "id2", session3.ID())
	assert.Equal(h.T, 1, h.UseCountFor(cache, "id2"))

	store.PutSessionUse(session1)
	assert.Equal(h.T, 1, h.UseCountFor(cache, "id1"))

	store.PutSessionUse(session2)
	assert.Equal(h.T, -1, h.UseCountFor(cache, "id1"))

	store.PutSessionUse(session3)
	assert.Equal(h.T, 0, h.UseCountFor(cache, "id2"))

	store.MarkOrDestroySessionByID("id2")
	assert.Equal(h.T, -1, h.UseCountFor(cache, "id2"))
}

func newRequestWithCookie(name, value string) *http.Request {
	request := newRequest()
	cookie := http.Cookie{
		Name:  name,
		Value: value,
	}
	request.AddCookie(&cookie)
	return request
}

func newRequest() *http.Request {
	return httptest.NewRequest("", "/", nil)
}

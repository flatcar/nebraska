package sessions

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionValues(t *testing.T) {
	_, _, session := newMockSessionFast()
	key := "test"

	assert.False(t, session.Has(key))
	assert.Nil(t, session.Get(key))

	session.Set(key, 42)
	assert.True(t, session.Has(key))
	assert.Equal(t, 42, session.Get(key))

	session.Drop(key)
	assert.False(t, session.Has(key))
	assert.Nil(t, session.Get(key))

	session.Set(key, nil)
	assert.True(t, session.Has(key))
	assert.Nil(t, session.Get(key))
}

func TestSessionName(t *testing.T) {
	_, _, session := newMockSessionFast()
	assert.Equal(t, "test", session.Name())
}

func TestSessionID(t *testing.T) {
	_, _, session := newMockSessionFast()
	assert.Equal(t, "", session.ID())
	assert.NoError(t, session.Save(httptest.NewRecorder(), false))
	assert.Equal(t, "id1", session.ID())
}

func TestValidSessionSave(t *testing.T) {
	_, _, session := newMockSessionFast()
	writer := httptest.NewRecorder()
	assert.NoError(t, session.Save(writer, false))
	response := writer.Result()
	cookies := response.Cookies()
	if assert.Len(t, cookies, 1) {
		cookie := cookies[0]
		assert.Equal(t, "test", cookie.Name)
		assert.Equal(t, "val1", cookie.Value)
		assert.Equal(t, 86400*30, cookie.MaxAge)
	}
}

func TestSessionCookieAttributes(t *testing.T) {
	for _, tc := range []struct {
		name       string
		secure     bool
		wantSecure bool
	}{
		{name: "over TLS", secure: true, wantSecure: true},
		{name: "over plain HTTP", secure: false, wantSecure: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, session := newMockSessionFast()
			writer := httptest.NewRecorder()
			assert.NoError(t, session.Save(writer, tc.secure))
			cookies := writer.Result().Cookies()
			if assert.Len(t, cookies, 1) {
				cookie := cookies[0]
				assert.Equal(t, tc.wantSecure, cookie.Secure)
				assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
				assert.True(t, cookie.HttpOnly)
			}
		})
	}
}

func TestExpiredSessionSave(t *testing.T) {
	_, _, session := newMockSessionFast()
	session.Mark()
	writer := httptest.NewRecorder()
	assert.NoError(t, session.Save(writer, false))
	response := writer.Result()
	cookies := response.Cookies()
	if assert.Len(t, cookies, 1) {
		cookie := cookies[0]
		assert.Equal(t, "test", cookie.Name)
		assert.Equal(t, "", cookie.Value)
		assert.Equal(t, -1, cookie.MaxAge)
	}
}

func TestSavedExpiredSessionSave(t *testing.T) {
	_, _, session := newMockSessionFast()
	assert.NoError(t, session.Save(httptest.NewRecorder(), false))
	session.Mark()
	writer := httptest.NewRecorder()
	assert.NoError(t, session.Save(writer, false))
	response := writer.Result()
	cookies := response.Cookies()
	if assert.Len(t, cookies, 1) {
		cookie := cookies[0]
		assert.Equal(t, "test", cookie.Name)
		assert.Equal(t, "", cookie.Value)
		assert.Equal(t, -1, cookie.MaxAge)
	}
}

func TestSessionExt(t *testing.T) {
	extSession := newMockSessionExt("test", NewMockCache(), NewMockCodec())
	assert.Equal(t, "test", extSession.Name())
	assert.Equal(t, "", extSession.ID())
	extSession.SetID("foo")
	assert.Equal(t, "foo", extSession.ID())
	session := extSession.Session()
	session.Set("key", "value")
	values := extSession.GetValues()
	if assert.NotNil(t, values) {
		assert.Len(t, values, 1)
		assert.Equal(t, "value", values["key"])
	}
}

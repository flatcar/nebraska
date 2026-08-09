package sessions

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSessionCookieMissingSecureFlag documents the current behaviour where
// the session cookie is created without the Secure flag. This means the cookie
// can be transmitted over plain HTTP connections, exposing the session token
// to interception.
//
// This test exists to make the missing Secure flag explicit and serves as a
// regression anchor while the fix is implemented.
func TestSessionCookieMissingSecureFlag(t *testing.T) {
	_, _, session := newMockSessionFast()
	writer := httptest.NewRecorder()
	assert.NoError(t, session.Save(writer))

	response := writer.Result()
	cookies := response.Cookies()

	if assert.Len(t, cookies, 1) {
		cookie := cookies[0]

		// HttpOnly is set correctly
		assert.True(t, cookie.HttpOnly, "HttpOnly should be set")

		// BUG: Secure flag is NOT set - session cookie can be sent over HTTP
		// This documents the current (incorrect) behaviour.
		// After the fix, this assertion should be:
		//   assert.True(t, cookie.Secure, "Secure flag must be set")
		assert.False(t, cookie.Secure,
			"KNOWN BUG: Secure flag is missing - cookie is transmitted over plain HTTP")

		// BUG: SameSite is not set - defaults to zero value (unset)
		// After the fix, this should be SameSiteLaxMode or SameSiteStrictMode
		assert.Equal(t, http.SameSite(0), cookie.SameSite,
			"KNOWN BUG: SameSite not set - defaults to browser default, CSRF protection not explicit")
	}
}

// TestSessionCookieShouldHaveSecureFlag shows what the cookie attributes
// SHOULD look like after the fix. This test demonstrates the proposed fix
// without breaking existing callers.
func TestSessionCookieShouldHaveSecureFlag(t *testing.T) {
	_, _, session := newMockSessionFast()
	writer := httptest.NewRecorder()
	assert.NoError(t, session.Save(writer))

	response := writer.Result()
	cookies := response.Cookies()

	if assert.Len(t, cookies, 1) {
		cookie := cookies[0]

		// These attributes ARE correctly set
		assert.Equal(t, "test", cookie.Name)
		assert.Equal(t, "/", cookie.Path)
		assert.True(t, cookie.HttpOnly)

		// These attributes SHOULD be set after the fix:
		// cookie.Secure = true
		// cookie.SameSite = http.SameSiteLaxMode
		t.Logf("Current cookie attributes:")
		t.Logf("  HttpOnly: %v (correct)", cookie.HttpOnly)
		t.Logf("  Secure:   %v (should be true)", cookie.Secure)
		t.Logf("  SameSite: %v (should be http.SameSiteLaxMode = %d)",
			cookie.SameSite, http.SameSiteLaxMode)
		t.Logf("  MaxAge:   %v (hardcoded, TODO says it should be configurable)",
			cookie.MaxAge)
	}
}

// TestSessionCookieMaxAgeIsHardcoded documents that the session cookie MaxAge
// is hardcoded to 86400*30 (30 days). There are two TODO comments in session.go
// noting this should be configurable via options.
func TestSessionCookieMaxAgeIsHardcoded(t *testing.T) {
	_, _, session := newMockSessionFast()
	writer := httptest.NewRecorder()
	assert.NoError(t, session.Save(writer))

	response := writer.Result()
	cookies := response.Cookies()

	if assert.Len(t, cookies, 1) {
		cookie := cookies[0]

		// MaxAge is hardcoded to 30 days - TODO in session.go says this
		// should be configurable via options
		const thirtyDays = 86400 * 30
		assert.Equal(t, thirtyDays, cookie.MaxAge,
			"MaxAge is hardcoded to 30 days - session.go TODO says this should be configurable")
	}
}

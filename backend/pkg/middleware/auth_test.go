package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	if os.Getenv("NEBRASKA_SKIP_TESTS") != "" {
		return
	}

	os.Exit(m.Run())
}

// mockAuthenticator implements the auth.Authenticator interface for testing
type mockAuthenticator struct {
	shouldReply bool
	teamID      string
	callCount   int
}

func (m *mockAuthenticator) Authorize(_ echo.Context) (string, bool) {
	m.callCount++
	return m.teamID, m.shouldReply
}

func (m *mockAuthenticator) Login(_ echo.Context) error {
	return nil
}

func (m *mockAuthenticator) LoginCb(_ echo.Context) error {
	return nil
}

func (m *mockAuthenticator) ValidateToken(_ echo.Context) error {
	return nil
}

func (m *mockAuthenticator) LoginWebhook(_ echo.Context) error {
	return nil
}

func TestAuthSkipper_OIDC_SkippedPaths(t *testing.T) {
	skipper := NewAuthSkipper("oidc")

	testCases := []struct {
		path     string
		expected bool
		desc     string
	}{
		{"/health", true, "health endpoint should be skipped"},
		{"/config", true, "config endpoint should be skipped"},
		{"/flatcar/*", true, "flatcar wildcard pattern itself should be skipped"},
		{"/flatcar/packages", true, "flatcar packages path should be skipped"},
		{"/flatcar/updates", true, "flatcar updates path should be skipped"},
		{"/flatcar/", true, "flatcar root with trailing slash should be skipped"},
		{"/flatcar", true, "flatcar root without trailing slash should be skipped"},
		{"/v1/update", true, "v1/update should be skipped"},
		{"/assets/main.js", true, "assets files should be skipped"},
		{"/assets/styles.css", true, "assets CSS should be skipped"},
		{"/apps", true, "apps path should be skipped for frontend"},
		{"/apps/123", true, "apps with ID should be skipped for frontend"},
		{"/instances", true, "instances path should be skipped for frontend"},
		{"/instances/456", true, "instances with ID should be skipped for frontend"},
		{"/404", true, "404 page should be skipped"},
		{"/auth/callback", true, "OIDC callback should be skipped"},
		{"/api/apps", false, "API endpoints should not be skipped"},
		{"/api/apps/123/groups", false, "API group endpoints should not be skipped"},
		{"/api/apps/123/packages", false, "API package endpoints should not be skipped"},
		{"/", true, "root path should be skipped to serve the FE"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath(tc.path)

			result := skipper(c)
			assert.Equal(t, tc.expected, result, "Path %s should return %v", tc.path, tc.expected)
		})
	}
}

func TestAuthSkipper_GitHub_SkippedPaths(t *testing.T) {
	skipper := NewAuthSkipper("github")

	testCases := []struct {
		path     string
		expected bool
		desc     string
	}{
		{"/health", true, "health endpoint should be skipped"},
		{"/v1/update", true, "v1/update should be skipped"},
		{"/login/cb", true, "login callback should be skipped"},
		{"/login/webhook", true, "login webhook should be skipped"},
		{"/flatcar/packages", true, "flatcar paths should be skipped"},
		{"/assets/main.js", true, "assets files should be skipped"},
		{"/apps", true, "apps path should be skipped for frontend"},
		{"/apps/123", true, "apps with ID should be skipped for frontend"},
		{"/instances", true, "instances path should be skipped for frontend"},
		{"/instances/456", true, "instances with ID should be skipped for frontend"},
		{"/404", true, "404 page should be skipped"},
		{"/", true, "root path should be skipped"},
		{"/config", false, "config endpoint should not be skipped for GitHub"},
		{"/api/apps", false, "API endpoints should not be skipped"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath(tc.path)

			result := skipper(c)
			assert.Equal(t, tc.expected, result, "Path %s should return %v", tc.path, tc.expected)
		})
	}
}

func TestAuthSkipper_UnknownAuth(t *testing.T) {
	skipper := NewAuthSkipper("unknown")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/apps", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/apps")

	result := skipper(c)
	assert.False(t, result, "Unknown auth mode should not skip any paths")
}

func TestAuth_SkipperCalled(t *testing.T) {
	mockAuth := &mockAuthenticator{
		shouldReply: false,
		teamID:      "test-team",
	}

	skipperCalled := false
	mockSkipper := func(_ echo.Context) bool {
		skipperCalled = true
		return true // Skip this request
	}

	config := AuthConfig{Skipper: mockSkipper}
	middleware := Auth(mockAuth, config)

	// Create a dummy handler
	nextCalled := false
	handler := func(_ echo.Context) error {
		nextCalled = true
		return nil
	}

	wrappedHandler := middleware(handler)

	// Create test request
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := wrappedHandler(c)

	require.NoError(t, err)
	assert.True(t, skipperCalled, "Skipper should be called")
	assert.True(t, nextCalled, "Next handler should be called when skipped")
	assert.Equal(t, 0, mockAuth.callCount, "Authenticator should not be called when skipped")
}

func TestAuth_AuthenticatorCalled(t *testing.T) {
	mockAuth := &mockAuthenticator{
		shouldReply: false,
		teamID:      "test-team",
	}

	mockSkipper := func(_ echo.Context) bool {
		return false // Don't skip
	}

	config := AuthConfig{Skipper: mockSkipper}
	middleware := Auth(mockAuth, config)

	// Create a dummy handler
	nextCalled := false
	handler := func(_ echo.Context) error {
		nextCalled = true
		return nil
	}
	wrappedHandler := middleware(handler)

	// Create test request
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/apps", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := wrappedHandler(c)

	require.NoError(t, err)
	assert.Equal(t, 1, mockAuth.callCount, "Authenticator should be called once")
	assert.True(t, nextCalled, "Next handler should be called when auth succeeds")
}

func TestAuth_TeamIDSet(t *testing.T) {
	expectedTeamID := "test-team-123"
	mockAuth := &mockAuthenticator{
		shouldReply: false,
		teamID:      expectedTeamID,
	}

	mockSkipper := func(_ echo.Context) bool {
		return false // Don't skip
	}

	config := AuthConfig{Skipper: mockSkipper}
	middleware := Auth(mockAuth, config)

	// Create a dummy handler that checks team_id
	var actualTeamID interface{}
	handler := func(c echo.Context) error {
		actualTeamID = c.Get("team_id")
		return nil
	}
	wrappedHandler := middleware(handler)

	// Create test request
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/apps", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := wrappedHandler(c)

	require.NoError(t, err)
	assert.Equal(t, expectedTeamID, actualTeamID, "team_id should be set in context")
}

func TestAuth_AuthenticatorReplied(t *testing.T) {
	mockAuth := &mockAuthenticator{
		shouldReply: true, // Authenticator handles the response
		teamID:      "test-team",
	}

	mockSkipper := func(_ echo.Context) bool {
		return false // Don't skip
	}

	config := AuthConfig{Skipper: mockSkipper}
	middleware := Auth(mockAuth, config)

	// Create a dummy handler
	nextCalled := false
	handler := func(_ echo.Context) error {
		nextCalled = true
		return nil
	}
	wrappedHandler := middleware(handler)

	// Create test request
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/apps", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := wrappedHandler(c)

	require.NoError(t, err)
	assert.Equal(t, 1, mockAuth.callCount, "Authenticator should be called once")
	assert.False(t, nextCalled, "Next handler should not be called when authenticator replies")
	assert.Nil(t, c.Get("team_id"), "team_id should not be set when authenticator replies")
}

package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	
	"github.com/kinvolk/nebraska/backend/pkg/middleware"
)

func TestMain(m *testing.M) {
	if os.Getenv("NEBRASKA_SKIP_TESTS") != "" {
		return
	}

	os.Exit(m.Run())
}

// mockAuthenticatorForViewerTest simulates a viewer user
type mockAuthenticatorForViewerTest struct {
	shouldDenyWrite bool
}

func (m *mockAuthenticatorForViewerTest) Authorize(c echo.Context) (string, bool) {
	if m.shouldDenyWrite && (c.Request().Method == "POST" || c.Request().Method == "PUT" || c.Request().Method == "DELETE") {
		c.NoContent(http.StatusForbidden)
		return "", true
	}
	return "test-team", false
}

func (m *mockAuthenticatorForViewerTest) Login(c echo.Context) error {
	return nil
}

func (m *mockAuthenticatorForViewerTest) LoginCb(c echo.Context) error {
	return nil
}

func (m *mockAuthenticatorForViewerTest) LoginToken(c echo.Context) error {
	return nil
}

func (m *mockAuthenticatorForViewerTest) ValidateToken(c echo.Context) error {
	return nil
}

func (m *mockAuthenticatorForViewerTest) LoginWebhook(c echo.Context) error {
	return nil
}

// mockAuthenticatorForAdminTest simulates an admin user
type mockAuthenticatorForAdminTest struct{}

func (m *mockAuthenticatorForAdminTest) Authorize(c echo.Context) (string, bool) {
	return "test-team", false
}

func (m *mockAuthenticatorForAdminTest) Login(c echo.Context) error {
	return nil
}

func (m *mockAuthenticatorForAdminTest) LoginCb(c echo.Context) error {
	return nil
}

func (m *mockAuthenticatorForAdminTest) LoginToken(c echo.Context) error {
	return nil
}

func (m *mockAuthenticatorForAdminTest) ValidateToken(c echo.Context) error {
	return nil
}

func (m *mockAuthenticatorForAdminTest) LoginWebhook(c echo.Context) error {
	return nil
}

// mockAuthenticatorThatDeniesAll denies all requests
type mockAuthenticatorThatDeniesAll struct{}

func (m *mockAuthenticatorThatDeniesAll) Authorize(c echo.Context) (string, bool) {
	c.NoContent(http.StatusForbidden)
	return "", true
}

func (m *mockAuthenticatorThatDeniesAll) Login(c echo.Context) error {
	return nil
}

func (m *mockAuthenticatorThatDeniesAll) LoginCb(c echo.Context) error {
	return nil
}

func (m *mockAuthenticatorThatDeniesAll) LoginToken(c echo.Context) error {
	return nil
}

func (m *mockAuthenticatorThatDeniesAll) ValidateToken(c echo.Context) error {
	return nil
}

func (m *mockAuthenticatorThatDeniesAll) LoginWebhook(c echo.Context) error {
	return nil
}

// Test that authorization happens before JSON validation
func TestUnifiedAuthMiddleware_ViewerDeniedBeforeValidation(t *testing.T) {
	e := echo.New()
	
	mockAuth := &mockAuthenticatorForViewerTest{shouldDenyWrite: true}
	skipper := func(c echo.Context) bool {
		path := c.Path()
		skippedPaths := []string{"/health", "/config"}
		for _, skipPath := range skippedPaths {
			if path == skipPath {
				return true
			}
		}
		return false
	}
	
	e.Use(middleware.Auth(mockAuth, middleware.AuthConfig{Skipper: skipper}))
	e.POST("/api/apps/:id/groups", func(c echo.Context) error {
		t.Error("Handler should not be called when authorization is denied")
		var payload map[string]interface{}
		if err := c.Bind(&payload); err != nil {
			return c.NoContent(http.StatusBadRequest)
		}
		return c.NoContent(http.StatusOK)
	})
	
	// Test with invalid JSON payload (missing required fields)
	invalidJSON := `{"name": "test", "missing_required_fields": true}`
	
	req := httptest.NewRequest(http.MethodPost, "/api/apps/123/groups", strings.NewReader(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	
	e.ServeHTTP(rec, req)
	
	assert.Equal(t, http.StatusForbidden, rec.Code, "Should return 403 Forbidden, not 400 Bad Request")
}

// Test that admin users can proceed to JSON validation
func TestUnifiedAuthMiddleware_AdminProceedsToValidation(t *testing.T) {
	e := echo.New()
	
	mockAuth := &mockAuthenticatorForAdminTest{}
	skipper := func(c echo.Context) bool {
		path := c.Path()
		skippedPaths := []string{"/health", "/config"}
		for _, skipPath := range skippedPaths {
			if path == skipPath {
				return true
			}
		}
		return false
	}
	
	// Apply the unified auth middleware (runs for all HTTP methods)
	e.Use(middleware.Auth(mockAuth, middleware.AuthConfig{Skipper: skipper}))
	
	handlerCalled := false
	// Add a handler that validates JSON and returns appropriate errors
	e.POST("/api/apps/:id/groups", func(c echo.Context) error {
		handlerCalled = true
		
		// Simulate OpenAPI validation that checks required fields
		var payload map[string]interface{}
		if err := c.Bind(&payload); err != nil {
			return c.NoContent(http.StatusBadRequest)
		}
		
		// Check for required fields (simulating OpenAPI validation)
		if _, ok := payload["policy_max_updates_per_period"]; !ok {
			return c.NoContent(http.StatusBadRequest)
		}
		if _, ok := payload["policy_period_interval"]; !ok {
			return c.NoContent(http.StatusBadRequest)
		}
		
		return c.NoContent(http.StatusOK)
	})
	
	// Test with invalid JSON payload
	invalidJSON := `{"name": "test", "missing_required_fields": true}`
	
	req := httptest.NewRequest(http.MethodPost, "/api/apps/123/groups", strings.NewReader(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	
	e.ServeHTTP(rec, req)
	
	// Admin should pass authorization and reach JSON validation, which should return 400
	assert.True(t, handlerCalled, "Handler should be called for admin users")
	assert.Equal(t, http.StatusBadRequest, rec.Code, "Should return 400 Bad Request for invalid JSON")
}

// Test that GET requests work normally with the unified auth middleware
func TestUnifiedAuthMiddleware_GETRequestsWorkNormally(t *testing.T) {
	e := echo.New()
	
	// Create mock authenticator that allows GET but denies writes (like viewer)
	mockAuth := &mockAuthenticatorForViewerTest{shouldDenyWrite: true}
	
	// Create a simple skipper
	skipper := func(c echo.Context) bool {
		return false // Don't skip any paths for this test
	}
	
	// Apply the unified auth middleware (now runs for ALL HTTP methods including GET)
	e.Use(middleware.Auth(mockAuth, middleware.AuthConfig{Skipper: skipper}))
	
	handlerCalled := false
	// Add a GET handler
	e.GET("/api/apps/:id/groups", func(c echo.Context) error {
		handlerCalled = true
		return c.NoContent(http.StatusOK)
	})
	
	req := httptest.NewRequest(http.MethodGet, "/api/apps/123/groups", nil)
	rec := httptest.NewRecorder()
	
	e.ServeHTTP(rec, req)
	
	// GET requests should work normally - viewer can read, just not write
	assert.True(t, handlerCalled, "GET handler should be called")
	assert.Equal(t, http.StatusOK, rec.Code, "GET request should succeed")
}

// Test that skipped paths bypass the unified auth middleware
func TestUnifiedAuthMiddleware_SkippedPathsBypassed(t *testing.T) {
	e := echo.New()
	
	// Create mock authenticator that denies everything
	mockAuth := &mockAuthenticatorThatDeniesAll{}
	
	// Create a skipper that skips /health
	skipper := func(c echo.Context) bool {
		return c.Path() == "/health"
	}
	
	// Apply the unified auth middleware
	e.Use(middleware.Auth(mockAuth, middleware.AuthConfig{Skipper: skipper}))
	
	handlerCalled := false
	// Add a POST handler for /health
	e.POST("/health", func(c echo.Context) error {
		handlerCalled = true
		return c.NoContent(http.StatusOK)
	})
	
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()
	
	e.ServeHTTP(rec, req)
	
	// Skipped paths should bypass authorization entirely
	assert.True(t, handlerCalled, "Handler should be called for skipped paths")
	assert.Equal(t, http.StatusOK, rec.Code, "Skipped paths should succeed even with denying auth")
}
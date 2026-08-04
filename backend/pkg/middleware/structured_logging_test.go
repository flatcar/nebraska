package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStructuredLogging(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		query          string
		handler        echo.HandlerFunc
		expectedStatus int
		expectedFields []string
	}{
		{
			name:   "successful GET request",
			method: http.MethodGet,
			path:   "/api/apps",
			handler: func(c echo.Context) error {
				return c.String(http.StatusOK, "success")
			},
			expectedStatus: http.StatusOK,
			expectedFields: []string{"method", "uri", "path", "status", "duration_ms", "request_id"},
		},
		{
			name:   "POST request with query params",
			method: http.MethodPost,
			path:   "/api/apps",
			query:  "name=test&version=1.0",
			handler: func(c echo.Context) error {
				return c.JSON(http.StatusCreated, map[string]string{"id": "123"})
			},
			expectedStatus: http.StatusCreated,
			expectedFields: []string{"method", "query", "status"},
		},
		{
			name:   "handler returns error",
			method: http.MethodGet,
			path:   "/api/error",
			handler: func(c echo.Context) error {
				return errors.New("test error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedFields: []string{"error", "error_message", "status"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture log output
			var logBuffer bytes.Buffer
			logger := zerolog.New(&logBuffer).With().Timestamp().Logger()

			// Setup Echo
			e := echo.New()
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.query != "" {
				req.URL.RawQuery = tt.query
			}
			req.Header.Set(echo.HeaderXRequestID, "test-request-id")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Create middleware
			middleware := StructuredLogging(logger)
			handler := middleware(tt.handler)

			// Execute
			err := handler(c)

			// Check error handling
			if tt.expectedStatus >= 500 {
				assert.Error(t, err)
			}

			// Parse log output
			logOutput := logBuffer.String()
			assert.NotEmpty(t, logOutput, "log output should not be empty")

			// Verify log is valid JSON
			var logData map[string]interface{}
			err = json.Unmarshal([]byte(logOutput), &logData)
			require.NoError(t, err, "log output should be valid JSON")

			// Verify expected fields are present
			for _, field := range tt.expectedFields {
				assert.Contains(t, logData, field, "log should contain field: %s", field)
			}

			// Verify message field
			assert.Equal(t, "http_request", logData["message"])

			// Verify request_id
			assert.Equal(t, "test-request-id", logData["request_id"])

			// Verify method and path
			assert.Equal(t, tt.method, logData["method"])
			assert.Equal(t, tt.path, logData["path"])
		})
	}
}

func TestStructuredLoggingSkipper(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := zerolog.New(&logBuffer).With().Timestamp().Logger()

	// Create config with skipper that skips /health
	config := StructuredLoggingConfig{
		Logger: logger,
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Request().URL.Path, "/health")
		},
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := StructuredLoggingWithConfig(config)
	handler := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	assert.NoError(t, err)

	// Log should be empty because the request was skipped
	assert.Empty(t, logBuffer.String(), "skipped requests should not be logged")
}

func TestStructuredLoggingDuration(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := zerolog.New(&logBuffer).With().Timestamp().Logger()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/slow", nil)
	req.Header.Set(echo.HeaderXRequestID, "test-duration")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := StructuredLogging(logger)
	handler := middleware(func(c echo.Context) error {
		// Simulate some work
		return c.String(http.StatusOK, "done")
	})

	err := handler(c)
	assert.NoError(t, err)

	// Parse log
	var logData map[string]interface{}
	err = json.Unmarshal(logBuffer.Bytes(), &logData)
	require.NoError(t, err)

	// Verify duration fields exist
	assert.Contains(t, logData, "duration_ms")
	assert.Contains(t, logData, "duration_seconds")

	// Duration should be non-negative
	durationMs, ok := logData["duration_ms"].(float64)
	assert.True(t, ok)
	assert.GreaterOrEqual(t, durationMs, 0.0)
}

func TestStructuredLoggingErrorLevels(t *testing.T) {
	tests := []struct {
		name          string
		status        int
		handler       echo.HandlerFunc
		expectedLevel string
	}{
		{
			name:   "5xx error logs as error",
			status: http.StatusInternalServerError,
			handler: func(c echo.Context) error {
				c.Response().WriteHeader(http.StatusInternalServerError)
				return errors.New("server error")
			},
			expectedLevel: "error",
		},
		{
			name:   "4xx error logs as warning",
			status: http.StatusNotFound,
			handler: func(c echo.Context) error {
				c.Response().WriteHeader(http.StatusNotFound)
				return errors.New("not found")
			},
			expectedLevel: "warn",
		},
		{
			name:   "2xx success logs as info",
			status: http.StatusOK,
			handler: func(c echo.Context) error {
				return c.String(http.StatusOK, "success")
			},
			expectedLevel: "info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logBuffer bytes.Buffer
			logger := zerolog.New(&logBuffer).With().Timestamp().Logger()

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set(echo.HeaderXRequestID, "test-level")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			middleware := StructuredLogging(logger)
			handler := middleware(tt.handler)
			_ = handler(c)

			// Parse log
			var logData map[string]interface{}
			err := json.Unmarshal(logBuffer.Bytes(), &logData)
			require.NoError(t, err)

			// Verify log level
			assert.Equal(t, tt.expectedLevel, logData["level"])
			
			// Verify status code is recorded
			assert.Equal(t, float64(tt.status), logData["status"])
		})
	}
}

func TestRequestLogger(t *testing.T) {
	var logBuffer bytes.Buffer
	baseLogger := zerolog.New(&logBuffer).With().Timestamp().Logger()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set(echo.HeaderXRequestID, "context-test")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Apply RequestLogger middleware
	middleware := RequestLogger(baseLogger)
	handler := middleware(func(c echo.Context) error {
		// Retrieve logger from context
		logger := GetLogger(c)

		// Log something
		logger.Info().Msg("test message from handler")

		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	assert.NoError(t, err)

	// Parse log
	logOutput := logBuffer.String()
	assert.NotEmpty(t, logOutput)

	var logData map[string]interface{}
	err = json.Unmarshal([]byte(logOutput), &logData)
	require.NoError(t, err)

	// Verify logger had request context
	assert.Equal(t, "context-test", logData["request_id"])
	assert.Equal(t, http.MethodGet, logData["method"])
	assert.Equal(t, "/api/test", logData["path"])
	assert.Equal(t, "test message from handler", logData["message"])
}

func TestGetLogger(t *testing.T) {
	t.Run("logger exists in context", func(t *testing.T) {
		var logBuffer bytes.Buffer
		testLogger := zerolog.New(&logBuffer).With().Str("test", "value").Logger()

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Set logger in context
		c.Set("logger", testLogger)

		// Retrieve logger
		logger := GetLogger(c)

		// Use the logger
		logger.Info().Msg("test")

		// Verify it's the correct logger
		assert.Contains(t, logBuffer.String(), "test")
		assert.Contains(t, logBuffer.String(), `"test":"value"`)
	})

	t.Run("no logger in context returns Nop", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Don't set a logger
		logger := GetLogger(c)

		// Should return Nop logger (no panic)
		logger.Info().Msg("this should not panic")
		// No assertion needed - we just verify it doesn't panic
	})
}

func TestStructuredLoggingResponseSize(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := zerolog.New(&logBuffer).With().Timestamp().Logger()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	req.Header.Set(echo.HeaderXRequestID, "size-test")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testData := "This is test response data"
	middleware := StructuredLogging(logger)
	handler := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, testData)
	})

	err := handler(c)
	assert.NoError(t, err)

	// Parse log
	var logData map[string]interface{}
	err = json.Unmarshal(logBuffer.Bytes(), &logData)
	require.NoError(t, err)

	// Verify response_size is tracked
	assert.Contains(t, logData, "response_size")
	responseSize, ok := logData["response_size"].(float64)
	assert.True(t, ok)
	assert.Greater(t, responseSize, float64(0))
}

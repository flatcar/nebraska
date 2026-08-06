package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// StructuredLoggingConfig defines the config for StructuredLogging middleware.
type StructuredLoggingConfig struct {
	// Logger is the zerolog.Logger instance to use for logging
	Logger zerolog.Logger

	// Skipper defines a function to skip middleware
	Skipper func(c echo.Context) bool

	// LogRequestBody enables logging of request body (disabled by default for security)
	LogRequestBody bool

	// LogResponseBody enables logging of response body (disabled by default for performance)
	LogResponseBody bool
}

// DefaultStructuredLoggingConfig returns a default configuration for the middleware.
var DefaultStructuredLoggingConfig = StructuredLoggingConfig{
	Skipper:         func(c echo.Context) bool { return false },
	LogRequestBody:  false,
	LogResponseBody: false,
}

// StructuredLogging returns a middleware that logs HTTP requests in structured JSON format.
// It captures request details, response status, duration, and errors.
func StructuredLogging(logger zerolog.Logger) echo.MiddlewareFunc {
	return StructuredLoggingWithConfig(StructuredLoggingConfig{
		Logger: logger,
	})
}

// StructuredLoggingWithConfig returns a StructuredLogging middleware with config.
func StructuredLoggingWithConfig(config StructuredLoggingConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultStructuredLoggingConfig.Skipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			req := c.Request()
			res := c.Response()
			start := time.Now()

			// Create log event with request context
			logEvent := config.Logger.Info().
				Str("request_id", req.Header.Get(echo.HeaderXRequestID)).
				Str("method", req.Method).
				Str("uri", req.RequestURI).
				Str("path", req.URL.Path).
				Str("remote_ip", c.RealIP()).
				Str("user_agent", req.UserAgent()).
				Str("referer", req.Referer()).
				Str("protocol", req.Proto)

			// Add query parameters if present
			if len(req.URL.RawQuery) > 0 {
				logEvent = logEvent.Str("query", req.URL.RawQuery)
			}

			// Add content type and length
			if contentType := req.Header.Get(echo.HeaderContentType); contentType != "" {
				logEvent = logEvent.Str("content_type", contentType)
			}
			if contentLength := req.ContentLength; contentLength > 0 {
				logEvent = logEvent.Int64("content_length", contentLength)
			}

			// Execute the handler
			err := next(c)

			// Calculate duration
			duration := time.Since(start)

			// Continue building log event with response details
			logEvent = logEvent.
				Int("status", res.Status).
				Int64("response_size", res.Size).
				Dur("duration_ms", duration).
				Float64("duration_seconds", duration.Seconds())

			// Determine log level based on status code and errors
			var finalLogEvent *zerolog.Event
			if err != nil {
				// Add error details
				logEvent = logEvent.
					Err(err).
					Str("error_message", err.Error())

				// Log as error if status is 5xx
				if res.Status >= 500 {
					finalLogEvent = config.Logger.Error()
				} else if res.Status >= 400 {
					// Log as warning if status is 4xx
					finalLogEvent = config.Logger.Warn()
				} else {
					finalLogEvent = config.Logger.Info()
				}
			} else {
				finalLogEvent = config.Logger.Info()
			}

			// Copy fields from logEvent to finalLogEvent
			finalLogEvent.
				Str("request_id", req.Header.Get(echo.HeaderXRequestID)).
				Str("method", req.Method).
				Str("uri", req.RequestURI).
				Str("path", req.URL.Path).
				Str("remote_ip", c.RealIP()).
				Str("user_agent", req.UserAgent()).
				Str("referer", req.Referer()).
				Str("protocol", req.Proto).
				Int("status", res.Status).
				Int64("response_size", res.Size).
				Dur("duration_ms", duration).
				Float64("duration_seconds", duration.Seconds())

			// Add query if present
			if len(req.URL.RawQuery) > 0 {
				finalLogEvent = finalLogEvent.Str("query", req.URL.RawQuery)
			}

			// Add content type and length
			if contentType := req.Header.Get(echo.HeaderContentType); contentType != "" {
				finalLogEvent = finalLogEvent.Str("content_type", contentType)
			}
			if contentLength := req.ContentLength; contentLength > 0 {
				finalLogEvent = finalLogEvent.Int64("content_length", contentLength)
			}

			// Add error if present
			if err != nil {
				finalLogEvent = finalLogEvent.Err(err).Str("error_message", err.Error())
			}

			// Log the request
			finalLogEvent.Msg("http_request")

			return err
		}
	}
}

// RequestLogger creates a middleware that extracts a logger from the context
// and adds request-specific fields to it for use in handlers.
func RequestLogger(baseLogger zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()

			// Create a logger with request-specific context
			logger := baseLogger.With().
				Str("request_id", req.Header.Get(echo.HeaderXRequestID)).
				Str("method", req.Method).
				Str("path", req.URL.Path).
				Str("remote_ip", c.RealIP()).
				Logger()

			// Store logger in context for use in handlers
			c.Set("logger", logger)

			return next(c)
		}
	}
}

// GetLogger retrieves the logger from the Echo context.
// If no logger is found, it returns a disabled logger.
func GetLogger(c echo.Context) zerolog.Logger {
	if logger, ok := c.Get("logger").(zerolog.Logger); ok {
		return logger
	}
	// Return a disabled logger if none is set
	return zerolog.Nop()
}

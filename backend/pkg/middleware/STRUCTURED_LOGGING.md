# Structured Logging Middleware

This package provides structured logging middleware for Nebraska's HTTP API using zerolog. It enables comprehensive request/response tracking, error monitoring, and performance analysis through structured JSON logs.

## Features

- **Structured JSON Logging**: All logs are emitted in JSON format for easy parsing and analysis
- **Request/Response Tracking**: Captures HTTP method, path, query params, status codes, response sizes
- **Performance Metrics**: Records request duration in milliseconds and seconds
- **Error Context**: Automatically logs errors with full context and appropriate severity levels
- **Contextual Logging**: Provides request-scoped loggers for use in handlers
- **Flexible Configuration**: Support for skipping specific endpoints (health checks, metrics)
- **Zero Dependencies**: Uses Nebraska's existing zerolog setup

## Usage

### Basic Setup

Add the structured logging middleware to your Echo server:

```go
import (
    "github.com/flatcar/nebraska/backend/pkg/logger"
    "github.com/flatcar/nebraska/backend/pkg/middleware"
)

// Create a logger instance
l := logger.New("nebraska")

// Add structured logging middleware
e.Use(middleware.StructuredLogging(l))
```

### With Configuration

Use `StructuredLoggingWithConfig` for advanced configuration:

```go
config := middleware.StructuredLoggingConfig{
    Logger: l,
    Skipper: func(c echo.Context) bool {
        // Skip health and metrics endpoints
        path := c.Request().URL.Path
        return path == "/health" || path == "/metrics"
    },
    LogRequestBody: false,  // Keep disabled for security
    LogResponseBody: false, // Keep disabled for performance
}

e.Use(middleware.StructuredLoggingWithConfig(config))
```

### Request-Scoped Logging

For logging within handlers, use `RequestLogger` to inject request-scoped loggers:

```go
// Add RequestLogger middleware
e.Use(middleware.RequestLogger(l))

// Use in handlers
func (h *Handler) GetApp(c echo.Context, appID string) error {
    logger := middleware.GetLogger(c)
    
    logger.Info().Str("app_id", appID).Msg("fetching application")
    
    app, err := h.db.GetApp(appID)
    if err != nil {
        logger.Error().Err(err).Msg("failed to fetch application")
        return err
    }
    
    return c.JSON(http.StatusOK, app)
}
```

## Log Output Format

### Successful Request
```json
{
  "level": "info",
  "request_id": "abc123",
  "method": "GET",
  "uri": "/api/v1/apps",
  "path": "/api/v1/apps",
  "remote_ip": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "protocol": "HTTP/1.1",
  "status": 200,
  "response_size": 1234,
  "duration_ms": 45.2,
  "duration_seconds": 0.0452,
  "time": "2026-01-15T10:30:45Z",
  "message": "http_request"
}
```

### Request with Error (4xx)
```json
{
  "level": "warn",
  "request_id": "def456",
  "method": "POST",
  "path": "/api/v1/apps",
  "status": 404,
  "error": "app not found",
  "error_message": "app not found",
  "duration_ms": 12.5,
  "message": "http_request"
}
```

### Server Error (5xx)
```json
{
  "level": "error",
  "request_id": "ghi789",
  "method": "PUT",
  "path": "/api/v1/apps/123",
  "status": 500,
  "error": "database connection failed",
  "error_message": "database connection failed",
  "duration_ms": 5000.3,
  "message": "http_request"
}
```

## Log Fields Reference

| Field | Type | Description |
|-------|------|-------------|
| `level` | string | Log level: `info`, `warn`, or `error` |
| `request_id` | string | Unique request identifier (X-Request-ID header) |
| `method` | string | HTTP method (GET, POST, PUT, DELETE, etc.) |
| `uri` | string | Full request URI including query string |
| `path` | string | Request path without query string |
| `query` | string | Query parameters (only if present) |
| `remote_ip` | string | Client IP address |
| `user_agent` | string | Client User-Agent header |
| `referer` | string | Referer header (if present) |
| `protocol` | string | HTTP protocol version |
| `content_type` | string | Request Content-Type header (if present) |
| `content_length` | int64 | Request body size in bytes (if present) |
| `status` | int | HTTP response status code |
| `response_size` | int64 | Response body size in bytes |
| `duration_ms` | float64 | Request duration in milliseconds |
| `duration_seconds` | float64 | Request duration in seconds |
| `error` | string | Error details (if error occurred) |
| `error_message` | string | Human-readable error message (if error occurred) |
| `message` | string | Always set to "http_request" |

## Log Levels

The middleware automatically determines the appropriate log level based on response status:

- **info**: Successful requests (2xx, 3xx status codes)
- **warn**: Client errors (4xx status codes)
- **error**: Server errors (5xx status codes)

## Performance Impact

The middleware is designed for minimal performance overhead:

- No reflection or complex serialization
- Efficient zerolog encoding
- Optional request/response body logging (disabled by default)
- Configurable endpoint skipping for high-traffic routes

## Integration with Existing Nebraska Logging

This middleware complements Nebraska's existing logging infrastructure:

1. Uses the same `logger.New()` pattern from `pkg/logger`
2. Respects `NEBRASKA_LOG_FORMAT` environment variable (json/pretty)
3. Compatible with existing zerolog configuration
4. Works alongside Echo's built-in logging

## Configuration Options

### Skipper Function

Skip logging for specific endpoints:

```go
Skipper: func(c echo.Context) bool {
    return middleware.MatchesOneOfPatterns(
        c.Request().URL.Path,
        "/health",
        "/metrics",
        "/assets/*",
    )
}
```

### Log Format

Control log format via environment variable:

```bash
# JSON format (recommended for production)
export NEBRASKA_LOG_FORMAT=json

# Pretty format (human-readable, for development)
export NEBRASKA_LOG_FORMAT=pretty
```

## Best Practices

1. **Always use request-scoped loggers**: Use `GetLogger(c)` in handlers instead of package-level loggers
2. **Skip health/metrics endpoints**: These generate high volume with low value
3. **Keep body logging disabled**: Enable only for debugging, never in production
4. **Use structured fields**: Add context with `.Str()`, `.Int()`, etc., not string concatenation
5. **Consistent context keys**: Use standard field names (app_id, user_id, etc.)

## Example: Complete Setup

```go
package server

import (
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "github.com/flatcar/nebraska/backend/pkg/logger"
    custommiddleware "github.com/flatcar/nebraska/backend/pkg/middleware"
)

func New(conf *config.Config) *echo.Echo {
    e := echo.New()
    l := logger.New("nebraska")

    // Standard middleware
    e.Use(middleware.Recover())
    e.Use(middleware.RequestID())
    e.Use(middleware.CORS())

    // Structured logging (automatic request/response logging)
    skipperFunc := func(c echo.Context) bool {
        return custommiddleware.MatchesOneOfPatterns(
            c.Request().URL.Path,
            "/health",
            "/metrics",
        )
    }
    
    e.Use(custommiddleware.StructuredLoggingWithConfig(
        custommiddleware.StructuredLoggingConfig{
            Logger:  l,
            Skipper: skipperFunc,
        },
    ))

    // Request logger (inject logger into context for handlers)
    e.Use(custommiddleware.RequestLogger(l))

    return e
}
```

## Monitoring and Observability

The structured logs can be easily integrated with log aggregation and monitoring tools:

- **Elasticsearch/Kibana**: Parse JSON logs for dashboards and alerts
- **Splunk**: Index logs for search and analytics
- **Grafana Loki**: Query logs by labels (status, method, path)
- **AWS CloudWatch**: Create metrics from log patterns
- **Datadog**: Analyze request performance and error rates

### Example Queries

**Find slow requests:**
```
duration_seconds > 1.0
```

**Track error rates:**
```
level:error AND status:5*
```

**Monitor specific endpoints:**
```
path:/api/v1/apps AND method:POST
```

## Testing

The middleware includes comprehensive unit tests covering:

- Basic request/response logging
- Error level determination (info/warn/error)
- Skipper functionality
- Duration tracking
- Request-scoped logger injection
- Response size tracking

Run tests:
```bash
cd backend/pkg/middleware
go test -v -run TestStructuredLogging
```

## Troubleshooting

### Logs not appearing

Check that:
1. The middleware is registered with `e.Use()`
2. The logger is properly initialized
3. The endpoint isn't being skipped
4. Log level is set appropriately (zerolog.SetGlobalLevel)

### Duplicate logs

If you see duplicate logs:
1. Ensure you're not using both `StructuredLogging` and Echo's `RequestLogger()`
2. Check for multiple middleware registrations

### Missing request context

If `GetLogger(c)` returns a nop logger:
1. Ensure `RequestLogger` middleware is registered
2. Verify middleware order (RequestLogger should come before handlers)

## Future Enhancements

Potential improvements for future iterations:

- [ ] Request/response body sampling for debugging
- [ ] Automatic PII redaction for sensitive fields
- [ ] Custom field extractors for specific routes
- [ ] Trace ID propagation for distributed tracing
- [ ] Log sampling for high-volume endpoints
- [ ] Integration with OpenTelemetry

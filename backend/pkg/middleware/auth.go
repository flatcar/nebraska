package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/flatcar/nebraska/backend/pkg/auth"
)

func NewAuthSkipper(auth string) middleware.Skipper {
	return func(c echo.Context) bool {
		path := c.Request().URL.Path
		switch auth {
		case "oidc":
			paths := []string{"/", "/auth/callback", "/health", "/config", "/flatcar/*", "/v1/update", "/assets/*", "/apps", "/apps/*", "/instances", "/instances/*", "/404"}
			for _, pattern := range paths {
				if MatchesPattern(pattern, path) {
					return true
				}
			}
		case "github":
			paths := []string{"/", "/health", "/v1/update", "/login/cb", "/login/webhook", "/flatcar/*", "/assets/*", "/apps", "/apps/*", "/instances", "/instances/*", "/404"}
			for _, pattern := range paths {
				if MatchesPattern(pattern, path) {
					return true
				}
			}
		}
		return false
	}
}

// MatchesPattern checks if a path matches a pattern with wildcard support
func MatchesPattern(pattern, path string) bool {
	// Handle empty inputs
	if pattern == "" || path == "" {
		return pattern == path
	}

	// Exact match case
	if !strings.Contains(pattern, "*") {
		return pattern == path
	}

	// Handle /* suffix for matching any subpath
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(path, prefix+"/") || path == prefix
	}

	// For any other wildcard patterns, fall back to exact match
	return pattern == path
}

type AuthConfig struct {
	Skipper middleware.Skipper
}

func Auth(auth auth.Authenticator, conf AuthConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if conf.Skipper(c) {
				return next(c)
			}
			teamID, replied := auth.Authorize(c)
			if replied {
				return nil
			}
			c.Set("team_id", teamID)
			return next(c)
		}
	}
}

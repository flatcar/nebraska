package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/flatcar/nebraska/backend/pkg/auth"
)

var commonPaths = []string{"/", "/health", "/v1/update", "/flatcar/*", "/assets/*", "/apps", "/apps/*", "/instances", "/instances/*", "/404"}

func MatchesOneOfPatterns(path string, patterns ...string) bool {
	for _, pattern := range patterns {
		if MatchesPattern(pattern, path) {
			return true
		}
	}
	return false
}

func NewAuthSkipper(auth string) middleware.Skipper {
	return func(c echo.Context) bool {
		path := c.Request().URL.Path
		if MatchesOneOfPatterns(path, commonPaths...) {
			return true
		}
		switch auth {
		case "oidc":
			return MatchesOneOfPatterns(path, "/auth/callback", "/config")
		case "github":
			return MatchesOneOfPatterns(path, "/login/cb", "/login/webhook")
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

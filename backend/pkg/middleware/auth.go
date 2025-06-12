package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"slices"

	"github.com/kinvolk/nebraska/backend/pkg/auth"
)

func NewAuthSkipper(auth string) middleware.Skipper {
	return func(c echo.Context) bool {
		switch auth {
		case "oidc":
			paths := []string{"/health", "/config", "/*", "/flatcar/*", "/login/cb", "/login/webhook", "/v1/update"}
			if slices.Contains(paths, c.Path()) {
				return true
			}
		case "github":
			paths := []string{"/health", "/v1/update", "/login/cb", "/login/webhook", "/flatcar/*"}
			if slices.Contains(paths, c.Path()) {
				return true
			}
		}

		return false
	}
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

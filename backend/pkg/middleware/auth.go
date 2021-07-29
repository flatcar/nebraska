package middleware

import (
	"github.com/kinvolk/nebraska/backend/pkg/auth"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewAuthSkipper(auth string) middleware.Skipper {
	return func(c echo.Context) bool {
		if auth == "github" {
			paths := []string{"/*", "/login/cb"}
			for _, path := range paths {
				if c.Path() == path {
					return true
				}
			}
		}
		if auth == "oidc" {
			paths := []string{"/login", "/config", "/*", "/login/cb"}
			for _, path := range paths {
				if c.Path() == path {
					return true
				}
			}
		}
		return false
	}
}

type AuthConfig struct {
	Skipper middleware.Skipper
}

func AuthMiddleware(auth auth.Authenticator, conf AuthConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if conf.Skipper(c) {
				return next(c)
			}
			teamID, replied := auth.Authenticate(c)
			if replied {
				return nil
			}
			c.Set("team_id", teamID)
			return next(c)
		}
	}
}

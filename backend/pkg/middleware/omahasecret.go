package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func OmahaSecret(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if strings.HasPrefix(c.Request().URL.Path, "/v1/update") {
				pathSecret := strings.TrimPrefix(c.Request().URL.Path, "/v1/update/")
				if secret == pathSecret {
					c.Request().URL.Path = "/v1/update"
				} else {
					return c.NoContent(http.StatusNotImplemented)
				}
			}
			return next(c)
		}
	}
}

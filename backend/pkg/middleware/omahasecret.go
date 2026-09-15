package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func OmahaSecret(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if strings.HasPrefix(c.Request().URL.Path, "/v1/update") {
				pathSecret := strings.TrimPrefix(c.Request().URL.Path, "/v1/update")
				// Compare in constant time so that the time taken to reject a
				// wrong suffix does not depend on how much of it was correct.
				if subtle.ConstantTimeCompare([]byte(secret), []byte(pathSecret)) == 1 {
					c.Request().URL.Path = "/v1/update"
				} else {
					return c.NoContent(http.StatusNotImplemented)
				}
			}
			return next(c)
		}
	}
}

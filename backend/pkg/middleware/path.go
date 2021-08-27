package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
)

func SanitizePath() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().URL.Path != "/" {
				c.Request().URL.Path = strings.TrimSuffix(c.Request().URL.Path, "/")
			}
			return next(c)
		}
	}
}

// Package shared holds the helpers used by more than one handler sub-package.
package shared

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	echosessions "github.com/flatcar/nebraska/backend/pkg/sessions/echo"
)

// GetTeamID returns the team the authentication middleware placed on the request.
func GetTeamID(c echo.Context) string {
	if val, ok := c.Get("team_id").(string); ok {
		return val
	}
	return ""
}

// LoggerWithUsername returns the given logger tagged with the session username, if there is one.
func LoggerWithUsername(l zerolog.Logger, ctx echo.Context) zerolog.Logger {
	session := echosessions.GetSession(ctx)
	if session == nil {
		return l
	}

	username := session.Get("username")

	return l.With().Str("username", username.(string)).Logger()
}

// AppNotFoundResponse answers a request naming an application that does not exist.
func AppNotFoundResponse(ctx echo.Context, appIDProductID string) error {
	return ctx.JSON(http.StatusBadRequest, map[string]string{
		"message": fmt.Sprintf("App not found for :%s", appIDProductID),
	})
}

package handler

import (
	echosessions "github.com/kinvolk/nebraska/backend/pkg/sessions/echo"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func getTeamID(c echo.Context) string {
	return teamID
}

func loggerWithUsername(l zerolog.Logger, ctx echo.Context) zerolog.Logger {
	session := echosessions.GetSession(ctx)
	if session == nil {
		return logger
	}

	username := session.Get("username")

	return logger.With().Str("username", username.(string)).Logger()
}

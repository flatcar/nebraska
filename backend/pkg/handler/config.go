package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *handler) GetConfig(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, h.clientConf)
}

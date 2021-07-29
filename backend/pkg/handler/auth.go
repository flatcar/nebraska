package handler

import (
	"github.com/labstack/echo/v4"
)

func (h *handler) Login(ctx echo.Context) error {
	return h.auth.Login(ctx)
}

func (h *handler) LoginCb(ctx echo.Context) error {
	return h.auth.LoginCb(ctx)
}

func (h *handler) ValidateToken(ctx echo.Context) error {
	return h.auth.ValidateToken(ctx)
}

func (h *handler) LoginWebhook(ctx echo.Context) error {
	return h.auth.LoginWebhook(ctx)
}

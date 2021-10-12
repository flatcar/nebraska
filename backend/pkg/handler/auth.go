package handler

import (
	"github.com/labstack/echo/v4"
)

func (h *Handler) Login(ctx echo.Context) error {
	return h.auth.Login(ctx)
}

func (h *Handler) LoginCb(ctx echo.Context) error {
	return h.auth.LoginCb(ctx)
}

func (h *Handler) LoginToken(ctx echo.Context) error {
	return h.auth.LoginToken(ctx)
}

func (h *Handler) ValidateToken(ctx echo.Context) error {
	return h.auth.ValidateToken(ctx)
}

func (h *Handler) LoginWebhook(ctx echo.Context) error {
	return h.auth.LoginWebhook(ctx)
}

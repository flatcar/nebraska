package handler

import (
	"github.com/labstack/echo/v4"

	"github.com/kinvolk/nebraska/backend/pkg/codegen"
)

func (h *Handler) Login(ctx echo.Context, params codegen.LoginParams) error {
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

func (h *Handler) LoginWebhook(ctx echo.Context, params codegen.LoginWebhookParams) error {
	return h.auth.LoginWebhook(ctx)
}

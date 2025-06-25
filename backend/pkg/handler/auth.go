package handler

import (
	"github.com/labstack/echo/v4"

	"github.com/flatcar/nebraska/backend/pkg/codegen"
)

// LoginCb handles OAuth callback for GitHub auth mode
// OIDC mode: Returns 501 Not Implemented
// GitHub mode: Processes OAuth callback from GitHub
func (h *Handler) LoginCb(ctx echo.Context) error {
	return h.auth.LoginCb(ctx)
}

// ValidateToken validates JWT access tokens
// OIDC mode: Validates JWT access token signature and expiration
// GitHub mode: Returns 501 Not Implemented
func (h *Handler) ValidateToken(ctx echo.Context) error {
	return h.auth.ValidateToken(ctx)
}

// LoginWebhook handles webhook events for auth providers
// OIDC mode: Not used
// GitHub mode: Processes GitHub webhook events (user/team changes)
func (h *Handler) LoginWebhook(ctx echo.Context, _ codegen.LoginWebhookParams) error {
	return h.auth.LoginWebhook(ctx)
}

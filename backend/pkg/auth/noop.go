package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type (
	// NoopAuthConfig is used to configure the noop authenticator.
	NoopAuthConfig struct {
		// DefaultTeamID is an ID of the team the the noop
		// authenticator will return in its Authenticate
		// function.
		DefaultTeamID string
	}

	noopAuth struct {
		defaultTeamID string
	}
)

var (
	_ Authenticator = &noopAuth{}
)

// NewNoopAuthenticator is an authenticator that does not really
// challenge the user to prove its identity - it will always let
// users' requests through.
func NewNoopAuthenticator(config *NoopAuthConfig) Authenticator {
	return &noopAuth{
		defaultTeamID: config.DefaultTeamID,
	}
}

// SetupRouter is a part of the Authenticator interface
// implementation.
func (noa *noopAuth) SetupRouter(router *echo.Echo) {
}

// Authenticate is a part of the Authenticator interface
// implementation.
func (noa *noopAuth) Authenticate(c echo.Context) (teamID string, replied bool) {
	teamID = noa.defaultTeamID
	replied = false
	return
}

func (noa *noopAuth) Login(ctx echo.Context) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

func (noa *noopAuth) LoginCb(ctx echo.Context) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

func (noa *noopAuth) LoginToken(ctx echo.Context) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

func (noa *noopAuth) ValidateToken(ctx echo.Context) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

func (noa *noopAuth) LoginWebhook(ctx echo.Context) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

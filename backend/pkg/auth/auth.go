package auth

import (
	"github.com/labstack/echo/v4"

	"github.com/kinvolk/nebraska/backend/pkg/util"
)

var (
	logger = util.NewLogger("auth")
)

// Authenticator provides a way to authorize a user sending an HTTP
// request with a bearer token.
type Authenticator interface {
	// Authorize validates the bearer token and checks if the user is
	// authorized to access the resource. It should return an ID of a team
	// from the database if user is authorized. If authorization fails,
	// the function should do the 401/403 HTTP reply and return true for
	// the "replied" return value.
	Authorize(c echo.Context) (teamID string, replied bool)

	Login(ctx echo.Context) error

	LoginCb(ctx echo.Context) error

	LoginToken(ctx echo.Context) error

	ValidateToken(ctx echo.Context) error

	LoginWebhook(ctx echo.Context) error
}

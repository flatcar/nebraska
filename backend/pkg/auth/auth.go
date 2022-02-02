package auth

import (
	"github.com/labstack/echo/v4"

	"github.com/kinvolk/nebraska/backend/pkg/util"
)

var (
	logger = util.NewLogger("auth")
)

// Authenticator provides a way to authenticate a user sending an HTTP
// request.
type Authenticator interface {
	// Authenticate checks if the user is authenticated. It should
	// return an ID of a team from the database if user is
	// authenticated. If authentication fails, the function should
	// do the 403 HTTP reply and return true for the "replied"
	// return value. Similar should happen if the authentication
	// routine requires redirection - issue a redirection HTTP
	// reply and return true for "replied".
	Authenticate(c echo.Context) (teamID string, replied bool)

	Login(ctx echo.Context) error

	LoginCb(ctx echo.Context) error

	LoginToken(ctx echo.Context) error

	ValidateToken(ctx echo.Context) error

	LoginWebhook(ctx echo.Context) error
}

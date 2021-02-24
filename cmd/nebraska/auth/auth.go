package auth

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/kinvolk/nebraska/cmd/nebraska/ginhelpers"
)

var (
	logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(
		zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
			e.Str("context", "auth")
		}))
)

// Authenticator provides a way to authenticate a user sending an HTTP
// request.
type Authenticator interface {
	// SetupRouter allows the authenticator to add more
	// middlewares and routes to the router. It may be useful for
	// authenticators talking to third party services that provide
	// authentication functionality.
	SetupRouter(router ginhelpers.Router)
	// Authenticate checks if the user is authenticated. It should
	// return an ID of a team from the database if user is
	// authenticated. If authentication fails, the function should
	// do the 403 HTTP reply and return true for the "replied"
	// return value. Similar should happen if the authentication
	// routine requires redirection - issue a redirection HTTP
	// reply and return true for "replied".
	Authenticate(c *gin.Context) (teamID string, replied bool)
}

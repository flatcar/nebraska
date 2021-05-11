package auth

import (
	"github.com/gin-gonic/gin"

	"github.com/kinvolk/nebraska/backend/cmd/nebraska/ginhelpers"
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
func (noa *noopAuth) SetupRouter(router ginhelpers.Router) {
}

// Authenticate is a part of the Authenticator interface
// implementation.
func (noa *noopAuth) Authenticate(c *gin.Context) (teamID string, replied bool) {
	teamID = noa.defaultTeamID
	replied = false
	return
}

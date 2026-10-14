// Package admin serves the endpoints that write admin-owned tables, which only
// a node accepting admin writes serves. It writes runtime state only through
// the one-method LocalUpdatesBrake.
package admin

import (
	apiadmin "github.com/flatcar/nebraska/backend/pkg/api/admin"
	"github.com/flatcar/nebraska/backend/pkg/config"
	"github.com/flatcar/nebraska/backend/pkg/logger"
)

var l = logger.New("handler/admin")

// LocalUpdatesBrake releases the node-local safe-mode brake. A single-instance
// deployment presents one updates-enabled switch, so re-enabling updates on the
// group has to release the brake this node tripped for itself.
type LocalUpdatesBrake interface {
	ClearUpdatesEnabledOverride(groupID string) error
}

// Handler serves the endpoints that write admin-owned tables.
type Handler struct {
	admin *apiadmin.Service
	brake LocalUpdatesBrake
	conf  *config.Config
}

// New creates the handler for the endpoints that write admin-owned tables.
func New(adminSvc *apiadmin.Service, brake LocalUpdatesBrake, conf *config.Config) *Handler {
	return &Handler{
		admin: adminSvc,
		brake: brake,
		conf:  conf,
	}
}

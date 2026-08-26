// Package runtime serves the endpoints available on every node, whatever its
// instance mode. It holds the runtime write surface and never the admin one.
package runtime

import (
	apiruntime "github.com/flatcar/nebraska/backend/pkg/api/runtime"
	"github.com/flatcar/nebraska/backend/pkg/config"
	"github.com/flatcar/nebraska/backend/pkg/logger"
	"github.com/flatcar/nebraska/backend/pkg/omaha"
)

const UpdateMaxRequestSize = 64 * 1024

var l = logger.New("handler/runtime")

var defaultPage = 1
var defaultPerPage = 10

// Handler serves the endpoints that read, or write runtime-owned tables,
// including the Omaha entry point.
type Handler struct {
	runtime      *apiruntime.Service
	omahaHandler *omaha.Handler
	conf         *config.Config
}

// New creates the handler for the endpoints served on every node.
func New(runtimeSvc *apiruntime.Service, conf *config.Config) *Handler {
	return &Handler{
		runtime:      runtimeSvc,
		omahaHandler: omaha.NewHandler(runtimeSvc),
		conf:         conf,
	}
}

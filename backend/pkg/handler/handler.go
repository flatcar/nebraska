package handler

import (
	"github.com/flatcar/nebraska/backend/pkg/codegen"
	"github.com/flatcar/nebraska/backend/pkg/handler/admin"
	"github.com/flatcar/nebraska/backend/pkg/handler/base"
	"github.com/flatcar/nebraska/backend/pkg/handler/runtime"
)

// Go names an embedded field after its unqualified type name, and all three
// sub-handlers are called Handler, so they are embedded through these aliases.
type (
	adminHandler   = admin.Handler
	baseHandler    = base.Handler
	runtimeHandler = runtime.Handler
)

// Handler composes the role-scoped sub-handlers and serves nothing itself, so
// an endpoint is only reachable through the handler its node role allows.
type Handler struct {
	*adminHandler
	*baseHandler
	*runtimeHandler
}

var _ codegen.ServerInterface = (*Handler)(nil)

func New(adminH *admin.Handler, baseH *base.Handler, runtimeH *runtime.Handler) *Handler {
	return &Handler{
		adminHandler:   adminH,
		baseHandler:    baseH,
		runtimeHandler: runtimeH,
	}
}

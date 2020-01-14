package ginhelpers

import (
	"github.com/gin-gonic/gin"
)

// Router is an interface quite similar to the gin's IRouter
// interface, but with different Group function to allow creating
// named routers.
type Router interface {
	gin.IRoutes
	Group(relativePath, name string, handlers ...gin.HandlerFunc) Router
}

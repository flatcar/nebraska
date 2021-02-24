package main

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kinvolk/nebraska/cmd/nebraska/ginhelpers"
)

const requestIDKey = "github.com/kinvolk/nebraska/request-id"

var requestID uint64

func setupUsedRouterLogging(router gin.IRoutes, name string) {
	router.Use(func(c *gin.Context) {
		reqID, ok := c.Get(requestIDKey)
		if !ok {
			reqID = -1
		}
		logger.Debug().Msgf("router debug request id %s router name %s", reqID, name)
		c.Next()
	})
}

func setupRequestLifetimeLogging(router gin.IRoutes) {
	router.Use(func(c *gin.Context) {
		reqID := atomic.AddUint64(&requestID, 1)
		c.Set(requestIDKey, reqID)

		start := time.Now()
		logger.Debug().Msgf("request debug request ID %d start time %s method %s URL %s client IP %s", reqID, start, c.Request.Method, c.Request.URL.String(), c.ClientIP())

		// Process request
		c.Next()

		stop := time.Now()
		latency := stop.Sub(start)
		logger.Debug().Msgf("request debug request ID %d stop time %s latency %s status %d", reqID, stop, latency, c.Writer.Status())
	})
}

type wrappedRouter struct {
	router  gin.IRouter
	httpLog bool
}

func wrapRouter(router gin.IRouter, httpLog bool) ginhelpers.Router {
	return &wrappedRouter{
		router:  router,
		httpLog: httpLog,
	}
}

var _ ginhelpers.Router = &wrappedRouter{}

func (r *wrappedRouter) Use(handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.router.Use(handlers...)
}

func (r *wrappedRouter) Handle(httpMethod, relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.router.Handle(httpMethod, relativePath, handlers...)
}

func (r *wrappedRouter) Any(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.router.Any(relativePath, handlers...)
}

func (r *wrappedRouter) GET(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.router.GET(relativePath, handlers...)
}

func (r *wrappedRouter) POST(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.router.POST(relativePath, handlers...)
}

func (r *wrappedRouter) DELETE(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.router.DELETE(relativePath, handlers...)
}

func (r *wrappedRouter) PATCH(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.router.PATCH(relativePath, handlers...)
}

func (r *wrappedRouter) PUT(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.router.PUT(relativePath, handlers...)
}

func (r *wrappedRouter) OPTIONS(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.router.OPTIONS(relativePath, handlers...)
}

func (r *wrappedRouter) HEAD(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.router.HEAD(relativePath, handlers...)
}

func (r *wrappedRouter) StaticFile(relativePath, filePath string) gin.IRoutes {
	return r.router.StaticFile(relativePath, filePath)
}

func (r *wrappedRouter) Static(relativePath, root string) gin.IRoutes {
	return r.router.Static(relativePath, root)
}

func (r *wrappedRouter) StaticFS(relativePath string, fs http.FileSystem) gin.IRoutes {
	return r.router.StaticFS(relativePath, fs)
}

func (r *wrappedRouter) Group(relativePath, name string, handlers ...gin.HandlerFunc) ginhelpers.Router {
	group := r.router.Group(relativePath, handlers...)
	if r.httpLog {
		setupUsedRouterLogging(group, name)
	}
	return wrapRouter(group, r.httpLog)
}

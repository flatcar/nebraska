package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/mgutz/logxi/v1"

	ginsessions "github.com/kinvolk/nebraska/pkg/sessions/gin"
)

var (
	enableSyncer        = flag.Bool("enable-syncer", false, "Enable Flatcar packages syncer")
	hostFlatcarPackages = flag.Bool("host-flatcar-packages", false, "Host Flatcar packages in Nebraska")
	flatcarPackagesPath = flag.String("flatcar-packages-path", "", "Path where Flatcar packages files should be stored")
	nebraskaURL         = flag.String("nebraska-url", "", "nebraska URL (http://host:port - required when hosting Flatcar packages in nebraska)")
	httpLog             = flag.Bool("http-log", false, "Enable http requests logging")
	httpStaticDir       = flag.String("http-static-dir", "../frontend/built", "Path to frontend static files")
	clientID            = flag.String("client-id", "", fmt.Sprintf("Client ID used for authentication; can be taken from %s env var too", clientIDEnvName))
	clientSecret        = flag.String("client-secret", "", fmt.Sprintf("Client secret used for authentication; can be taken from %s env var too", clientSecretEnvName))
	sessionAuthKey      = flag.String("session-secret", "", fmt.Sprintf("Session secret used authenticating sessions in cookies, will be generated if none is passed; can be taken from %s env var too", sessionAuthKeyEnvName))
	sessionCryptKey     = flag.String("session-crypt-key", "", fmt.Sprintf("Session key used for encrypting sessions in cookies, will be generated if none is passed; can be taken from %s env var too", sessionCryptKeyEnvName))
	webhookSecret       = flag.String("webhook-secret", "", fmt.Sprintf("Webhook secret used for validing webhook messages; can be taken from %s env var too", webhookSecretEnvName))
	readWriteTeams      = flag.String("rw-teams", "", "comma-separated list of read-write teams in the org/team format")
	readOnlyTeams       = flag.String("ro-teams", "", "comma-separated list of read-only teams in the org/team format")
	logger              = log.New("nebraska")
)

func main() {
	flag.Parse()

	if err := checkArgs(); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	conf := &controllerConfig{
		enableSyncer:        *enableSyncer,
		hostFlatcarPackages: *hostFlatcarPackages,
		flatcarPackagesPath: *flatcarPackagesPath,
		nebraskaURL:         *nebraskaURL,
		sessionAuthKey:      *sessionAuthKey,
		sessionCryptKey:     *sessionCryptKey,
		oauthClientID:       *clientID,
		oauthClientSecret:   *clientSecret,
		webhookSecret:       *webhookSecret,
		readWriteTeams:      strings.Split(*readWriteTeams, ","),
		readOnlyTeams:       strings.Split(*readOnlyTeams, ","),
	}
	ctl, err := newController(conf)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer ctl.close()

	engine := setupRoutes(ctl, *httpLog)

	var params []string
	if os.Getenv("PORT") == "" {
		params = append(params, ":8000")
	}
	err = engine.Run(params...)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func checkArgs() error {
	if *hostFlatcarPackages {
		if *flatcarPackagesPath == "" {
			return errors.New("Invalid Flatcar packages path. Please ensure you provide a valid path using -flatcar-packages-path")
		}
		tmpFile, err := ioutil.TempFile(*flatcarPackagesPath, "")
		if err != nil {
			return errors.New("Invalid Flatcar packages path: " + err.Error())
		}
		defer os.Remove(tmpFile.Name())

		if _, err := url.ParseRequestURI(*nebraskaURL); err != nil {
			return errors.New("invalid Nebraska URL, please ensure the value provided using -nebraska-url is a valid url")
		}
	}

	return nil
}

const requestIDKey = "github.com/kinvolk/nebraska/request-id"

func setupRouter(router gin.IRoutes, name string, httpLog bool) {
	if httpLog {
		router.Use(func(c *gin.Context) {
			reqID, ok := c.Get(requestIDKey)
			if !ok {
				reqID = -1
			}
			logger.Debug("router debug",
				"request id", reqID,
				"router name", name,
			)
			c.Next()
		})
	}
}

var requestID uint64

func setupRoutes(ctl *controller, httpLog bool) *gin.Engine {
	engine := gin.New()
	if httpLog {
		engine.Use(func(c *gin.Context) {
			reqID := atomic.AddUint64(&requestID, 1)
			c.Set(requestIDKey, reqID)

			start := time.Now()
			logger.Debug("request debug",
				"request ID", reqID,
				"start time", start,
				"method", c.Request.Method,
				"URL", c.Request.URL.String(),
				"client IP", c.ClientIP(),
			)

			// Process request
			c.Next()

			stop := time.Now()
			latency := stop.Sub(start)
			logger.Debug("request debug",
				"request ID", reqID,
				"stop time", stop,
				"latency", latency,
				"status", c.Writer.Status(),
			)
		})
	}
	engine.Use(gin.Recovery())
	setupRouter(engine, "top", httpLog)
	engine.Use(ginsessions.SessionsMiddleware(ctl.sessionsStore, "nebraska"))
	// API router setup
	apiRouter := engine.Group("/api")
	setupRouter(apiRouter, "api", httpLog)
	apiRouter.Use(ctl.authenticate)

	// API routes

	// Users
	apiRouter.PUT("/password", ctl.updateUserPassword)

	// Applications
	apiRouter.POST("/apps", ctl.addApp)
	apiRouter.PUT("/apps/:app_id", ctl.updateApp)
	apiRouter.DELETE("/apps/:app_id", ctl.deleteApp)
	apiRouter.GET("/apps/:app_id", ctl.getApp)
	apiRouter.GET("/apps", ctl.getApps)

	// Groups
	apiRouter.POST("/apps/:app_id/groups", ctl.addGroup)
	apiRouter.PUT("/apps/:app_id/groups/:group_id", ctl.updateGroup)
	apiRouter.DELETE("/apps/:app_id/groups/:group_id", ctl.deleteGroup)
	apiRouter.GET("/apps/:app_id/groups/:group_id", ctl.getGroup)
	apiRouter.GET("/apps/:app_id/groups", ctl.getGroups)
	apiRouter.GET("/apps/:app_id/groups/:group_id/version_timeline", ctl.getGroupVersionCountTimeline)
	apiRouter.GET("/apps/:app_id/groups/:group_id/status_timeline", ctl.getGroupStatusCountTimeline)

	// Channels
	apiRouter.POST("/apps/:app_id/channels", ctl.addChannel)
	apiRouter.PUT("/apps/:app_id/channels/:channel_id", ctl.updateChannel)
	apiRouter.DELETE("/apps/:app_id/channels/:channel_id", ctl.deleteChannel)
	apiRouter.GET("/apps/:app_id/channels/:channel_id", ctl.getChannel)
	apiRouter.GET("/apps/:app_id/channels", ctl.getChannels)

	// Packages
	apiRouter.POST("/apps/:app_id/packages", ctl.addPackage)
	apiRouter.PUT("/apps/:app_id/packages/:package_id", ctl.updatePackage)
	apiRouter.DELETE("/apps/:app_id/packages/:package_id", ctl.deletePackage)
	apiRouter.GET("/apps/:app_id/packages/:package_id", ctl.getPackage)
	apiRouter.GET("/apps/:app_id/packages", ctl.getPackages)

	// Instances
	apiRouter.GET("/apps/:app_id/groups/:group_id/instances/:instance_id/status_history", ctl.getInstanceStatusHistory)
	apiRouter.GET("/apps/:app_id/groups/:group_id/instances", ctl.getInstances)

	// Activity
	apiRouter.GET("/activity", ctl.getActivity)

	// Omaha server router setup
	omahaRouter := engine.Group("/")
	setupRouter(omahaRouter, "omaha", httpLog)
	omahaRouter.POST("/omaha", ctl.processOmahaRequest)
	omahaRouter.POST("/v1/update", ctl.processOmahaRequest)

	// Host Flatcar packages payloads
	if *hostFlatcarPackages {
		flatcarPkgsRouter := engine.Group("/flatcar")
		setupRouter(flatcarPkgsRouter, "flatcar", httpLog)
		flatcarPkgsRouter.Static("/", *flatcarPackagesPath)
	}

	// Metrics
	metricsRouter := engine.Group("/metrics")
	setupRouter(metricsRouter, "metrics", httpLog)
	metricsRouter.Use(ctl.authenticate)
	metricsRouter.GET("/", ctl.getMetrics)

	// oauth
	oauthRouter := engine.Group("/login")
	setupRouter(oauthRouter, "oauth", httpLog)
	oauthRouter.GET("/cb", ctl.loginCb)
	oauthRouter.POST("/webhook", ctl.loginWebhook)

	// Serve frontend static content
	staticRouter := engine.Group("/")
	setupRouter(staticRouter, "static", httpLog)
	staticRouter.Use(ctl.authenticate)
	staticRouter.StaticFile("", filepath.Join(*httpStaticDir, "index.html"))
	for _, file := range []string{"index.html", "favicon.png"} {
		staticRouter.StaticFile(file, filepath.Join(*httpStaticDir, file))
	}
	for _, dir := range []string{"js", "font", "img"} {
		staticRouter.Static(dir, filepath.Join(*httpStaticDir, dir))
	}

	return engine
}

package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	log "github.com/mgutz/logxi/v1"
)

var (
	enableSyncer        = flag.Bool("enable-syncer", false, "Enable Flatcar packages syncer")
	hostFlatcarPackages = flag.Bool("host-flatcar-packages", false, "Host Flatcar packages in Nebraska")
	flatcarPackagesPath = flag.String("flatcar-packages-path", "", "Path where Flatcar packages files should be stored")
	nebraskaURL         = flag.String("nebraska-url", "", "nebraska URL (http://host:port - required when hosting Flatcar packages in nebraska)")
	httpLog             = flag.Bool("http-log", false, "Enable http requests logging")
	httpStaticDir       = flag.String("http-static-dir", "../frontend/built", "Path to frontend static files")
	logger              = log.New("nebraska")
)

func main() {
	if err := mainWithError(); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func mainWithError() error {
	flag.Parse()

	if err := checkArgs(); err != nil {
		return err
	}

	conf := &controllerConfig{
		enableSyncer:        *enableSyncer,
		hostFlatcarPackages: *hostFlatcarPackages,
		flatcarPackagesPath: *flatcarPackagesPath,
		nebraskaURL:         *nebraskaURL,
	}
	ctl, err := newController(conf)
	if err != nil {
		return err
	}
	defer ctl.close()

	engine := setupRoutes(ctl, *httpLog)

	var params []string
	if os.Getenv("PORT") == "" {
		params = append(params, ":8000")
	}
	return engine.Run(params...)
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

func setupRouter(router gin.IRoutes, name string, httpLog bool) {
	if httpLog {
		setupUsedRouterLogging(router, name)
	}
}

func setupRoutes(ctl *controller, httpLog bool) *gin.Engine {
	engine := gin.New()
	if httpLog {
		setupRequestLifetimeLogging(engine)
	}
	engine.Use(gin.Recovery())
	setupRouter(engine, "top", httpLog)
	wrappedEngine := wrapRouter(engine, httpLog)

	// API router setup
	apiRouter := wrappedEngine.Group("/api", "api")
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
	omahaRouter := wrappedEngine.Group("/", "omaha")
	omahaRouter.POST("/omaha", ctl.processOmahaRequest)
	omahaRouter.POST("/v1/update", ctl.processOmahaRequest)

	// Host Flatcar packages payloads
	if *hostFlatcarPackages {
		flatcarPkgsRouter := wrappedEngine.Group("/flatcar", "flatcar")
		flatcarPkgsRouter.Static("/", *flatcarPackagesPath)
	}

	// Metrics
	metricsRouter := wrappedEngine.Group("/metrics", "metrics")
	setupRouter(metricsRouter, "metrics", httpLog)
	metricsRouter.Use(ctl.authenticate)
	metricsRouter.GET("/", ctl.getMetrics)

	// Serve frontend static content
	staticRouter := wrappedEngine.Group("/", "static")
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

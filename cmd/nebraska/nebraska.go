package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/mgutz/logxi/v1"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
)

const (
	flatcarPkgsRouterPrefix = "/flatcar/"
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
	sessionSecret       = flag.String("session-secret", "", fmt.Sprintf("Session secret used for storing sessions, will be generated if none is passed; can be taken from %s env var too", sessionSecretEnvName))
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
		sessionSecret:       *sessionSecret,
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

	setupRoutes(ctl)

	if !*httpLog {
		goji.Abandon(middleware.Logger)
	}
	goji.Serve()
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
			return errors.New("Invalid Nebraska URL. Please ensure the value provided using -nebraska-url is a valid url.")
		}
	}

	return nil
}

func setupRouter(router *web.Mux, name string) {
	router.Use(func(c *web.C, h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			logger.Debug("router debug", "request", fmt.Sprintf("%s %s", r.Method, r.URL.String()), "router name", name)
			h.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	})
}

func setupRoutes(ctl *controller) {
	setupRouter(goji.DefaultMux, "top")
	goji.Use(ctl.sessions.Middleware())
	// API router setup
	apiRouter := web.New()
	setupRouter(apiRouter, "api")
	apiRouter.Use(ctl.authenticate)
	goji.Handle("/api/*", apiRouter)

	// API routes

	// Users
	apiRouter.Put("/api/password", ctl.updateUserPassword)

	// Applications
	apiRouter.Post("/api/apps", ctl.addApp)
	apiRouter.Put("/api/apps/:app_id", ctl.updateApp)
	apiRouter.Delete("/api/apps/:app_id", ctl.deleteApp)
	apiRouter.Get("/api/apps/:app_id", ctl.getApp)
	apiRouter.Get("/api/apps", ctl.getApps)

	// Groups
	apiRouter.Post("/api/apps/:app_id/groups", ctl.addGroup)
	apiRouter.Put("/api/apps/:app_id/groups/:group_id", ctl.updateGroup)
	apiRouter.Delete("/api/apps/:app_id/groups/:group_id", ctl.deleteGroup)
	apiRouter.Get("/api/apps/:app_id/groups/:group_id", ctl.getGroup)
	apiRouter.Get("/api/apps/:app_id/groups", ctl.getGroups)

	// Channels
	apiRouter.Post("/api/apps/:app_id/channels", ctl.addChannel)
	apiRouter.Put("/api/apps/:app_id/channels/:channel_id", ctl.updateChannel)
	apiRouter.Delete("/api/apps/:app_id/channels/:channel_id", ctl.deleteChannel)
	apiRouter.Get("/api/apps/:app_id/channels/:channel_id", ctl.getChannel)
	apiRouter.Get("/api/apps/:app_id/channels", ctl.getChannels)

	// Packages
	apiRouter.Post("/api/apps/:app_id/packages", ctl.addPackage)
	apiRouter.Put("/api/apps/:app_id/packages/:package_id", ctl.updatePackage)
	apiRouter.Delete("/api/apps/:app_id/packages/:package_id", ctl.deletePackage)
	apiRouter.Get("/api/apps/:app_id/packages/:package_id", ctl.getPackage)
	apiRouter.Get("/api/apps/:app_id/packages", ctl.getPackages)

	// Instances
	apiRouter.Get("/api/apps/:app_id/groups/:group_id/instances/:instance_id/status_history", ctl.getInstanceStatusHistory)
	apiRouter.Get("/api/apps/:app_id/groups/:group_id/instances", ctl.getInstances)

	// Activity
	apiRouter.Get("/api/activity", ctl.getActivity)

	// Omaha server router setup
	omahaRouter := web.New()
	setupRouter(omahaRouter, "omaha")
	omahaRouter.Use(middleware.SubRouter)
	goji.Handle("/omaha/*", omahaRouter)
	goji.Handle("/v1/update/*", omahaRouter)

	// Omaha server routes
	omahaRouter.Post("/", ctl.processOmahaRequest)

	// Host Flatcar packages payloads
	if *hostFlatcarPackages {
		flatcarPkgsRouter := web.New()
		setupRouter(flatcarPkgsRouter, "flatcar")
		flatcarPkgsRouter.Use(middleware.SubRouter)
		goji.Handle(flatcarPkgsRouterPrefix+"*", flatcarPkgsRouter)
		flatcarPkgsRouter.Handle("/*", http.FileServer(http.Dir(*flatcarPackagesPath)))
	}

	// Metrics
	metricsRouter := web.New()
	setupRouter(metricsRouter, "metrics")
	metricsRouter.Use(ctl.authenticate)
	goji.Handle("/metrics", metricsRouter)
	metricsRouter.Get("/metrics", ctl.getMetrics)

	// oauth
	oauthRouter := web.New()
	setupRouter(oauthRouter, "oauth")
	goji.Handle("/login/*", oauthRouter)
	oauthRouter.Get("/login/cb", ctl.loginCb)
	oauthRouter.Post("/login/webhook", ctl.loginWebhook)

	// Serve frontend static content
	staticRouter := web.New()
	setupRouter(staticRouter, "static")
	staticRouter.Use(ctl.authenticate)
	goji.Handle("/*", staticRouter)
	staticRouter.Handle("/*", http.FileServer(http.Dir(*httpStaticDir)))
}

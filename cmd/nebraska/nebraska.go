package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/Depado/ginprom"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/kinvolk/nebraska/cmd/nebraska/auth"
	"github.com/kinvolk/nebraska/pkg/api"
	"github.com/kinvolk/nebraska/pkg/random"
)

const (
	ghClientIDEnvName        = "NEBRASKA_GITHUB_OAUTH_CLIENT_ID"
	ghClientSecretEnvName    = "NEBRASKA_GITHUB_OAUTH_CLIENT_SECRET"
	ghSessionAuthKeyEnvName  = "NEBRASKA_GITHUB_SESSION_SECRET"
	ghSessionCryptKeyEnvName = "NEBRASKA_GITHUB_SESSION_CRYPT_KEY"
	ghWebhookSecretEnvName   = "NEBRASKA_GITHUB_WEBHOOK_SECRET"
	ghEnterpriseURLEnvName   = "NEBRASKA_GITHUB_ENTERPRISE_URL"
)

var (
	enableSyncer        = flag.Bool("enable-syncer", false, "Enable Flatcar packages syncer")
	hostFlatcarPackages = flag.Bool("host-flatcar-packages", false, "Host Flatcar packages in Nebraska")
	flatcarPackagesPath = flag.String("flatcar-packages-path", "", "Path where Flatcar packages files should be stored")
	nebraskaURL         = flag.String("nebraska-url", "", "nebraska URL (http://host:port - required when hosting Flatcar packages in nebraska)")
	httpLog             = flag.Bool("http-log", false, "Enable http requests logging")
	httpStaticDir       = flag.String("http-static-dir", "../frontend/built", "Path to frontend static files")
	authMode            = flag.String("auth-mode", "github", "authentication mode, available modes: noop, github")
	ghClientID          = flag.String("gh-client-id", "", fmt.Sprintf("GitHub client ID used for authentication; can be taken from %s env var too", ghClientIDEnvName))
	ghClientSecret      = flag.String("gh-client-secret", "", fmt.Sprintf("GitHub client secret used for authentication; can be taken from %s env var too", ghClientSecretEnvName))
	ghSessionAuthKey    = flag.String("gh-session-secret", "", fmt.Sprintf("Session secret used for authenticating sessions in cookies used for storing GitHub info , will be generated if none is passed; can be taken from %s env var too", ghSessionAuthKeyEnvName))
	ghSessionCryptKey   = flag.String("gh-session-crypt-key", "", fmt.Sprintf("Session key used for encrypting sessions in cookies used for storing GitHub info, will be generated if none is passed; can be taken from %s env var too", ghSessionCryptKeyEnvName))
	ghWebhookSecret     = flag.String("gh-webhook-secret", "", fmt.Sprintf("GitHub webhook secret used for validing webhook messages; can be taken from %s env var too", ghWebhookSecretEnvName))
	ghReadWriteTeams    = flag.String("gh-rw-teams", "", "comma-separated list of read-write GitHub teams in the org/team format")
	ghReadOnlyTeams     = flag.String("gh-ro-teams", "", "comma-separated list of read-only GitHub teams in the org/team format")
	ghEnterpriseURL     = flag.String("gh-enterprise-url", "", fmt.Sprintf("base URL of the enterprise instance if using GHE; can be taken from %s env var too", ghEnterpriseURLEnvName))
	logger              = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(
		zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
			e.Str("context", "nebraska")
		}))
	flatcarUpdatesURL = flag.String("sync-update-url", "https://public.update.flatcar-linux.net/v1/update/", "Flatcar update URL to sync from")
	checkFrequencyVal = flag.String("sync-interval", "1h", "Sync check interval (the minimum depends on the number of channels to sync, e.g., 8m for 8 channels incl. different architectures)")
	appLogoPath       = flag.String("client-logo", "", "Client app logo, should be a path to svg file")
	appTitle          = flag.String("client-title", "", "Client app title")
	appHeaderStyle    = flag.String("client-header-style", "light", "Client app header style, should be either dark or light")
	apiEndpointSuffix = flag.String("api-endpoint-suffix", "", "Additional suffix for the API endpoint to serve Omaha clients on; use a secret to only serve your clients, e.g., mysecret results in /v1/update/mysecret")
)

func main() {
	if err := mainWithError(); err != nil {
		logger.Error().Err(err).Send()
		os.Exit(1)
	}
}

func mainWithError() error {
	flag.Parse()

	if err := checkArgs(); err != nil {
		return err
	}

	api, err := api.New()
	if err != nil {
		return err
	}

	var (
		noopAuthConfig *auth.NoopAuthConfig
		ghAuthConfig   *auth.GithubAuthConfig
	)

	switch *authMode {
	case "noop":
		defaultTeam, err := api.GetTeam()
		if err != nil {
			return err
		}
		noopAuthConfig = &auth.NoopAuthConfig{
			DefaultTeamID: defaultTeam.ID,
		}
	case "github":
		defaultTeam, err := api.GetTeam()
		if err != nil {
			return err
		}
		oauthClientID, err := obtainOAuthClientID(*ghClientID)
		if err != nil {
			return err
		}
		oauthClientSecret, err := obtainOAuthClientSecret(*ghClientSecret)
		if err != nil {
			return err
		}
		ghWebhookSecret, err := obtainWebhookSecret(*ghWebhookSecret)
		if err != nil {
			return err
		}
		// enterprise URL is optional
		ghEnterpriseURL, _ := obtainEnterpriseURL(*ghEnterpriseURL)
		ghAuthConfig = &auth.GithubAuthConfig{
			SessionAuthKey:    obtainSessionAuthKey(*ghSessionAuthKey),
			SessionCryptKey:   obtainSessionCryptKey(*ghSessionCryptKey),
			OAuthClientID:     oauthClientID,
			OAuthClientSecret: oauthClientSecret,
			WebhookSecret:     ghWebhookSecret,
			ReadWriteTeams:    strings.Split(*ghReadWriteTeams, ","),
			ReadOnlyTeams:     strings.Split(*ghReadOnlyTeams, ","),
			DefaultTeamID:     defaultTeam.ID,
			EnterpriseURL:     ghEnterpriseURL,
		}
	default:
		return fmt.Errorf("unknown auth mode %q", *authMode)
	}

	checkFrequency, err := time.ParseDuration(*checkFrequencyVal)
	if err != nil {
		return err
	}
	conf := &controllerConfig{
		api:                 api,
		enableSyncer:        *enableSyncer,
		hostFlatcarPackages: *hostFlatcarPackages,
		flatcarPackagesPath: *flatcarPackagesPath,
		nebraskaURL:         *nebraskaURL,
		noopAuthConfig:      noopAuthConfig,
		githubAuthConfig:    ghAuthConfig,
		flatcarUpdatesURL:   *flatcarUpdatesURL,
		checkFrequency:      checkFrequency,
	}
	ctl, err := newController(conf)
	if err != nil {
		return err
	}
	defer ctl.close()

	engine := setupRoutes(ctl, *httpLog)

	// Register Application metrics and Instrument.
	err = registerAndInstrumentMetrics(ctl)
	if err != nil {
		return err
	}

	var params []string
	if os.Getenv("PORT") == "" {
		params = append(params, ":8000")
	}
	return engine.Run(params...)
}

func obtainSessionAuthKey(potentialSecret string) []byte {
	if secret := getPotentialOrEnv(potentialSecret, ghSessionAuthKeyEnvName); secret != "" {
		return []byte(secret)
	}
	return random.Data(64)
}

func obtainSessionCryptKey(potentialKey string) []byte {
	if key := getPotentialOrEnv(potentialKey, ghSessionCryptKeyEnvName); key != "" {
		return []byte(key)
	}
	return random.Data(32)
}

func obtainOAuthClientID(potentialID string) (string, error) {
	if id := getPotentialOrEnv(potentialID, ghClientIDEnvName); id != "" {
		return id, nil
	}
	return "", errors.New("no oauth client ID")
}

func obtainOAuthClientSecret(potentialSecret string) (string, error) {
	if secret := getPotentialOrEnv(potentialSecret, ghClientSecretEnvName); secret != "" {
		return secret, nil
	}
	return "", errors.New("no oauth client secret")
}

func obtainWebhookSecret(potentialSecret string) (string, error) {
	if secret := getPotentialOrEnv(potentialSecret, ghWebhookSecretEnvName); secret != "" {
		return secret, nil
	}
	return "", errors.New("no webhook secret")
}

func obtainEnterpriseURL(potentialURL string) (string, error) {
	if secret := getPotentialOrEnv(potentialURL, ghEnterpriseURLEnvName); secret != "" {
		return secret, nil
	}
	return "", errors.New("no enterprise URL")
}

func getPotentialOrEnv(potentialValue, envName string) string {
	if potentialValue != "" {
		return potentialValue
	}
	return os.Getenv(envName)
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

	// Setup Middlewares

	// Recovery middleware to recover from panics
	engine.Use(gin.Recovery())

	setupRouter(engine, "top", httpLog)

	// Prometheus Metrics Middleware
	p := ginprom.New(
		ginprom.Engine(engine),
		ginprom.Namespace("nebraska"),
		ginprom.Subsystem("gin"),
		ginprom.Path("/metrics"),
	)
	engine.Use(p.Instrument())

	wrappedEngine := wrapRouter(engine, httpLog)

	ctl.auth.SetupRouter(wrappedEngine)

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
	apiRouter.GET("/apps/:app_id/groups/:group_id/instances_stats", ctl.getGroupInstancesStats)
	apiRouter.GET("/apps/:app_id/groups/:group_id/version_breakdown", ctl.getGroupVersionBreakdown)

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
	apiRouter.GET("/apps/:app_id/groups/:group_id/instancescount", ctl.getInstancesCount)
	apiRouter.GET("/apps/:app_id/groups/:group_id/instances/:instance_id", ctl.getInstance)
	apiRouter.PUT("/instances/:instance_id", ctl.updateInstance)

	// Activity
	apiRouter.GET("/activity", ctl.getActivity)

	// Omaha server router setup
	omahaRouter := wrappedEngine.Group("/", "omaha")
	omahaRouter.POST("/omaha", ctl.processOmahaRequest)
	omahaRouter.POST(path.Join("/v1/update", *apiEndpointSuffix), ctl.processOmahaRequest)

	// Config router setup
	configRouter := wrappedEngine.Group("/config", "config")
	configRouter.Use(ctl.authenticate)
	configRouter.GET("/", ctl.getConfig)

	// Host Flatcar packages payloads
	if *hostFlatcarPackages {
		flatcarPkgsRouter := wrappedEngine.Group("/flatcar", "flatcar")
		flatcarPkgsRouter.Static("/", *flatcarPackagesPath)
	}

	// Serve frontend static content
	staticRouter := wrappedEngine.Group("/", "static")
	staticRouter.Use(ctl.authenticate)
	staticRouter.StaticFile("", filepath.Join(*httpStaticDir, "index.html"))
	for _, file := range []string{"index.html", "favicon.png", "robots.txt"} {
		staticRouter.StaticFile(file, filepath.Join(*httpStaticDir, file))
	}
	for _, dir := range []string{"static"} {
		staticRouter.Static(dir, filepath.Join(*httpStaticDir, dir))
	}
	// catch all route with static middleware
	engine.NoRoute(func(c *gin.Context) {
		c.File(filepath.Join(*httpStaticDir, "index.html"))
	})
	return engine
}

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

	"github.com/gin-gonic/gin"
	log "github.com/mgutz/logxi/v1"

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
	conf := &controllerConfig{
		api:                 api,
		enableSyncer:        *enableSyncer,
		hostFlatcarPackages: *hostFlatcarPackages,
		flatcarPackagesPath: *flatcarPackagesPath,
		nebraskaURL:         *nebraskaURL,
		noopAuthConfig:      noopAuthConfig,
		githubAuthConfig:    ghAuthConfig,
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
	if id := getPotentialOrEnv(potentialID, ghClientIDEnvName); potentialID != "" {
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
	engine.Use(gin.Recovery())
	setupRouter(engine, "top", httpLog)
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

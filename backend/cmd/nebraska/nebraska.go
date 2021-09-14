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
	"github.com/gin-contrib/requestid"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	swagger "github.com/kinvolk/nebraska/backend/api"

	"github.com/kinvolk/nebraska/backend/cmd/nebraska/auth"
	"github.com/kinvolk/nebraska/backend/pkg/api"
	"github.com/kinvolk/nebraska/backend/pkg/random"
	"github.com/kinvolk/nebraska/backend/pkg/util"
)

const (
	ghClientIDEnvName          = "NEBRASKA_GITHUB_OAUTH_CLIENT_ID"
	ghClientSecretEnvName      = "NEBRASKA_GITHUB_OAUTH_CLIENT_SECRET"
	ghSessionAuthKeyEnvName    = "NEBRASKA_GITHUB_SESSION_SECRET"
	ghSessionCryptKeyEnvName   = "NEBRASKA_GITHUB_SESSION_CRYPT_KEY"
	ghWebhookSecretEnvName     = "NEBRASKA_GITHUB_WEBHOOK_SECRET"
	ghEnterpriseURLEnvName     = "NEBRASKA_GITHUB_ENTERPRISE_URL"
	oidcClientIDEnvName        = "NEBRASKA_OIDC_CLIENT_ID"
	oidcClientSecretEnvName    = "NEBRASKA_OIDC_CLIENT_SECRET"
	oidcSessionAuthKeyEnvName  = "NEBRASKA_OIDC_SESSION_SECRET"
	oidcSessionCryptKeyEnvName = "NEBRASKA_OIDC_SESSION_CRYPT_KEY"
)

var (
	enableSyncer          = flag.Bool("enable-syncer", false, "Enable Flatcar packages syncer")
	hostFlatcarPackages   = flag.Bool("host-flatcar-packages", false, "Host Flatcar packages in Nebraska")
	flatcarPackagesPath   = flag.String("flatcar-packages-path", "", "Path where Flatcar packages files should be stored")
	nebraskaURL           = flag.String("nebraska-url", "http://localhost:8000", "nebraska URL (http://host:port - required when hosting Flatcar packages in nebraska)")
	syncerPkgsURL         = flag.String("syncer-packages-url", "", "use this URL instead of the original one for packages created by the syncer; any {{ARCH}} and {{VERSION}} in the URL will be replaced by the original package's architecture and version, respectively. If this option is not used but the 'host-flatcar-packages' one is, then the URL will be nebraska-url/flatcar/ .")
	httpLog               = flag.Bool("http-log", false, "Enable http requests logging")
	httpStaticDir         = flag.String("http-static-dir", "../frontend/build", "Path to frontend static files")
	authMode              = flag.String("auth-mode", "github", "authentication mode, available modes: noop, github, oidc")
	ghClientID            = flag.String("gh-client-id", "", fmt.Sprintf("GitHub client ID used for authentication; can be taken from %s env var too", ghClientIDEnvName))
	ghClientSecret        = flag.String("gh-client-secret", "", fmt.Sprintf("GitHub client secret used for authentication; can be taken from %s env var too", ghClientSecretEnvName))
	ghSessionAuthKey      = flag.String("gh-session-secret", "", fmt.Sprintf("Session secret used for authenticating sessions in cookies used for storing GitHub info , will be generated if none is passed; can be taken from %s env var too", ghSessionAuthKeyEnvName))
	ghSessionCryptKey     = flag.String("gh-session-crypt-key", "", fmt.Sprintf("Session key used for encrypting sessions in cookies used for storing GitHub info, will be generated if none is passed; can be taken from %s env var too", ghSessionCryptKeyEnvName))
	ghWebhookSecret       = flag.String("gh-webhook-secret", "", fmt.Sprintf("GitHub webhook secret used for validing webhook messages; can be taken from %s env var too", ghWebhookSecretEnvName))
	ghReadWriteTeams      = flag.String("gh-rw-teams", "", "comma-separated list of read-write GitHub teams in the org/team format")
	ghReadOnlyTeams       = flag.String("gh-ro-teams", "", "comma-separated list of read-only GitHub teams in the org/team format")
	ghEnterpriseURL       = flag.String("gh-enterprise-url", "", fmt.Sprintf("base URL of the enterprise instance if using GHE; can be taken from %s env var too", ghEnterpriseURLEnvName))
	oidcClientID          = flag.String("oidc-client-id", "", "OIDC client ID used for authentication")
	oidcClientSecret      = flag.String("oidc-client-secret", "", fmt.Sprintf("OIDC client Secret used for authentication; can be taken from %s env var too", oidcClientIDEnvName))
	oidcIssuerURL         = flag.String("oidc-issuer-url", "", fmt.Sprintf("OIDC issuer URL used for authentication;can be taken from %s env var too", oidcClientSecretEnvName))
	oidcValidRedirectURLs = flag.String("oidc-valid-redirect-urls", "http://localhost:8000/*", "OIDC valid Redirect URLs")
	oidcAdminRoles        = flag.String("oidc-admin-roles", "", "comma-separated list of accepted roles with admin access")
	oidcViewerRoles       = flag.String("oidc-viewer-roles", "", "comma-separated list of accepted roles with viewer access")
	oidcRolesPath         = flag.String("oidc-roles-path", "roles", "json path in which the roles array is present in the id token")
	oidcScopes            = flag.String("oidc-scopes", "openid", "comma-separated list of scopes to be used in OIDC")
	oidcSessionAuthKey    = flag.String("oidc-session-secret", "", fmt.Sprintf("Session secret used for authenticating sessions in cookies used for storing OIDC info , will be generated if none is passed; can be taken from %s env var too", oidcSessionAuthKeyEnvName))
	oidcSessionCryptKey   = flag.String("oidc-session-crypt-key", "", fmt.Sprintf("Session key used for encrypting sessions in cookies used for storing OIDC info, will be generated if none is passed; can be taken from %s env var too", oidcSessionCryptKeyEnvName))
	oidcManagementURL     = flag.String("oidc-management-url", "", "OIDC management url for managing the account")
	oidcLogutURL          = flag.String("oidc-logout-url", "", "URL to logout the user from current session")
	flatcarUpdatesURL     = flag.String("sync-update-url", "https://public.update.flatcar-linux.net/v1/update/", "Flatcar update URL to sync from")
	checkFrequencyVal     = flag.String("sync-interval", "1h", "Sync check interval (the minimum depends on the number of channels to sync, e.g., 8m for 8 channels incl. different architectures)")
	appLogoPath           = flag.String("client-logo", "", "Client app logo, should be a path to svg file")
	appTitle              = flag.String("client-title", "", "Client app title")
	appHeaderStyle        = flag.String("client-header-style", "light", "Client app header style, should be either dark or light")
	apiEndpointSuffix     = flag.String("api-endpoint-suffix", "", "Additional suffix for the API endpoint to serve Omaha clients on; use a secret to only serve your clients, e.g., mysecret results in /v1/update/mysecret")
	debug                 = flag.Bool("debug", false, "sets log level to debug")
	logger                = util.NewLogger("nebraska")
)

func main() {
	if err := mainWithError(); err != nil {
		logger.Error().Err(err).Send()
		os.Exit(1)
	}
}

func mainWithError() error {
	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

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
		oidcAuthConfig *auth.OIDCAuthConfig
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
		logger.Warn().Msg("github auth-mode support will be deprecated, oidc auth-mode is recommended")
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
	case "oidc":
		defaultTeam, err := api.GetTeam()
		if err != nil {
			return err
		}

		url, err := url.Parse(*nebraskaURL)
		if err != nil {
			return fmt.Errorf("nebraska-url is invalid, can't generate oidc callback URL, Err: %w", err)
		}

		url.Path = "/login/cb"

		clientID, err := obtainOIDCClientID(*oidcClientID)
		if err != nil {
			return err
		}
		clientSecret, err := obtainOIDCClientSecret(*oidcClientSecret)
		if err != nil {
			return err
		}

		oidcAuthConfig = &auth.OIDCAuthConfig{
			DefaultTeamID:     defaultTeam.ID,
			ClientID:          clientID,
			ClientSecret:      clientSecret,
			IssuerURL:         *oidcIssuerURL,
			CallbackURL:       url.String(),
			ValidRedirectURLs: strings.Split(*oidcValidRedirectURLs, ","),
			ManagementURL:     *oidcManagementURL,
			LogoutURL:         *oidcLogutURL,
			AdminRoles:        strings.Split(*oidcAdminRoles, ","),
			ViewerRoles:       strings.Split(*oidcViewerRoles, ","),
			Scopes:            strings.Split(*oidcScopes, ","),
			SessionAuthKey:    obtainSessionOIDCAuthKey(*oidcSessionAuthKey),
			SessionCryptKey:   obtainSessionOIDCCryptKey(*oidcSessionCryptKey),
			RolesPath:         *oidcRolesPath,
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
		oidcAuthConfig:      oidcAuthConfig,
		flatcarUpdatesURL:   *flatcarUpdatesURL,
		checkFrequency:      checkFrequency,
		syncerPkgsURL:       *syncerPkgsURL,
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

func obtainSessionOIDCAuthKey(potentialKey string) []byte {
	if key := getPotentialOrEnv(potentialKey, oidcSessionAuthKeyEnvName); key != "" {
		return []byte(key)
	}
	return random.Data(32)
}

func obtainSessionOIDCCryptKey(potentialKey string) []byte {
	if key := getPotentialOrEnv(potentialKey, oidcSessionCryptKeyEnvName); key != "" {
		return []byte(key)
	}
	return random.Data(32)
}

func obtainOIDCClientID(potentialID string) (string, error) {
	if id := getPotentialOrEnv(potentialID, oidcClientIDEnvName); id != "" {
		return id, nil
	}
	return "", errors.New("no OIDC client ID")
}

func obtainOIDCClientSecret(potentialSecret string) (string, error) {
	if secret := getPotentialOrEnv(potentialSecret, oidcClientSecretEnvName); secret != "" {
		return secret, nil
	}
	return "", errors.New("no OIDC client secret")
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

	engine.Use(requestid.New())
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
	if *authMode != "oidc" {
		configRouter.Use(ctl.authenticate)
	}
	configRouter.GET("/", ctl.getConfig)

	// Host Flatcar packages payloads
	if *hostFlatcarPackages {
		flatcarPkgsRouter := wrappedEngine.Group("/flatcar", "flatcar")
		flatcarPkgsRouter.Static("/", *flatcarPackagesPath)
	}

	// Serve frontend static content
	staticRouter := wrappedEngine.Group("/", "static")
	if *authMode != "oidc" {
		staticRouter.Use(ctl.authenticate)
	}
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

	// Gin Swagger setup
	swagger.SwaggerInfo.Title = "Swagger API - Nebraska"
	swagger.SwaggerInfo.Description = "Nebraska Swagger Documentation"
	swagger.SwaggerInfo.Version = "1.0"
	swagger.SwaggerInfo.Host = strings.TrimPrefix(strings.TrimPrefix(*nebraskaURL, "https://"), "http://")
	swagger.SwaggerInfo.BasePath = "/api"
	swagger.SwaggerInfo.Schemes = []string{"http", "https"}
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return engine
}

package config

import (
	"io/ioutil"
	"net/url"
	"os"

	"github.com/caarlos0/env"
	"github.com/kinvolk/nebraska/backend/pkg/random"
	"github.com/pkg/errors"
)

type Config struct {
	EnableSyncer          bool   `env:"ENABLE_SYNCER"`
	HostFlatcarPackages   bool   `env:"HOST_FLATCAR_PACKAGES"`
	FlatcarPackagesPath   string `env:"FLATCAR_PACKAGES_PATH"`
	NebraskaURL           string `env:"NEBRASKA_URL" envDefault:"http://localhost:8000"`
	HttpLog               bool   `env:"HTTP_LOG"`
	HttpStaticDir         string `env:"HTTP_STATIC_DIR" envDefault:"../frontend/build"`
	AuthMode              string `env:"AUTH_MODE" envDefault:"noop"`
	GhClientID            string `env:"GITHUB_OAUTH_CLIENT_ID"`
	GhClientSecret        string `env:"GITHUB_OAUTH_CLIENT_SECRET"`
	GhSessionAuthKey      string `env:"GITHUB_SESSION_SECRET"`
	GhSessionCryptKey     string `env:"GITHUB_SESSION_CRYPT_KEY"`
	GhWebhookSecret       string `env:"GITHUB_WEBHOOK_SECRET"`
	GhReadWriteTeams      string `env:"GITHUB_READ_WRITE_TEAMS"`
	GhReadOnlyTeams       string `env:"GITHUB_READ_ONLY_TEAMS"`
	GhEnterpriseURL       string `env:"GITHUB_ENTERPRISE_URL"`
	OidcClientID          string `env:"OIDC_CLIENT_ID"`
	OidcClientSecret      string `env:"OIDC_CLIENT_SECRET"`
	OidcIssuerURL         string `env:"OIDC_ISSUER_URL"`
	OidcValidRedirectURLs string `env:"OIDC_VALID_REDIRECT_URLS" envDefault:"http://localhost:8000/*"`
	OidcAdminRoles        string `env:"OIDC_ADMIN_ROLES"`
	OidcViewerRoles       string `env:"OIDC_VIEWER_ROLES"`
	OidcRolesPath         string `env:"OIDC_ROLES_PATH" envDefault:"roles"`
	OidcScopes            string `env:"OIDC_SCOPES" envDefault:"openid"`
	OidcSessionAuthKey    string `env:"OIDC_SESSION_SECRET"`
	OidcSessionCryptKey   string `env:"OIDC_SESSION_CRYPT_KEY"`
	OidcManagementURL     string `env:"OIDC_MANAGEMENT_URL"`
	OidcLogutURL          string `env:"OIDC_LOGOUT_URL"`
	FlatcarUpdatesURL     string `env:"FLATCAR_UPDATES_URL" envDefault:"https://public.update.flatcar-linux.net/v1/update/"`
	CheckFrequencyVal     string `env:"CHECK_FREQUENCY_VAL" envDefault:"1h"`
	AppLogoPath           string `env:"APP_LOGO_PATH"`
	AppTitle              string `env:"APP_TITLE"`
	AppHeaderStyle        string `env:"APP_HEADER_STYLE"`
	ApiEndpointSuffix     string `env:"API_ENDPOINT_SUFFIX"`
	Debug                 bool   `env:"DEBUG"`
	ServerPort            uint   `env:"PORT" envDefault:"8000"`
}

func (c *Config) Validate() error {
	if c.HostFlatcarPackages {
		if c.FlatcarPackagesPath == "" {
			return errors.New("Invalid Flatcar packages path. Please ensure you provide a valid path using -flatcar-packages-path")
		}

		tmpFile, err := ioutil.TempFile(c.FlatcarPackagesPath, "")
		if err != nil {
			return errors.New("Invalid Flatcar packages path: " + err.Error())
		}
		defer os.Remove(tmpFile.Name())

		if _, err := url.ParseRequestURI(c.NebraskaURL); err != nil {
			return errors.New("invalid Nebraska URL, please ensure the value provided using -nebraska-url is a valid url")
		}
	}
	return nil
}

func Parse() (*Config, error) {

	var config Config

	err := env.Parse(&config)
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing environment variables")
	}

	if config.GhSessionAuthKey == "" {
		config.GhSessionAuthKey = string(random.Data(32))
	}
	if config.GhSessionCryptKey == "" {
		config.GhSessionCryptKey = string(random.Data(32))
	}
	if config.OidcSessionAuthKey == "" {
		config.OidcSessionAuthKey = string(random.Data(32))
	}
	if config.OidcSessionCryptKey == "" {
		config.OidcSessionCryptKey = string(random.Data(32))
	}
	return &config, nil
}

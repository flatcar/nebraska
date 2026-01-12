package config

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/basicflag"

	"github.com/flatcar/nebraska/backend/pkg/random"
)

type Config struct {
	EnableSyncer        bool   `koanf:"enable-syncer"`
	HostFlatcarPackages bool   `koanf:"host-flatcar-packages"`
	FlatcarPackagesPath string `koanf:"flatcar-packages-path"`
	NebraskaURL         string `koanf:"nebraska-url"`
	SyncerPkgsURL       string `koanf:"syncer-packages-url"`
	HTTPLog             bool   `koanf:"http-log"`
	HTTPStaticDir       string `koanf:"http-static-dir"`
	AuthMode            string `koanf:"auth-mode"`
	FlatcarUpdatesURL   string `koanf:"sync-update-url"`
	CheckFrequencyVal   string `koanf:"sync-interval"`
	AppLogoPath         string `koanf:"client-logo"`
	AppTitle            string `koanf:"client-title"`
	AppHeaderStyle      string `koanf:"client-header-style"`
	APIEndpointSuffix   string `koanf:"api-endpoint-suffix"`
	Debug               bool   `koanf:"debug"`
	ServerPort          uint   `koanf:"port"`
	RollbackDBTo        string `koanf:"rollback-db-to"`

	GhClientID        string `koanf:"gh-client-id"`
	GhClientSecret    string `koanf:"gh-client-secret"`
	GhSessionAuthKey  string `koanf:"gh-session-secret"`
	GhSessionCryptKey string `koanf:"gh-session-crypt-key"`
	GhWebhookSecret   string `koanf:"gh-webhook-secret"`
	GhReadWriteTeams  string `koanf:"gh-rw-teams"`
	GhReadOnlyTeams   string `koanf:"gh-ro-teams"`
	GhEnterpriseURL   string `koanf:"gh-enterprise-url"`

	OidcClientID      string `koanf:"oidc-client-id"`
	OidcIssuerURL     string `koanf:"oidc-issuer-url"`
	OidcAdminRoles    string `koanf:"oidc-admin-roles"`
	OidcViewerRoles   string `koanf:"oidc-viewer-roles"`
	OidcRolesPath     string `koanf:"oidc-roles-path"`
	OidcScopes        string `koanf:"oidc-scopes"`
	OidcManagementURL string `koanf:"oidc-management-url"`
	OidcLogoutURL     string `koanf:"oidc-logout-url"`
	OidcAudience      string `koanf:"oidc-audience"`
	OidcUseUserInfo   bool   `koanf:"oidc-use-userinfo"`
}

const (
	oidcClientIDEnvName      = "NEBRASKA_OIDC_CLIENT_ID"
	ghClientIDEnvName        = "NEBRASKA_GITHUB_OAUTH_CLIENT_ID"
	ghClientSecretEnvName    = "NEBRASKA_GITHUB_OAUTH_CLIENT_SECRET"
	ghSessionAuthKeyEnvName  = "NEBRASKA_GITHUB_SESSION_SECRET"
	ghSessionCryptKeyEnvName = "NEBRASKA_GITHUB_SESSION_CRYPT_KEY"
	ghWebhookSecretEnvName   = "NEBRASKA_GITHUB_WEBHOOK_SECRET"
	ghEnterpriseURLEnvName   = "NEBRASKA_GITHUB_ENTERPRISE_URL"
)

func (c *Config) Validate() error {
	if c.HostFlatcarPackages {
		if c.FlatcarPackagesPath == "" {
			return errors.New("invalid Flatcar packages path. Please ensure you provide a valid path using -flatcar-packages-path")
		}

		tmpFile, err := os.CreateTemp(c.FlatcarPackagesPath, "")
		if err != nil {
			return fmt.Errorf("invalid Flatcar packages path: %w", err)
		}
		defer os.Remove(tmpFile.Name())

		if _, err := url.ParseRequestURI(c.NebraskaURL); err != nil {
			return errors.New("invalid Nebraska URL, please ensure the value provided using -nebraska-url is a valid url")
		}
	}

	switch c.AuthMode {
	case "github":
		if c.GhClientID == "" || c.GhClientSecret == "" || c.GhReadOnlyTeams == "" || c.GhReadWriteTeams == "" {
			return errors.New("invalid github configuration")
		}
	case "oidc":
		if c.OidcClientID == "" || c.OidcIssuerURL == "" || c.OidcAdminRoles == "" || c.OidcViewerRoles == "" {
			return errors.New("invalid OIDC configuration")
		}
	}

	return nil
}

func Parse() (*Config, error) {
	var config Config

	f := flag.NewFlagSet("config", flag.ContinueOnError)
	f.String("rollback-db-to", "", "Rollback db migration to the provided id, eg:0003")
	f.Bool("enable-syncer", false, "Enable Flatcar packages syncer")
	f.Bool("host-flatcar-packages", false, "Host Flatcar packages in Nebraska")
	f.String("flatcar-packages-path", "", "Path where Flatcar packages files should be stored")
	f.String("nebraska-url", "http://localhost:8000", "nebraska URL (http://host:port - required when hosting Flatcar packages in nebraska)")
	f.String("syncer-packages-url", "", "use this URL instead of the original one for packages created by the syncer; any {{ARCH}} and {{VERSION}} in the URL will be replaced by the original package's architecture and version, respectively. If this option is not used but the 'host-flatcar-packages' one is, then the URL will be nebraska-url/flatcar/ .")
	f.Bool("http-log", false, "Enable http requests logging")
	f.String("http-static-dir", "../frontend/dist", "Path to frontend static files")
	f.String("auth-mode", "oidc", "authentication mode, available modes: noop, github, oidc")

	f.String("gh-client-id", "", fmt.Sprintf("GitHub client ID used for authentication; can be taken from %s env var too", ghClientIDEnvName))
	f.String("gh-client-secret", "", fmt.Sprintf("GitHub client secret used for authentication; can be taken from %s env var too", ghClientSecretEnvName))
	f.String("gh-session-secret", "", fmt.Sprintf("Session secret used for authenticating sessions in cookies used for storing GitHub info , will be generated if none is passed; can be taken from %s env var too", ghSessionAuthKeyEnvName))
	f.String("gh-session-crypt-key", "", fmt.Sprintf("Session key used for encrypting sessions in cookies used for storing GitHub info, will be generated if none is passed; can be taken from %s env var too", ghSessionCryptKeyEnvName))
	f.String("gh-webhook-secret", "", fmt.Sprintf("GitHub webhook secret used for validing webhook messages; can be taken from %s env var too", ghWebhookSecretEnvName))
	f.String("gh-rw-teams", "", "comma-separated list of read-write GitHub teams in the org/team format")
	f.String("gh-ro-teams", "", "comma-separated list of read-only GitHub teams in the org/team format")
	f.String("gh-enterprise-url", "", fmt.Sprintf("base URL of the enterprise instance if using GHE; can be taken from %s env var too", ghEnterpriseURLEnvName))

	f.String("oidc-client-id", "", fmt.Sprintf("OIDC client ID used for authentication;can be taken from %s env var too", oidcClientIDEnvName))
	f.String("oidc-issuer-url", "", "OIDC issuer URL used for authentication")
	f.String("oidc-admin-roles", "", "comma-separated list of accepted roles with admin access")
	f.String("oidc-viewer-roles", "", "comma-separated list of accepted roles with viewer access")
	f.String("oidc-roles-path", "roles", "json path in which the roles array is present in the id token")
	f.String("oidc-scopes", "openid,profile,email", "comma-separated list of scopes to be used in OIDC")
	f.String("oidc-management-url", "", "OIDC management url for managing the account")
	f.String("oidc-logout-url", "", "OIDC logout URL (optional fallback when end_session_endpoint is not available in discovery)")
	f.String("oidc-audience", "", "OIDC audience parameter for the access token")
	f.Bool("oidc-use-userinfo", false, "Use OIDC UserInfo endpoint for role extraction (for providers that don't include roles in access token)")
	f.String("sync-update-url", "https://public.update.flatcar-linux.net/v1/update/", "Flatcar update URL to sync from")
	f.String("sync-interval", "1h", "Sync check interval (the minimum depends on the number of channels to sync, e.g., 8m for 8 channels incl. different architectures)")
	f.String("client-logo", "", "Client app logo, should be a path to svg file")
	f.String("client-title", "", "Client app title")
	f.String("client-header-style", "light", "Client app header style, should be either dark or light")
	f.String("api-endpoint-suffix", "", "Additional suffix for the API endpoint to serve Omaha clients on; use a secret to only serve your clients, e.g., mysecret results in /v1/update/mysecret")
	f.Bool("debug", false, "sets log level to debug")
	f.Uint("port", 8000, "port to run server")

	k := koanf.New(".")

	// Load from flag if args are provided
	if len(os.Args) == 0 {
		return nil, errors.New("no args provided")
	}

	if err := f.Parse(os.Args[1:]); err != nil {
		return nil, fmt.Errorf("error parsing flags: %w", err)
	}
	if err := k.Load(basicflag.Provider(f, "."), nil); err != nil {
		return nil, fmt.Errorf("error loading config from flags: %w", err)
	}

	if err := k.Unmarshal("", &config); err != nil {
		return nil, fmt.Errorf("error unmarshal config: %w", err)
	}

	switch config.AuthMode {
	case "oidc":
		config.OidcClientID = getPotentialOrEnv(config.OidcClientID, oidcClientIDEnvName)

	case "github":
		config.GhClientID = getPotentialOrEnv(config.GhClientID, ghClientIDEnvName)
		config.GhClientSecret = getPotentialOrEnv(config.GhClientSecret, ghClientSecretEnvName)
		config.GhSessionAuthKey = getPotentialOrEnv(config.GhSessionAuthKey, ghSessionAuthKeyEnvName)
		config.GhSessionCryptKey = getPotentialOrEnv(config.GhSessionCryptKey, ghSessionCryptKeyEnvName)
		config.GhWebhookSecret = getPotentialOrEnv(config.GhWebhookSecret, ghWebhookSecretEnvName)
		config.GhEnterpriseURL = getPotentialOrEnv(config.GhEnterpriseURL, ghEnterpriseURLEnvName)

		if config.GhSessionAuthKey == "" {
			config.GhSessionAuthKey = string(random.Data(32))
		}

		if config.GhSessionCryptKey == "" {
			config.GhSessionCryptKey = string(random.Data(32))
		}
	}

	return &config, nil
}

func getPotentialOrEnv(potentialValue, envName string) string {
	if potentialValue != "" {
		return potentialValue
	}
	return os.Getenv(envName)
}

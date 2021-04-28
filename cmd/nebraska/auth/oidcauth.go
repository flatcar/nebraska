package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/kinvolk/nebraska/cmd/nebraska/ginhelpers"
)

type OIDCAuthConfig struct {
	DefaultTeamID string
	ClientID      string
	IssuerURL     string
	AdminRoles    []string
	ViewerRoles   []string
}

type oidcAuth struct {
	clientID      string
	issuerURL     string
	verifier      *oidc.IDTokenVerifier
	defaultTeamID string
	AdminRoles    []string
	ViewerRoles   []string
}

func NewOIDCAuthenticator(oidcConfig *OIDCAuthConfig) Authenticator {

	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, oidcConfig.IssuerURL)
	if err != nil {
		fmt.Printf("setup error %v\n", err)
		return nil
	}

	fmt.Printf("provider %+v\n", provider)
	oidcProviderConfig := &oidc.Config{
		ClientID:          oidcConfig.ClientID,
		SkipClientIDCheck: true,
	}

	verifier := provider.Verifier(oidcProviderConfig)

	return &oidcAuth{oidcConfig.ClientID, oidcConfig.IssuerURL, verifier, oidcConfig.DefaultTeamID, oidcConfig.AdminRoles, oidcConfig.ViewerRoles}
}

func (oa *oidcAuth) SetupRouter(router ginhelpers.Router) {

}

type claims struct {
	Roles []string `json:"roles"`
}

func (oa *oidcAuth) Authenticate(c *gin.Context) (teamID string, replied bool) {
	ctx := c.Request.Context()

	// Verify Token
	token := c.Request.Header.Get("Authorization")
	fmt.Printf("In verifier\n")
	split := strings.Split(token, " ")
	if len(split) == 2 {
		token = split[1]
	} else {
		redirectTo(c, "http://localhost:3000/")
	}
	tk, err := oa.verifier.Verify(ctx, token)
	if err != nil {
		fmt.Println("Error", err)
		return "", true
	}

	var tokenClaims claims
	// var rawClaim interface{}
	tk.Claims(&tokenClaims)

	accessLevel := ""

	// Check and set access level
checkloop:
	for _, role := range tokenClaims.Roles {

		if accessLevel != "viewer" {
			for _, roRole := range oa.ViewerRoles {
				if roRole == role {
					accessLevel = "viewer"
				}
			}
		}

		for _, rwRole := range oa.AdminRoles {
			if rwRole == role {
				accessLevel = "admin"
				break checkloop
			}
		}
	}

	if accessLevel == "" {
		return "", true
	} else if accessLevel != "admin" {
		if c.Request.Method != "HEAD" && c.Request.Method != "GET" {
			httpError(c, http.StatusForbidden)
			return "", true
		}
	}
	// If access level and method doesn't match return

	// tk.Claims(&rawClaim)
	// fmt.Println(rawClaim)
	fmt.Println("sending", oa.defaultTeamID, "with access level", accessLevel)
	return oa.defaultTeamID, false
}

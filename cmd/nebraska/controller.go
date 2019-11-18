package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v28/github"
	"github.com/ymichael/sessions"
	"github.com/zenazn/goji/web"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"

	"github.com/kinvolk/nebraska/pkg/api"
	"github.com/kinvolk/nebraska/pkg/omaha"
	"github.com/kinvolk/nebraska/pkg/syncer"
)

const (
	clientIDEnvName      = "NEBRASKA_OAUTH_CLIENT_ID"
	clientSecretEnvName  = "NEBRASKA_OAUTH_CLIENT_SECRET"
	sessionSecretEnvName = "NEBRASKA_SESSION_SECRET"
	webhookSecretEnvName = "NEBRASKA_WEBHOOK_SECRET"
)

type ghTeamData struct {
	org  string
	team *string
}

type stringSet map[string]struct{}
type teamToUsersMap map[string]stringSet
type sessionIDToTeamDataMap map[string]ghTeamData
type userSessionMap map[string]sessionIDToTeamDataMap

type controller struct {
	api            *api.API
	omahaHandler   *omaha.Handler
	syncer         *syncer.Syncer
	sessions       *sessions.SessionOptions
	oauthConfig    *oauth2.Config
	userSessionIDs userSessionMap
	teamToUsers    teamToUsersMap
	webhookSecret  string
	readWriteTeams []string
	readOnlyTeams  []string
}

type controllerConfig struct {
	enableSyncer        bool
	hostFlatcarPackages bool
	flatcarPackagesPath string
	nebraskaURL         string
	sessionSecret       string
	oauthClientID       string
	oauthClientSecret   string
	webhookSecret       string
	readWriteTeams      []string
	readOnlyTeams       []string
}

func getPotentialOrEnv(potentialValue, envName string) string {
	if potentialValue != "" {
		return potentialValue
	}
	return os.Getenv(envName)
}

func obtainSessionSecret(potentialSecret string) string {
	if secret := getPotentialOrEnv(potentialSecret, sessionSecretEnvName); secret != "" {
		return secret
	}
	return sessions.GenerateRandomString(64)
}

func obtainOAuthClientID(potentialID string) (string, error) {
	if id := getPotentialOrEnv(potentialID, clientIDEnvName); potentialID != "" {
		return id, nil
	}
	return "", errors.New("no oauth client ID")
}

func obtainOAuthClientSecret(potentialSecret string) (string, error) {
	if secret := getPotentialOrEnv(potentialSecret, clientSecretEnvName); secret != "" {
		return secret, nil
	}
	return "", errors.New("no oauth client secret")
}

func obtainWebhookSecret(potentialSecret string) (string, error) {
	if secret := getPotentialOrEnv(potentialSecret, webhookSecretEnvName); secret != "" {
		return secret, nil
	}
	return "", errors.New("no webhook secret")
}

type lockedStore struct {
	store sessions.Store
	lock  sync.RWMutex
}

var _ sessions.Store = &lockedStore{}

func newLockedStore(store sessions.Store) *lockedStore {
	return &lockedStore{
		store: store,
	}
}

func (s *lockedStore) Get(key string) (map[string]interface{}, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.store.Get(key)
}

func (s *lockedStore) Save(key string, object map[string]interface{}) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.store.Save(key, object)
}

func (s *lockedStore) Destroy(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.store.Destroy(key)
}

func newController(conf *controllerConfig) (*controller, error) {
	api, err := api.New()
	if err != nil {
		return nil, err
	}

	sessionSecret := obtainSessionSecret(conf.sessionSecret)
	clientID, err := obtainOAuthClientID(conf.oauthClientID)
	if err != nil {
		return nil, err
	}
	clientSecret, err := obtainOAuthClientSecret(conf.oauthClientSecret)
	if err != nil {
		return nil, err
	}
	webhookSecret, err := obtainWebhookSecret(conf.webhookSecret)
	if err != nil {
		return nil, err
	}

	c := &controller{
		api:          api,
		omahaHandler: omaha.NewHandler(api),
		sessions:     sessions.NewSessionOptions(sessionSecret, newLockedStore(sessions.MemoryStore{})),
		oauthConfig: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			// We are using following APIs:
			//
			// https://developer.github.com/v3/teams/#list-user-teams
			//
			// https://developer.github.com/v3/orgs/#list-your-organizations
			//
			// https://developer.github.com/v3/users/#get-the-authenticated-user
			//
			// Common required scope in those APIs seems
			// to be "user". Listing teams and orgs can be
			// done also with "read:org" scope. We don't
			// need "user" scope really as all we need is
			// just login and that's public information
			// accessible without any scope at all.
			Scopes:   []string{"read:org"},
			Endpoint: githuboauth.Endpoint,
		},
		userSessionIDs: make(userSessionMap),
		teamToUsers:    make(teamToUsersMap),
		webhookSecret:  webhookSecret,
		readWriteTeams: conf.readWriteTeams,
		readOnlyTeams:  conf.readOnlyTeams,
	}

	if conf.enableSyncer {
		syncerConf := &syncer.Config{
			API:          api,
			HostPackages: conf.hostFlatcarPackages,
			PackagesPath: conf.flatcarPackagesPath,
			PackagesURL:  conf.nebraskaURL + flatcarPkgsRouterPrefix,
		}
		syncer, err := syncer.New(syncerConf)
		if err != nil {
			return nil, err
		}
		c.syncer = syncer
		go syncer.Start()
	}

	return c, nil
}

func (ctl *controller) close() {
	if ctl.syncer != nil {
		ctl.syncer.Stop()
	}
	ctl.api.Close()
}

// ----------------------------------------------------------------------------
// OAuth
//

func redirectTo(w http.ResponseWriter, r *http.Request, where string) {
	http.Redirect(w, r, where, http.StatusTemporaryRedirect)
}

func makeTeamName(org, team string) string {
	return fmt.Sprintf("%s/%s", org, team)
}

func (ctl *controller) cleanupSession(c *web.C, w http.ResponseWriter) {
	defer ctl.sessions.DestroySession(c, w)
	obj := ctl.sessions.GetSessionObject(c)
	usernameAny, ok := obj["username"]
	if !ok {
		return
	}
	username, ok := usernameAny.(string)
	if !ok {
		return
	}
	sessionIDs, ok := ctl.userSessionIDs[username]
	if !ok {
		return
	}
	sessionID := ctl.sessions.GetSessionId(c)
	if teamData, ok := sessionIDs[sessionID]; ok {
		if teamData.team != nil {
			teamName := makeTeamName(teamData.org, *teamData.team)
			if usersSet, ok := ctl.teamToUsers[teamName]; ok {
				delete(usersSet, username)
				if len(usersSet) == 0 {
					delete(ctl.teamToUsers, teamName)
				}
			}
		}
		delete(sessionIDs, sessionID)
		if len(sessionIDs) == 0 {
			delete(ctl.userSessionIDs, username)
		}
	}
}

func httpError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (ctl *controller) loginCb(c web.C, w http.ResponseWriter, r *http.Request) {
	const (
		resultOK = iota
		resultUnauthorized
		resultInternalFailure
	)
	result := resultInternalFailure
	defer func() {
		switch result {
		case resultOK:
		case resultUnauthorized:
			ctl.cleanupSession(&c, w)
			httpError(w, http.StatusUnauthorized)
		case resultInternalFailure:
			ctl.cleanupSession(&c, w)
			httpError(w, http.StatusInternalServerError)
		}
	}()

	obj := ctl.sessions.GetSessionObject(&c)
	desiredURLAny, ok := obj["desiredurl"]
	if !ok {
		logger.Error("login cb", "expected to have desiredurl item in session data")
		return
	}
	desiredURL, ok := desiredURLAny.(string)
	if !ok {
		logger.Error("login cb", "expected the desiredurl item in session data to be a string, but it was something else", fmt.Sprintf("%T", desiredURLAny))
		return
	}
	state := r.FormValue("state")
	logger.Debug("login cb", "received oauth state", state)
	expectedStateAny, ok := obj["state"]
	if !ok {
		logger.Error("login cb", "expected to have state item in session data")
		return
	}
	expectedState, ok := expectedStateAny.(string)
	if !ok {
		logger.Error("login cb", "expected the expectedstate item in session data to be a string, but it was something else", fmt.Sprintf("%T", desiredURLAny))
		return
	}

	if expectedState != state {
		logger.Error("login cb: invalid oauth state", "expected", expectedState, "got", state)
		return
	}
	code := r.FormValue("code")
	logger.Debug("login cb", "received code", code)
	ctx := context.Background()
	token, err := ctl.oauthConfig.Exchange(ctx, code)
	if err != nil {
		logger.Error("login cb: oauth exchange failed", "error", err)
		return
	}
	logger.Debug("login cb", "received token", token)
	if !token.Valid() {
		logger.Error("login cb", "got invalid token")
		return
	}

	oauthClient := ctl.oauthConfig.Client(ctx, token)
	result = resultOK
	if replied := ctl.doLoginDance(ctx, oauthClient, &c, w); !replied {
		redirectTo(w, r, desiredURL)
	}
}

type GhAppAuthPayload struct {
	Action string `json:"action"`
	Sender GhUser `json:"sender"`
}

type GhUser struct {
	Login string `json:"login"`
}

func (ctl *controller) loginWebhookAuthorizationEvent(c web.C, w http.ResponseWriter, payloadReader io.Reader) {
	var payload GhAppAuthPayload
	if err := json.NewDecoder(payloadReader).Decode(&payload); err != nil {
		logger.Error("webhook", "error unmarshalling github_app_authorization payload", err.Error())
		httpError(w, http.StatusBadRequest)
		return
	}
	logger.Debug("webhook", "got github_app_authorization event with action", payload.Action)
	if payload.Action != "revoked" {
		logger.Debug("webhook", "ignoring github_app_authorization event with action", payload.Action)
		return
	}
	username := payload.Sender.Login
	logger.Debug("webhook", "dropping all the sessions of user", username)
	if sessionIDs, ok := ctl.userSessionIDs[username]; ok {
		for sessionID := range sessionIDs {
			logger.Debug("webhook", "dropping session", sessionID)
			ctl.sessions.Store.Destroy(sessionID)
		}
		delete(ctl.userSessionIDs, username)
	}
}

type GhMembership struct {
	User GhUser `json:"user"`
}

type GhOrganizationPayload struct {
	Action     string       `json:"action"`
	Membership GhMembership `json:"membership"`
	Org        GhUser       `json:"organization"`
}

func (ctl *controller) loginWebhookOrganizationEvent(c web.C, w http.ResponseWriter, payloadReader io.Reader) {
	var payload GhOrganizationPayload
	if err := json.NewDecoder(payloadReader).Decode(&payload); err != nil {
		logger.Error("webhook", "error unmarshalling organization payload", err.Error())
		httpError(w, http.StatusBadRequest)
		return
	}
	logger.Debug("webhook", "got organization event with action", payload.Action)
	if payload.Action != "member_removed" {
		logger.Debug("webhook", "ignoring organization event with action", payload.Action)
		return
	}
	username := payload.Membership.User.Login
	org := payload.Org.Login
	if sessionIDs, ok := ctl.userSessionIDs[username]; ok {
		for sessionID, teamData := range sessionIDs {
			if teamData.org == org && teamData.team == nil {
				logger.Debug("webhook", "dropping session", sessionID)
				ctl.sessions.Store.Destroy(sessionID)
				delete(sessionIDs, sessionID)
			}
		}
		if len(sessionIDs) == 0 {
			logger.Debug("webhook", "dropping all the sessions of user", username)
			delete(ctl.userSessionIDs, username)
		}
	}
}

type GhTeam struct {
	Name string `json:"name"`
}

type GhMembershipPayload struct {
	Action string `json:"action"`
	Scope  string `json:"scope"`
	Member GhUser `json:"member"`
	Team   GhTeam `json:"team"`
	Org    GhUser `json:"organization"`
}

func (ctl *controller) loginWebhookMembershipEvent(c web.C, w http.ResponseWriter, payloadReader io.Reader) {
	var payload GhMembershipPayload
	if err := json.NewDecoder(payloadReader).Decode(&payload); err != nil {
		logger.Error("webhook", "error unmarshalling membership payload", err.Error())
		httpError(w, http.StatusBadRequest)
		return
	}
	logger.Debug("webhook", "got membership event with action", payload.Action)
	if payload.Action != "removed" {
		logger.Debug("webhook", "ignoring membership event with action", payload.Action)
		return
	}
	logger.Debug("webhook", "got membership remove event with scope", payload.Scope)
	if payload.Scope != "team" {
		logger.Debug("webhook", "ignoring membership remove event with scope", payload.Scope)
		return
	}
	username := payload.Member.Login
	org := payload.Org.Login
	team := payload.Team.Name
	if sessionIDs, ok := ctl.userSessionIDs[username]; ok {
		for sessionID, teamData := range sessionIDs {
			if teamData.org == org && teamData.team != nil && *teamData.team == team {
				logger.Debug("webhook", "dropping session", sessionID, "user", username)
				ctl.sessions.Store.Destroy(sessionID)
				delete(sessionIDs, sessionID)
			}
		}
		if len(sessionIDs) == 0 {
			logger.Debug("webhook", "dropping all the sessions of user", username)
			delete(ctl.userSessionIDs, username)
		}
	}
}

type GhChangesName struct {
	From string `json:"from"`
}

type GhChanges struct {
	Name GhChangesName `json:"name"`
}

type GhTeamPayload struct {
	Action  string    `json:"action"`
	Changes GhChanges `json:"changes"`
	Team    GhTeam    `json:"team"`
	Org     GhUser    `json:"organization"`
}

func (ctl *controller) loginWebhookTeamEvent(c web.C, w http.ResponseWriter, payloadReader io.Reader) {
	var payload GhTeamPayload
	if err := json.NewDecoder(payloadReader).Decode(&payload); err != nil {
		logger.Error("webhook", "error unmarshalling team payload", err.Error())
		httpError(w, http.StatusBadRequest)
		return
	}
	logger.Debug("webhook", "got team event with action", payload.Action)
	org := payload.Org.Login
	team := ""
	switch payload.Action {
	case "deleted":
		team = payload.Team.Name
	case "edited":
		if payload.Changes.Name.From == "" {
			logger.Debug("ignoring edited team event that does not rename the team")
			return
		}
		team = payload.Changes.Name.From
	default:
		logger.Debug("webhook", "ignoring team event with action", payload.Action)
		return
	}
	teamName := makeTeamName(org, team)
	for username := range ctl.teamToUsers[teamName] {
		if sessionIDs, ok := ctl.userSessionIDs[username]; ok {
			for sessionID, teamData := range sessionIDs {
				if teamData.org == org && teamData.team != nil && *teamData.team == team {
					logger.Debug("webhook", "dropping session", sessionID, "user", username)
					ctl.sessions.Store.Destroy(sessionID)
					delete(sessionIDs, sessionID)
				}
			}
			if len(sessionIDs) == 0 {
				logger.Debug("webhook", "dropping all the sessions of user", username)
				delete(ctl.userSessionIDs, username)
			}
		}
	}
	delete(ctl.teamToUsers, teamName)
}

func (ctl *controller) loginWebhook(c web.C, w http.ResponseWriter, r *http.Request) {
	signature := r.Header.Get("X-Hub-Signature")
	if len(signature) == 0 {
		logger.Debug("webhook", "request with missing signature, ignoring it")
		return
	}
	eventType := r.Header.Get("X-Github-Event")
	rawPayload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Debug("webhook", "failed to read the contents of the message", eventType)
		return
	}
	mac := hmac.New(sha1.New, []byte(ctl.webhookSecret))
	_, _ = mac.Write(rawPayload)
	payloadMAC := hex.EncodeToString(mac.Sum(nil))
	// [5:] is to drop the "sha1-" part.
	if !hmac.Equal([]byte(signature[5:]), []byte(payloadMAC)) {
		logger.Debug("webhook", "message validation failed")
		return
	}
	payloadReader := bytes.NewBuffer(rawPayload)
	logger.Debug("webhook", "got event of type", eventType)
	switch eventType {
	case "github_app_authorization":
		ctl.loginWebhookAuthorizationEvent(c, w, payloadReader)
	case "organization":
		ctl.loginWebhookOrganizationEvent(c, w, payloadReader)
	case "membership":
		ctl.loginWebhookMembershipEvent(c, w, payloadReader)
	case "team":
		ctl.loginWebhookTeamEvent(c, w, payloadReader)
	default:
		logger.Debug("webhook", "ignoring event", eventType)
		return
	}
}

func (ctl *controller) doLoginDance(ctx context.Context, oauthClient *http.Client, c *web.C, w http.ResponseWriter) (replied bool) {
	const (
		resultOK = iota
		resultUnauthorized
		resultInternalFailure
	)

	result := resultUnauthorized
	obj := ctl.sessions.GetSessionObject(c)
	defer func() {
		replied = true
		switch result {
		case resultOK:
			replied = false
		case resultUnauthorized:
			ctl.cleanupSession(c, w)
			httpError(w, http.StatusUnauthorized)
		case resultInternalFailure:
			httpError(w, http.StatusInternalServerError)
		default:
			httpError(w, http.StatusInternalServerError)
		}
	}()

	client := github.NewClient(oauthClient)
	ghUser, _, err := client.Users.Get(ctx, "")
	if err != nil {
		logger.Error("login dance", "failed to get authenticated user", err)
		result = resultInternalFailure
		return
	}
	if ghUser.Login == nil {
		logger.Error("login dance", "authenticated as a user without a login, meh")
		return
	}

	rwTeams := ctl.readWriteTeams
	roTeams := ctl.readOnlyTeams

	defaultTeam, err := ctl.api.GetTeam()
	if err != nil {
		logger.Error("login dance", "failed to get default team", err)
		result = resultInternalFailure
		return
	}

	teamData := ghTeamData{}
	teamID := ""
	listOpts := github.ListOptions{
		Page:    1,
		PerPage: 50,
	}
	for {
		ghTeams, response, err := client.Teams.ListUserTeams(ctx, &listOpts)
		if err != nil {
			logger.Error("login dance", "failed to get user teams", err)
			result = resultInternalFailure
			return
		}
		for _, ghTeam := range ghTeams {
			if ghTeam.Name == nil {
				logger.Debug("login dance", "unnamed github team")
				continue
			}
			logger.Debug("login dance", "github team", *ghTeam.Name)
			if ghTeam.Organization == nil {
				logger.Debug("login dance", "github team with no org")
				continue
			}
			if ghTeam.Organization.Login == nil {
				logger.Debug("login dance", "github team in unnamed organization")
				continue
			}
			logger.Debug("login dance", "github team in organization", *ghTeam.Organization.Login)
			fullGithubTeamName := makeTeamName(*ghTeam.Organization.Login, *ghTeam.Name)
			logger.Debug("login dance", "trying to find a matching ro or rw team", fullGithubTeamName)
			for _, roTeam := range roTeams {
				if fullGithubTeamName == roTeam {
					logger.Debug("login dance", "found matching ro team", fullGithubTeamName)
					teamData.org = *ghTeam.Organization.Login
					teamData.team = ghTeam.Name
					teamID = defaultTeam.ID
					obj["accesslevel"] = "ro"
					break
				}
			}
			for _, rwTeam := range rwTeams {
				if fullGithubTeamName == rwTeam {
					logger.Debug("login dance", "found matching rw team", fullGithubTeamName)
					teamData.org = *ghTeam.Organization.Login
					teamData.team = ghTeam.Name
					teamID = defaultTeam.ID
					obj["accesslevel"] = "rw"
					break
				}
			}
		}
		if teamID != "" {
			break
		}
		// Next page being zero means that we are on the last
		// page.
		if response.NextPage == 0 {
			break
		}
		listOpts.Page = response.NextPage
	}
	if teamID == "" {
		logger.Debug("login dance", "no matching teams found, trying orgs")
		listOpts.Page = 1
		for {
			ghOrgs, response, err := client.Organizations.List(ctx, "", &listOpts)
			if err != nil {
				logger.Error("login dance", "failed to get user orgs", err)
				result = resultInternalFailure
				return
			}
			for _, ghOrg := range ghOrgs {
				if ghOrg.Login == nil {
					logger.Debug("login dance", "unnamed github organization")
					continue
				}
				logger.Debug("login dance", "github org", *ghOrg.Login)
				logger.Debug("login dance", "trying to find a matching ro or rw team", *ghOrg.Login)
				nebraskaOrgName := *ghOrg.Login
				for _, roTeam := range roTeams {
					if nebraskaOrgName == roTeam {
						logger.Debug("login dance", "found matching ro team", nebraskaOrgName)
						teamData.org = nebraskaOrgName
						teamID = defaultTeam.ID
						obj["accesslevel"] = "ro"
						break
					}
				}
				for _, rwTeam := range rwTeams {
					if nebraskaOrgName == rwTeam {
						logger.Debug("login dance", "found matching rw team", nebraskaOrgName)
						teamData.org = nebraskaOrgName
						teamID = defaultTeam.ID
						obj["accesslevel"] = "rw"
						break
					}
				}
			}
			if teamID != "" {
				break
			}
			// Next page being zero means that we are on the last
			// page.
			if response.NextPage == 0 {
				break
			}
			listOpts.Page = response.NextPage
		}
	}
	if teamID == "" {
		logger.Debug("login dance", "not authorized")
		return
	}
	username := *ghUser.Login
	obj["teamID"] = teamID
	obj["username"] = username
	sessionIDs := ctl.userSessionIDs[username]
	if sessionIDs == nil {
		sessionIDs = make(sessionIDToTeamDataMap)
		ctl.userSessionIDs[username] = sessionIDs
	}
	sessionIDs[ctl.sessions.GetSessionId(c)] = teamData
	if teamData.team != nil {
		teamName := makeTeamName(teamData.org, *teamData.team)
		users := ctl.teamToUsers[teamName]
		if users == nil {
			users = make(stringSet)
			ctl.teamToUsers[teamName] = users
		}
		users[username] = struct{}{}
	}
	result = resultOK
	return
}

// ----------------------------------------------------------------------------
// Authentication
//

func (ctl *controller) authMissingTeamID(c *web.C, w http.ResponseWriter, r *http.Request, obj map[string]interface{}) (teamID string, replied bool) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		oauthState := sessions.GenerateRandomString(64)
		obj["state"] = oauthState
		obj["desiredurl"] = r.URL.String()
		logger.Debug("authenticate", "oauthstate", oauthState)
		url := ctl.oauthConfig.AuthCodeURL(oauthState, oauth2.AccessTypeOnline)
		logger.Debug("authenticate", "redirecting to", url)
		redirectTo(w, r, url)
		replied = true
	} else {
		failed := true
		defer func() {
			if failed {
				replied = true
				ctl.cleanupSession(c, w)
				httpError(w, http.StatusUnauthorized)
			}
		}()
		if authHeader == "" {
			logger.Debug("auth metrics", "no authorization header in headers", r.Header)
			return "", true
		}
		splitToken := strings.Fields(authHeader)
		if len(splitToken) != 2 {
			logger.Debug("auth metrics", "malformed authorization header", authHeader)
			return
		}
		if strings.ToLower(strings.TrimSpace(splitToken[0])) != "bearer" {
			logger.Debug("auth metrics", "authorization is not a bearer token", authHeader)
			return
		}
		bearerToken := strings.TrimSpace(splitToken[1])
		logger.Debug("auth metrics", "going to do the login dance with token", bearerToken)
		token := oauth2.Token{
			AccessToken: bearerToken,
		}
		tokenSource := oauth2.StaticTokenSource(&token)
		ctx := context.Background()
		oauthClient := oauth2.NewClient(ctx, tokenSource)
		failed = false
		if replied = ctl.doLoginDance(ctx, oauthClient, c, w); !replied {
			teamID = obj["teamID"].(string)
		}
	}
	return
}

// authenticate is a middleware handler in charge of authenticating requests.
func (ctl *controller) authenticate(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		obj := ctl.sessions.GetSessionObject(c)
		teamID, ok := obj["teamID"]
		if !ok {
			newTeamID, replied := ctl.authMissingTeamID(c, w, r, obj)
			if replied {
				return
			}
			teamID = newTeamID
		}

		c.Env["team_id"] = teamID
		if obj["accesslevel"] == "ro" {
			if r.Method != "HEAD" && r.Method != "GET" {
				httpError(w, http.StatusForbidden)
				return
			}
		}
		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// ----------------------------------------------------------------------------
// API: users
//

func (ctl *controller) updateUserPassword(c web.C, w http.ResponseWriter, r *http.Request) {
}

// ----------------------------------------------------------------------------
// API: applications CRUD
//

func (ctl *controller) addApp(c web.C, w http.ResponseWriter, r *http.Request) {
	sourceAppID := r.URL.Query().Get("clone_from")

	app := &api.Application{}
	if err := json.NewDecoder(r.Body).Decode(app); err != nil {
		logger.Error("addApp - decoding payload", "error", err.Error())
		httpError(w, http.StatusBadRequest)
		return
	}
	app.TeamID = c.Env["team_id"].(string)

	_, err := ctl.api.AddAppCloning(app, sourceAppID)
	if err != nil {
		logger.Error("addApp - cloning app", "error", err.Error(), "app", app, "sourceAppID", sourceAppID)
		httpError(w, http.StatusBadRequest)
		return
	}

	app, err = ctl.api.GetApp(app.ID)
	if err != nil {
		logger.Error("addApp - getting added app", "error", err.Error(), "appID", app.ID)
		httpError(w, http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(app); err != nil {
		logger.Error("addApp - encoding app", "error", err.Error(), "app", app)
	}
}

func (ctl *controller) updateApp(c web.C, w http.ResponseWriter, r *http.Request) {
	app := &api.Application{}
	if err := json.NewDecoder(r.Body).Decode(app); err != nil {
		logger.Error("updateApp - decoding payload", "error", err.Error())
		httpError(w, http.StatusBadRequest)
		return
	}
	app.ID = c.URLParams["app_id"]
	app.TeamID = c.Env["team_id"].(string)

	err := ctl.api.UpdateApp(app)
	if err != nil {
		logger.Error("updatedApp - updating app", "error", err.Error(), "app", app)
		httpError(w, http.StatusBadRequest)
		return
	}

	app, err = ctl.api.GetApp(app.ID)
	if err != nil {
		logger.Error("updateApp - getting updated app", "error", err.Error(), "appID", app.ID)
		httpError(w, http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(app); err != nil {
		logger.Error("updateApp - encoding app", "error", err.Error(), "appID", app.ID)
	}
}

func (ctl *controller) deleteApp(c web.C, w http.ResponseWriter, r *http.Request) {
	appID := c.URLParams["app_id"]

	err := ctl.api.DeleteApp(appID)
	switch err {
	case nil:
		httpError(w, http.StatusNoContent)
	default:
		logger.Error("deleteApp", "error", err.Error(), "appID", appID)
		httpError(w, http.StatusBadRequest)
	}
}

func (ctl *controller) getApp(c web.C, w http.ResponseWriter, r *http.Request) {
	appID := c.URLParams["app_id"]

	app, err := ctl.api.GetApp(appID)
	switch err {
	case nil:
		if err := json.NewEncoder(w).Encode(app); err != nil {
			logger.Error("getApp - encoding app", "error", err.Error(), "appID", appID)
		}
	case sql.ErrNoRows:
		httpError(w, http.StatusNotFound)
	default:
		logger.Error("getApp - getting app", "error", err.Error(), "appID", appID)
		httpError(w, http.StatusBadRequest)
	}
}

func (ctl *controller) getApps(c web.C, w http.ResponseWriter, r *http.Request) {
	teamID, _ := c.Env["team_id"].(string)
	page, _ := strconv.ParseUint(r.URL.Query().Get("page"), 10, 64)
	perPage, _ := strconv.ParseUint(r.URL.Query().Get("perpage"), 10, 64)

	apps, err := ctl.api.GetApps(teamID, page, perPage)
	switch err {
	case nil:
		if err := json.NewEncoder(w).Encode(apps); err != nil {
			logger.Error("getApps - encoding apps", "error", err.Error(), "teamID", teamID)
		}
	case sql.ErrNoRows:
		httpError(w, http.StatusNotFound)
	default:
		logger.Error("getApps - getting apps", "error", err.Error(), "teamID", teamID)
		httpError(w, http.StatusBadRequest)
	}
}

// ----------------------------------------------------------------------------
// API: groups CRUD
//

func (ctl *controller) addGroup(c web.C, w http.ResponseWriter, r *http.Request) {
	group := &api.Group{}
	if err := json.NewDecoder(r.Body).Decode(group); err != nil {
		logger.Error("addGroup - decoding payload", "error", err.Error())
		httpError(w, http.StatusBadRequest)
		return
	}
	group.ApplicationID = c.URLParams["app_id"]

	_, err := ctl.api.AddGroup(group)
	if err != nil {
		logger.Error("addGroup - adding group", "error", err.Error(), "group", group)
		httpError(w, http.StatusBadRequest)
		return
	}

	group, err = ctl.api.GetGroup(group.ID)
	if err != nil {
		logger.Error("addGroup - getting added group", "error", err.Error(), "groupID", group.ID)
		httpError(w, http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(group); err != nil {
		logger.Error("addGroup - encoding group", "error", err.Error(), "group", group)
	}
}

func (ctl *controller) updateGroup(c web.C, w http.ResponseWriter, r *http.Request) {
	group := &api.Group{}
	if err := json.NewDecoder(r.Body).Decode(group); err != nil {
		logger.Error("updateGroup - decoding payload", "error", err.Error())
		httpError(w, http.StatusBadRequest)
		return
	}
	group.ID = c.URLParams["group_id"]
	group.ApplicationID = c.URLParams["app_id"]

	err := ctl.api.UpdateGroup(group)
	if err != nil {
		logger.Error("updateGroup - updating group", "error", err.Error(), "group", group)
		httpError(w, http.StatusBadRequest)
		return
	}

	group, err = ctl.api.GetGroup(group.ID)
	if err != nil {
		logger.Error("updateGroup - fetching updated group", "error", err.Error(), "groupID", group.ID)
		httpError(w, http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(group); err != nil {
		logger.Error("updateGroup - encoding group", "error", err.Error(), "group", group)
	}
}

func (ctl *controller) deleteGroup(c web.C, w http.ResponseWriter, r *http.Request) {
	groupID := c.URLParams["group_id"]

	err := ctl.api.DeleteGroup(groupID)
	switch err {
	case nil:
		httpError(w, http.StatusNoContent)
	default:
		logger.Error("deleteGroup", "error", err.Error(), "groupID", groupID)
		httpError(w, http.StatusBadRequest)
	}
}

func (ctl *controller) getGroup(c web.C, w http.ResponseWriter, r *http.Request) {
	groupID := c.URLParams["group_id"]

	group, err := ctl.api.GetGroup(groupID)
	switch err {
	case nil:
		if err := json.NewEncoder(w).Encode(group); err != nil {
			logger.Error("getGroup - encoding group", "error", err.Error(), "group", group)
		}
	case sql.ErrNoRows:
		httpError(w, http.StatusNotFound)
	default:
		logger.Error("getGroup - getting group", "error", err.Error(), "groupID", groupID)
		httpError(w, http.StatusBadRequest)
	}
}

func (ctl *controller) getGroups(c web.C, w http.ResponseWriter, r *http.Request) {
	appID := c.URLParams["app_id"]
	page, _ := strconv.ParseUint(r.URL.Query().Get("page"), 10, 64)
	perPage, _ := strconv.ParseUint(r.URL.Query().Get("perpage"), 10, 64)

	groups, err := ctl.api.GetGroups(appID, page, perPage)
	switch err {
	case nil:
		if err := json.NewEncoder(w).Encode(groups); err != nil {
			logger.Error("getGroups - encoding groups", "error", err.Error(), "appID", appID)
		}
	case sql.ErrNoRows:
		httpError(w, http.StatusNotFound)
	default:
		logger.Error("getGroups - getting groups", "error", err.Error(), "appID", appID)
		httpError(w, http.StatusBadRequest)
	}
}

func (ctl *controller) getGroupVersionCountTimeline(c web.C, w http.ResponseWriter, r *http.Request) {
	groupID := c.URLParams["group_id"]

	versionCountTimeline, err := ctl.api.GetGroupVersionCountTimeline(groupID)
	switch err {
	case nil:
		if err := json.NewEncoder(w).Encode(versionCountTimeline); err != nil {
			logger.Error("getGroupVersionCountTimeline - encoding group", "error", err.Error(), "count-timeline", versionCountTimeline)
		}
	case sql.ErrNoRows:
		httpError(w, http.StatusNotFound)
	default:
		logger.Error("getGroupVersionCountTimeline - getting version timeline", "error", err.Error(), "groupID", groupID)
		httpError(w, http.StatusBadRequest)
	}
}

func (ctl *controller) getGroupStatusCountTimeline(c web.C, w http.ResponseWriter, r *http.Request) {
	groupID := c.URLParams["group_id"]

	statusCountTimeline, err := ctl.api.GetGroupStatusCountTimeline(groupID)
	switch err {
	case nil:
		if err := json.NewEncoder(w).Encode(statusCountTimeline); err != nil {
			logger.Error("getGroupStatusCountTimeline - encoding group", "error", err.Error(), "count-timeline", statusCountTimeline)
		}
	case sql.ErrNoRows:
		httpError(w, http.StatusNotFound)
	default:
		logger.Error("getGroupStatusCountTimeline - getting status timeline", "error", err.Error(), "groupID", groupID)
		httpError(w, http.StatusBadRequest)
	}
}

// ----------------------------------------------------------------------------
// API: channels CRUD
//

func (ctl *controller) addChannel(c web.C, w http.ResponseWriter, r *http.Request) {
	channel := &api.Channel{}
	if err := json.NewDecoder(r.Body).Decode(channel); err != nil {
		logger.Error("addChannel", "error", err.Error())
		httpError(w, http.StatusBadRequest)
		return
	}
	channel.ApplicationID = c.URLParams["app_id"]

	_, err := ctl.api.AddChannel(channel)
	if err != nil {
		logger.Error("addChannel", "error", err.Error(), "channel", channel)
		httpError(w, http.StatusBadRequest)
		return
	}

	channel, err = ctl.api.GetChannel(channel.ID)
	if err != nil {
		logger.Error("addChannel", "error", err.Error(), "channelID", channel.ID)
		httpError(w, http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(channel); err != nil {
		logger.Error("addChannel - encoding channel", "error", err.Error(), "channelID", channel.ID)
	}
}

func (ctl *controller) updateChannel(c web.C, w http.ResponseWriter, r *http.Request) {
	channel := &api.Channel{}
	if err := json.NewDecoder(r.Body).Decode(channel); err != nil {
		logger.Error("updateChannel - decoding payload", "error", err.Error())
		httpError(w, http.StatusBadRequest)
		return
	}
	channel.ID = c.URLParams["channel_id"]
	channel.ApplicationID = c.URLParams["app_id"]

	err := ctl.api.UpdateChannel(channel)
	if err != nil {
		logger.Error("updateChannel - updating channel", "error", err.Error(), "channel", channel)
		httpError(w, http.StatusBadRequest)
		return
	}

	channel, err = ctl.api.GetChannel(channel.ID)
	if err != nil {
		logger.Error("updateChannel - getting channel updated", "error", err.Error(), "channelID", channel.ID)
		httpError(w, http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(channel); err != nil {
		logger.Error("updateChannel - encoding channel", "error", err.Error(), "channelID", channel.ID)
	}
}

func (ctl *controller) deleteChannel(c web.C, w http.ResponseWriter, r *http.Request) {
	channelID := c.URLParams["channel_id"]

	err := ctl.api.DeleteChannel(channelID)
	switch err {
	case nil:
		httpError(w, http.StatusNoContent)
	default:
		logger.Error("deleteChannel", "error", err.Error(), "channelID", channelID)
		httpError(w, http.StatusBadRequest)
	}
}

func (ctl *controller) getChannel(c web.C, w http.ResponseWriter, r *http.Request) {
	channelID := c.URLParams["channel_id"]

	channel, err := ctl.api.GetChannel(channelID)
	switch err {
	case nil:
		if err := json.NewEncoder(w).Encode(channel); err != nil {
			logger.Error("getChannel - encoding channel", "error", err.Error(), "channelID", channelID)
		}
	case sql.ErrNoRows:
		httpError(w, http.StatusNotFound)
	default:
		logger.Error("getChannel - getting updated channel", "error", err.Error(), "channelID", channelID)
		httpError(w, http.StatusBadRequest)
	}
}

func (ctl *controller) getChannels(c web.C, w http.ResponseWriter, r *http.Request) {
	appID := c.URLParams["app_id"]
	page, _ := strconv.ParseUint(r.URL.Query().Get("page"), 10, 64)
	perPage, _ := strconv.ParseUint(r.URL.Query().Get("perpage"), 10, 64)

	channels, err := ctl.api.GetChannels(appID, page, perPage)
	switch err {
	case nil:
		if err := json.NewEncoder(w).Encode(channels); err != nil {
			logger.Error("getChannels - encoding channel", "error", err.Error(), "appID", appID)
		}
	case sql.ErrNoRows:
		httpError(w, http.StatusNotFound)
	default:
		logger.Error("getChannels - getting channels", "error", err.Error(), "appID", appID)
		httpError(w, http.StatusBadRequest)
	}
}

// ----------------------------------------------------------------------------
// API: packages CRUD
//

func (ctl *controller) addPackage(c web.C, w http.ResponseWriter, r *http.Request) {
	pkg := &api.Package{}
	if err := json.NewDecoder(r.Body).Decode(pkg); err != nil {
		logger.Error("addPackage - decoding payload", "error", err.Error())
		httpError(w, http.StatusBadRequest)
		return
	}
	pkg.ApplicationID = c.URLParams["app_id"]

	_, err := ctl.api.AddPackage(pkg)
	if err != nil {
		logger.Error("addPackage - adding package", "error", err.Error(), "package", pkg)
		httpError(w, http.StatusBadRequest)
		return
	}

	pkg, err = ctl.api.GetPackage(pkg.ID)
	if err != nil {
		logger.Error("addPackage - getting added package", "error", err.Error(), "packageID", pkg.ID)
		httpError(w, http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(pkg); err != nil {
		logger.Error("addPackage - encoding package", "error", err.Error(), "packageID", pkg.ID)
	}
}

func (ctl *controller) updatePackage(c web.C, w http.ResponseWriter, r *http.Request) {
	pkg := &api.Package{}
	if err := json.NewDecoder(r.Body).Decode(pkg); err != nil {
		logger.Error("updatePackage - decoding payload", "error", err.Error())
		httpError(w, http.StatusBadRequest)
		return
	}
	pkg.ID = c.URLParams["package_id"]
	pkg.ApplicationID = c.URLParams["app_id"]

	err := ctl.api.UpdatePackage(pkg)
	if err != nil {
		logger.Error("updatePackage - updating package", "error", err.Error(), "package", pkg)
		httpError(w, http.StatusBadRequest)
		return
	}

	pkg, err = ctl.api.GetPackage(pkg.ID)
	if err != nil {
		logger.Error("addPackage - getting updated package", "error", err.Error(), "packageID", pkg.ID)
		httpError(w, http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(pkg); err != nil {
		logger.Error("updatePackage - encoding package", "error", err.Error(), "packageID", pkg.ID)
	}
}

func (ctl *controller) deletePackage(c web.C, w http.ResponseWriter, r *http.Request) {
	packageID := c.URLParams["package_id"]

	err := ctl.api.DeletePackage(packageID)
	switch err {
	case nil:
		httpError(w, http.StatusNoContent)
	default:
		logger.Error("deletePackage", "error", err.Error(), "packageID", packageID)
		httpError(w, http.StatusBadRequest)
	}
}

func (ctl *controller) getPackage(c web.C, w http.ResponseWriter, r *http.Request) {
	packageID := c.URLParams["package_id"]

	pkg, err := ctl.api.GetPackage(packageID)
	switch err {
	case nil:
		if err := json.NewEncoder(w).Encode(pkg); err != nil {
			logger.Error("getPackage - encoding package", "error", err.Error(), "packageID", packageID)
		}
	case sql.ErrNoRows:
		httpError(w, http.StatusNotFound)
	default:
		logger.Error("getPackage - getting package", "error", err.Error(), "packageID", packageID)
		httpError(w, http.StatusBadRequest)
	}
}

func (ctl *controller) getPackages(c web.C, w http.ResponseWriter, r *http.Request) {
	appID := c.URLParams["app_id"]
	page, _ := strconv.ParseUint(r.URL.Query().Get("page"), 10, 64)
	perPage, _ := strconv.ParseUint(r.URL.Query().Get("perpage"), 10, 64)

	pkgs, err := ctl.api.GetPackages(appID, page, perPage)
	switch err {
	case nil:
		if err := json.NewEncoder(w).Encode(pkgs); err != nil {
			logger.Error("getPackages - encoding packages", "error", err.Error(), "appID", appID)
		}
	case sql.ErrNoRows:
		httpError(w, http.StatusNotFound)
	default:
		logger.Error("getPackages - getting packages", "error", err.Error(), "appID", appID)
		httpError(w, http.StatusBadRequest)
	}
}

// ----------------------------------------------------------------------------
// API: instances
//

func (ctl *controller) getInstanceStatusHistory(c web.C, w http.ResponseWriter, r *http.Request) {
	appID := c.URLParams["app_id"]
	groupID := c.URLParams["group_id"]
	instanceID := c.URLParams["instance_id"]
	limit, _ := strconv.ParseUint(r.URL.Query().Get("limit"), 10, 64)

	instanceStatusHistory, err := ctl.api.GetInstanceStatusHistory(instanceID, appID, groupID, limit)
	switch err {
	case nil:
		if err := json.NewEncoder(w).Encode(instanceStatusHistory); err != nil {
			logger.Error("getInstanceStatusHistory - encoding status history", "error", err.Error(), "appID", appID, "groupID", groupID, "instanceID", instanceID, "limit", limit)
		}
	case sql.ErrNoRows:
		httpError(w, http.StatusNotFound)
	default:
		logger.Error("getInstanceStatusHistory - getting status history", "error", err.Error(), "appID", appID, "groupID", groupID, "instanceID", instanceID, "limit", limit)
		httpError(w, http.StatusBadRequest)
	}
}

func (ctl *controller) getInstances(c web.C, w http.ResponseWriter, r *http.Request) {
	appID := c.URLParams["app_id"]
	groupID := c.URLParams["group_id"]

	p := api.InstancesQueryParams{
		ApplicationID: appID,
		GroupID:       groupID,
		Version:       r.URL.Query().Get("version"),
	}
	p.Status, _ = strconv.Atoi(r.URL.Query().Get("status"))
	p.Page, _ = strconv.ParseUint(r.URL.Query().Get("page"), 10, 64)
	p.PerPage, _ = strconv.ParseUint(r.URL.Query().Get("perpage"), 10, 64)

	instances, err := ctl.api.GetInstances(p)
	switch err {
	case nil:
		if err := json.NewEncoder(w).Encode(instances); err != nil {
			logger.Error("getInstances - encoding instances", "error", err.Error(), "params", p)
		}
	case sql.ErrNoRows:
		httpError(w, http.StatusNotFound)
	default:
		logger.Error("getInstances - getting instances", "error", err.Error(), "params", p)
		httpError(w, http.StatusBadRequest)
	}
}

// ----------------------------------------------------------------------------
// API: activity
//

func (ctl *controller) getActivity(c web.C, w http.ResponseWriter, r *http.Request) {
	teamID, _ := c.Env["team_id"].(string)

	p := api.ActivityQueryParams{
		AppID:      r.URL.Query().Get("app"),
		GroupID:    r.URL.Query().Get("group"),
		ChannelID:  r.URL.Query().Get("channel"),
		InstanceID: r.URL.Query().Get("instance"),
		Version:    r.URL.Query().Get("version"),
	}
	p.Severity, _ = strconv.Atoi(r.URL.Query().Get("severity"))
	p.Start, _ = time.Parse(time.RFC3339, r.URL.Query().Get("start"))
	p.End, _ = time.Parse(time.RFC3339, r.URL.Query().Get("end"))
	p.Page, _ = strconv.ParseUint(r.URL.Query().Get("page"), 10, 64)
	p.PerPage, _ = strconv.ParseUint(r.URL.Query().Get("perpage"), 10, 64)

	activityEntries, err := ctl.api.GetActivity(teamID, p)
	switch err {
	case nil:
		if err := json.NewEncoder(w).Encode(activityEntries); err != nil {
			logger.Error("getActivity - encoding activity entries", "error", err.Error(), "params", p)
		}
	case sql.ErrNoRows:
		httpError(w, http.StatusNotFound)
	default:
		logger.Error("getActivity", "error", err, "teamID", teamID, "params", p)
		httpError(w, http.StatusBadRequest)
	}
}

// ----------------------------------------------------------------------------
// Metrics
//

const (
	appInstancesPerChannelMetricsProlog = `# HELP nebraska_application_instances_per_channel A number of applications from specific channel running on instances
# TYPE nebraska_application_instances_per_channel gauge`
	failedUpdatesMetricsProlog = `# HELP nebraska_failed_updates A number of failed updates of an application
# TYPE nebraska_failed_updates gauge`
)

func escapeMetricString(str string) string {
	str = strings.Replace(str, `\`, `\\`, -1)
	str = strings.Replace(str, `"`, `\"`, -1)
	str = strings.Replace(str, "\n", `\n`, -1)
	return str
}

func (ctl *controller) getMetrics(c web.C, w http.ResponseWriter, r *http.Request) {
	teamID, _ := c.Env["team_id"].(string)

	nowUnixMillis := time.Now().Unix() * 1000
	aipcMetrics, err := ctl.api.GetAppInstancesPerChannelMetrics(teamID)
	if err != nil {
		logger.Error("getMetrics - getting app instances per channel metrics", "error", err.Error(), "teamID", teamID)
		httpError(w, http.StatusBadRequest)
		return
	}
	fuMetrics, err := ctl.api.GetFailedUpdatesMetrics(teamID)
	if err != nil {
		logger.Error("getMetrics - getting failed updates metrics", "error", err.Error(), "teamID", teamID)
		httpError(w, http.StatusBadRequest)
		return
	}

	// "version" specifies a version of prometheus text file
	// format. For details see:
	//
	// https://github.com/prometheus/docs/blob/master/content/docs/instrumenting/exposition_formats.md#basic-info
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.WriteHeader(http.StatusOK)
	needEmptyLine := false
	if len(aipcMetrics) > 0 {
		if needEmptyLine {
			fmt.Fprintf(w, "\n")
		}
		fmt.Fprintf(w, "%s\n", appInstancesPerChannelMetricsProlog)
		for _, metric := range aipcMetrics {
			fmt.Fprintf(w, `nebraska_application_instances_per_channel{application="%s",version="%s",channel="%s"} %d %d%s`, escapeMetricString(metric.ApplicationName), escapeMetricString(metric.Version), escapeMetricString(metric.ChannelName), metric.InstancesCount, nowUnixMillis, "\n")
		}
		needEmptyLine = true
	}
	if len(fuMetrics) > 0 {
		if needEmptyLine {
			fmt.Fprintf(w, "\n")
		}
		fmt.Fprintf(w, "%s\n", failedUpdatesMetricsProlog)
		for _, metric := range fuMetrics {
			fmt.Fprintf(w, `nebraska_failed_updates{application="%s"} %d %d%s`, escapeMetricString(metric.ApplicationName), metric.FailureCount, nowUnixMillis, "\n")
		}
	}
}

// ----------------------------------------------------------------------------
// OMAHA server
//

func (ctl *controller) processOmahaRequest(c web.C, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/xml")
	if err := ctl.omahaHandler.Handle(r.Body, w, getRequestIP(r)); err != nil {
		logger.Error("process omaha request", "error", err)
	}
}

// ----------------------------------------------------------------------------
// Helpers
//

func getRequestIP(r *http.Request) string {
	ips := strings.Split(r.Header.Get("X-FORWARDED-FOR"), ", ")
	if ips[0] != "" && net.ParseIP(ips[0]) != nil {
		return ips[0]
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

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
	"time"

	"github.com/google/go-github/github"
	"github.com/ymichael/sessions"
	"github.com/zenazn/goji/web"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"

	"github.com/coreroller/coreroller/pkg/api"
	"github.com/coreroller/coreroller/pkg/omaha"
	"github.com/coreroller/coreroller/pkg/syncer"
)

const (
	clientIDEnvName      = "COREROLLER_OAUTH_CLIENT_ID"
	clientSecretEnvName  = "COREROLLER_OAUTH_CLIENT_SECRET"
	sessionSecretEnvName = "COREROLLER_SESSION_SECRET"
	webhookSecretEnvName = "COREROLLER_WEBHOOK_SECRET"
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
}

type controllerConfig struct {
	enableSyncer       bool
	hostCoreosPackages bool
	coreosPackagesPath string
	corerollerURL      string
	sessionSecret      string
	oauthClientID      string
	oauthClientSecret  string
	webhookSecret      string
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
	return "", errors.New("no oauth client ID passed to rollerd")
}

func obtainOAuthClientSecret(potentialSecret string) (string, error) {
	if secret := getPotentialOrEnv(potentialSecret, clientSecretEnvName); secret != "" {
		return secret, nil
	}
	return "", errors.New("no oauth client secret passed to rollerd")
}

func obtainWebhookSecret(potentialSecret string) (string, error) {
	if secret := getPotentialOrEnv(potentialSecret, webhookSecretEnvName); secret != "" {
		return secret, nil
	}
	return "", errors.New("no webhook secret passed to rollerd")
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
		sessions:     sessions.NewSessionOptions(sessionSecret, sessions.MemoryStore{}),
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
			Scopes:       []string{"read:org"},
			Endpoint:     githuboauth.Endpoint,
		},
		userSessionIDs: make(userSessionMap),
		teamToUsers:    make(teamToUsersMap),
		webhookSecret:  webhookSecret,
	}

	if conf.enableSyncer {
		syncerConf := &syncer.Config{
			Api:          api,
			HostPackages: conf.hostCoreosPackages,
			PackagesPath: conf.coreosPackagesPath,
			PackagesURL:  conf.corerollerURL + coreosPkgsRouterPrefix,
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
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		case resultInternalFailure:
			ctl.cleanupSession(&c, w)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
		logger.Error("login cb", "invalid oauth state, expected %q, got %q", expectedState, state)
		return
	}
	code := r.FormValue("code")
	logger.Debug("login cb", "received code", code)
	ctx := context.Background()
	token, err := ctl.oauthConfig.Exchange(ctx, code)
	if err != nil {
		logger.Error("login cb", "oauth exchange failed: %v", err)
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
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		case resultInternalFailure:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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

	teams, err := ctl.api.GetTeams()
	if err != nil {
		logger.Error("login dance", "failed to get teams", err)
		result = resultInternalFailure
		return
	}
	teamsMap := make(map[string]*api.Team, len(teams))
	for _, team := range teams {
		teamsMap[team.Name] = team
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
			corerollerTeamName := makeTeamName(*ghTeam.Organization.Login, *ghTeam.Name)
			// TODO(krnowak): This sucks. If coreroller
			// has two teams (say kubernetes and habitat)
			// and we have such teams in github kinbolk
			// organization then we are going to randomly
			// get an ID of either coreroller team…
			logger.Debug("login dance", "trying to find a matching coreroller team", corerollerTeamName)
			if team, ok := teamsMap[corerollerTeamName]; ok {
				logger.Debug("login dance", "found matching team", corerollerTeamName)
				teamData.org = *ghTeam.Organization.Login
				teamData.team = ghTeam.Name
				teamID = team.ID
				break
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
				logger.Debug("login dance", "trying to find a matching coreroller team", *ghOrg.Login)
				if team, ok := teamsMap[*ghOrg.Login]; ok {
					logger.Debug("login dance", "found matching team", *ghOrg.Login)
					teamData.org = *ghOrg.Login
					teamID = team.ID
					break
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

func (ctl *controller) authMissingTeamID(w http.ResponseWriter, r *http.Request, obj map[string]interface{}) {
	oauthState := sessions.GenerateRandomString(64)
	obj["state"] = oauthState
	obj["desiredurl"] = r.URL.String()
	logger.Debug("authenticate", "oauthstate", oauthState)
	url := ctl.oauthConfig.AuthCodeURL(oauthState, oauth2.AccessTypeOnline)
	logger.Debug("authenticate", "redirecting to", url)
	redirectTo(w, r, url)
}

// authenticate is a middleware handler in charge of authenticating requests.
func (ctl *controller) authenticate(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		obj := ctl.sessions.GetSessionObject(c)
		teamID, ok := obj["teamID"]
		if !ok {
			ctl.authMissingTeamID(w, r, obj)
			return
		}

		c.Env["team_id"] = teamID
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
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	app.TeamID = c.Env["team_id"].(string)

	_, err := ctl.api.AddAppCloning(app, sourceAppID)
	if err != nil {
		logger.Error("addApp - cloning app", "error", err.Error(), "app", app, "sourceAppID", sourceAppID)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	app, err = ctl.api.GetApp(app.ID)
	if err != nil {
		logger.Error("addApp - getting added app", "error", err.Error(), "appID", app.ID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	app.ID = c.URLParams["app_id"]
	app.TeamID = c.Env["team_id"].(string)

	err := ctl.api.UpdateApp(app)
	if err != nil {
		logger.Error("updatedApp - updating app", "error", err.Error(), "app", app)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	app, err = ctl.api.GetApp(app.ID)
	if err != nil {
		logger.Error("updateApp - getting updated app", "error", err.Error(), "appID", app.ID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
		http.Error(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)
	default:
		logger.Error("deleteApp", "error", err.Error(), "appID", appID)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
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
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	default:
		logger.Error("getApp - getting app", "error", err.Error(), "appID", appID)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
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
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	default:
		logger.Error("getApps - getting apps", "error", err.Error(), "teamID", teamID)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

// ----------------------------------------------------------------------------
// API: groups CRUD
//

func (ctl *controller) addGroup(c web.C, w http.ResponseWriter, r *http.Request) {
	group := &api.Group{}
	if err := json.NewDecoder(r.Body).Decode(group); err != nil {
		logger.Error("addGroup - decoding payload", "error", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	group.ApplicationID = c.URLParams["app_id"]

	_, err := ctl.api.AddGroup(group)
	if err != nil {
		logger.Error("addGroup - adding group", "error", err.Error(), "group", group)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	group, err = ctl.api.GetGroup(group.ID)
	if err != nil {
		logger.Error("addGroup - getting added group", "error", err.Error(), "groupID", group.ID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	group.ID = c.URLParams["group_id"]
	group.ApplicationID = c.URLParams["app_id"]

	err := ctl.api.UpdateGroup(group)
	if err != nil {
		logger.Error("updateGroup - updating group", "error", err.Error(), "group", group)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	group, err = ctl.api.GetGroup(group.ID)
	if err != nil {
		logger.Error("updateGroup - fetching updated group", "error", err.Error(), "groupID", group.ID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
		http.Error(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)
	default:
		logger.Error("deleteGroup", "error", err.Error(), "groupID", groupID)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
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
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	default:
		logger.Error("getGroup - getting group", "error", err.Error(), "groupID", groupID)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
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
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	default:
		logger.Error("getGroups - getting groups", "error", err.Error(), "appID", appID)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

// ----------------------------------------------------------------------------
// API: channels CRUD
//

func (ctl *controller) addChannel(c web.C, w http.ResponseWriter, r *http.Request) {
	channel := &api.Channel{}
	if err := json.NewDecoder(r.Body).Decode(channel); err != nil {
		logger.Error("addChannel", "error", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	channel.ApplicationID = c.URLParams["app_id"]

	_, err := ctl.api.AddChannel(channel)
	if err != nil {
		logger.Error("addChannel", "error", err.Error(), "channel", channel)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	channel, err = ctl.api.GetChannel(channel.ID)
	if err != nil {
		logger.Error("addChannel", "error", err.Error(), "channelID", channel.ID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	channel.ID = c.URLParams["channel_id"]
	channel.ApplicationID = c.URLParams["app_id"]

	err := ctl.api.UpdateChannel(channel)
	if err != nil {
		logger.Error("updateChannel - updating channel", "error", err.Error(), "channel", channel)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	channel, err = ctl.api.GetChannel(channel.ID)
	if err != nil {
		logger.Error("updateChannel - getting channel updated", "error", err.Error(), "channelID", channel.ID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
		http.Error(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)
	default:
		logger.Error("deleteChannel", "error", err.Error(), "channelID", channelID)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
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
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	default:
		logger.Error("getChannel - getting updated channel", "error", err.Error(), "channelID", channelID)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
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
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	default:
		logger.Error("getChannels - getting channels", "error", err.Error(), "appID", appID)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

// ----------------------------------------------------------------------------
// API: packages CRUD
//

func (ctl *controller) addPackage(c web.C, w http.ResponseWriter, r *http.Request) {
	pkg := &api.Package{}
	if err := json.NewDecoder(r.Body).Decode(pkg); err != nil {
		logger.Error("addPackage - decoding payload", "error", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	pkg.ApplicationID = c.URLParams["app_id"]

	_, err := ctl.api.AddPackage(pkg)
	if err != nil {
		logger.Error("addPackage - adding package", "error", err.Error(), "package", pkg)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	pkg, err = ctl.api.GetPackage(pkg.ID)
	if err != nil {
		logger.Error("addPackage - getting added package", "error", err.Error(), "packageID", pkg.ID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	pkg.ID = c.URLParams["package_id"]
	pkg.ApplicationID = c.URLParams["app_id"]

	err := ctl.api.UpdatePackage(pkg)
	if err != nil {
		logger.Error("updatePackage - updating package", "error", err.Error(), "package", pkg)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	pkg, err = ctl.api.GetPackage(pkg.ID)
	if err != nil {
		logger.Error("addPackage - getting updated package", "error", err.Error(), "packageID", pkg.ID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
		http.Error(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)
	default:
		logger.Error("deletePackage", "error", err.Error(), "packageID", packageID)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
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
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	default:
		logger.Error("getPackage - getting package", "error", err.Error(), "packageID", packageID)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
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
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	default:
		logger.Error("getPackages - getting packages", "error", err.Error(), "appID", appID)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
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
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	default:
		logger.Error("getInstanceStatusHistory - getting status history", "error", err.Error(), "appID", appID, "groupID", groupID, "instanceID", instanceID, "limit", limit)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
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
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	default:
		logger.Error("getInstances - getting instances", "error", err.Error(), "params", p)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
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
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	default:
		logger.Error("getActivity", "error", err, "teamID", teamID, "params", p)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

// ----------------------------------------------------------------------------
// Metrics
//

const (
	app_instances_per_channel_metrics_prolog = `# HELP coreroller_application_instances_per_channel A number of applications from specific channel running on instances
# TYPE coreroller_application_instances_per_channel gauge`
	failed_updates_metrics_prolog = `# HELP coreroller_failed_updates A number of failed updates of an application
# TYPE coreroller_failed_updates gauge`
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
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	fuMetrics, err := ctl.api.GetFailedUpdatesMetrics(teamID)
	if err != nil {
		logger.Error("getMetrics - getting failed updates metrics", "error", err.Error(), "teamID", teamID)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
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
		fmt.Fprintf(w, "%s\n", app_instances_per_channel_metrics_prolog)
		for _, metric := range aipcMetrics {
			fmt.Fprintf(w, `coreroller_application_instances_per_channel{application="%s",version="%s",channel="%s"} %d %d%s`, escapeMetricString(metric.ApplicationName), escapeMetricString(metric.Version), escapeMetricString(metric.ChannelName), metric.InstancesCount, nowUnixMillis, "\n")
		}
		needEmptyLine = true
	}
	if len(fuMetrics) > 0 {
		if needEmptyLine {
			fmt.Fprintf(w, "\n")
		}
		fmt.Fprintf(w, "%s\n", failed_updates_metrics_prolog)
		for _, metric := range fuMetrics {
			fmt.Fprintf(w, `coreroller_failed_updates{application="%s"} %d %d%s`, escapeMetricString(metric.ApplicationName), metric.FailureCount, nowUnixMillis, "\n")
		}
		needEmptyLine = true
	}
}

// ----------------------------------------------------------------------------
// OMAHA server
//

func (ctl *controller) processOmahaRequest(c web.C, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/xml")
	ctl.omahaHandler.Handle(r.Body, w, getRequestIP(r))
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

package updater

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"log"
	"os"
	"testing"

	omahaSpec "github.com/kinvolk/go-omaha/omaha"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/kinvolk/nebraska/backend/pkg/api"
	"github.com/kinvolk/nebraska/backend/pkg/omaha"
)

const (
	defaultTestDbURL string = "postgres://postgres:nebraska@127.0.0.1:5432/nebraska_tests?sslmode=disable&connect_timeout=10"
)

type testOmahaHandler struct {
	handler *omaha.Handler
}

func newTestHandler(api *api.API) *testOmahaHandler {
	return &testOmahaHandler{
		handler: omaha.NewHandler(api),
	}
}

func (h *testOmahaHandler) Handle(ctx context.Context, url string, req *omahaSpec.Request) (*omahaSpec.Response, error) {
	requestBuf := bytes.NewBuffer(nil)
	encoder := xml.NewEncoder(requestBuf)
	err := encoder.Encode(req)
	if err != nil {
		return nil, err
	}

	omahaRespXML := new(bytes.Buffer)
	if err = h.handler.Handle(requestBuf, omahaRespXML, "0.1.0.0"); err != nil {
		return nil, err
	}

	var omahaResp *omahaSpec.Response
	err = xml.NewDecoder(omahaRespXML).Decode(&omahaResp)
	if err != nil {
		return nil, err
	}

	return omahaResp, nil
}

func newForTest(t *testing.T) *api.API {
	if _, ok := os.LookupEnv("NEBRASKA_DB_URL"); !ok {
		log.Printf("NEBRASKA_DB_URL not set, setting to default %q\n", defaultTestDbURL)
		_ = os.Setenv("NEBRASKA_DB_URL", defaultTestDbURL)
	}
	api, err := api.New(api.OptionInitDB)

	require.NoError(t, err)
	require.NotNil(t, api)

	return api
}

func TestNewUpdater(t *testing.T) {
	t.Run("valid_config", func(t *testing.T) {
		conf := Config{
			OmahaURL:        "http://localhost:8000",
			AppID:           "io.phony.App",
			Channel:         "stable",
			InstanceID:      "instance001",
			InstanceVersion: "0.1.0",
		}
		_, err := New(conf)
		assert.NoError(t, err)
	})

	t.Run("invalid_config", func(t *testing.T) {
		conf := Config{
			OmahaURL:        "http://invalidurl.test\\",
			AppID:           "io.phony.App",
			Channel:         "stable",
			InstanceID:      "instance001",
			InstanceVersion: "0.1.0",
		}
		updater, err := New(conf)
		require.Error(t, err)
		assert.Nil(t, updater)
	})
}

func TestCheckForUpdates(t *testing.T) {
	apiInstance := newForTest(t)

	t.Cleanup(func() {
		apiInstance.Close()
	})

	appID, group, tChannel := setup(&config{t: t, api: apiInstance, pkgVersion: "0.1.0", policySafeMode: true, policyMaxUpdatesPerPeriod: 2})

	u, err := New(Config{
		OmahaURL:        "http://localhost:8000",
		AppID:           appID,
		Channel:         group.Track,
		InstanceID:      "instance001",
		InstanceVersion: "0.2.0",
		OmahaReqHandler: newTestHandler(apiInstance),
	})
	require.NoError(t, err)

	info, err := u.CheckForUpdates(context.TODO())
	require.NoError(t, err)
	assert.False(t, info.HasUpdate)
	assert.Equal(t, "", info.Version)

	newPkg, err := apiInstance.AddPackage(&api.Package{Type: api.PkgTypeOther, URL: "http://sample.url/pkg", Version: "0.3.0", ApplicationID: appID, Arch: api.ArchAMD64, Filename: null.StringFrom("updatefile.txt")})
	require.NoError(t, err)
	tChannel.PackageID = null.StringFrom(newPkg.ID)
	err = apiInstance.UpdateChannel(tChannel)
	require.NoError(t, err)

	info, err = u.CheckForUpdates(context.TODO())
	require.NoError(t, err)

	expected := UpdateInfo{
		AppID:        appID,
		HasUpdate:    true,
		Version:      "0.3.0",
		UpdateStatus: "ok",
		URLs: []string{
			"http://sample.url/pkg",
		},
		Packages: []*omahaSpec.Package{
			{Name: "updatefile.txt", Required: true},
		},
	}
	assert.Equal(t, expected.AppID, info.AppID)
	assert.Equal(t, expected.HasUpdate, info.HasUpdate)
	assert.Equal(t, expected.Version, info.Version)
	assert.Equal(t, expected.UpdateStatus, info.UpdateStatus)
	assert.Equal(t, expected.URLs, info.URLs)
	assert.Equal(t, expected.Packages, info.Packages)
}

type updateTestHandler struct {
	fetchUpdateResult error
	applyUpdateResult error
}

func (u updateTestHandler) FetchUpdate(ctx context.Context, info UpdateInfo) error {
	return u.fetchUpdateResult
}

func (u updateTestHandler) ApplyUpdate(ctx context.Context, info UpdateInfo) error {
	return u.applyUpdateResult
}

type config struct {
	t                         *testing.T
	api                       *api.API
	pkgVersion                string
	policySafeMode            bool
	policyMaxUpdatesPerPeriod int
}

func setup(cnf *config) (string, *api.Group, *api.Channel) {
	cnf.t.Helper()
	tTeam, err := cnf.api.AddTeam(&api.Team{Name: "test_team"})
	require.NoError(cnf.t, err)
	tApp, err := cnf.api.AddApp(&api.Application{Name: "io.phony.App", TeamID: tTeam.ID})
	require.NoError(cnf.t, err)
	tPkg, err := cnf.api.AddPackage(&api.Package{Type: api.PkgTypeOther, URL: "http://sample.url/pkg", Version: cnf.pkgVersion, ApplicationID: tApp.ID, Arch: api.ArchAMD64})
	require.NoError(cnf.t, err)
	tChannel, err := cnf.api.AddChannel(&api.Channel{Name: "channel1", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID), Arch: api.ArchAMD64})
	require.NoError(cnf.t, err)
	tGroup, err := cnf.api.AddGroup(&api.Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: cnf.policySafeMode, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: cnf.policyMaxUpdatesPerPeriod, PolicyUpdateTimeout: "60 minutes", Track: "stable"})
	require.NoError(cnf.t, err)
	return tApp.ID, tGroup, tChannel
}

func TestTryUpdate(t *testing.T) {
	api := newForTest(t)

	t.Cleanup(func() {
		api.Close()
	})

	oldVersion := "0.2.0"
	pkgVersion := "0.4.0"
	appID, group, _ := setup(&config{t: t, api: api, pkgVersion: pkgVersion, policySafeMode: false, policyMaxUpdatesPerPeriod: 10})

	t.Run("returns_error_when", func(t *testing.T) {
		tests := []struct {
			name              string
			fetchUpdateResult error
			applyUpdateResult error
		}{
			{
				name:              "fetching_update",
				fetchUpdateResult: errors.New("something went wrong fetching the update"),
				applyUpdateResult: nil,
			},
			{
				name:              "applying_update",
				fetchUpdateResult: nil,
				applyUpdateResult: errors.New("something went wrong fetching the update"),
			},
		}

		for _, tc := range tests {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				u, err := New(Config{
					OmahaURL:        "http://localhost:8000",
					AppID:           appID,
					Channel:         group.Track,
					InstanceID:      "instance001",
					InstanceVersion: oldVersion,
					OmahaReqHandler: newTestHandler(api),
					Debug:           false,
				})
				require.NoError(t, err)

				assert.Equal(t, oldVersion, u.InstanceVersion())

				err = u.TryUpdate(context.TODO(), &updateTestHandler{
					fetchUpdateResult: tc.fetchUpdateResult,
					applyUpdateResult: tc.applyUpdateResult,
				})

				assert.Error(t, err)
				assert.Equal(t, oldVersion, u.InstanceVersion())

				// Get Instance Last status and check if the status is 3
				// status 3 is the internal code for error state.
				statusHistory, err := api.GetInstanceStatusHistory("instance001", appID, group.ID, 1)
				require.NoError(t, err)
				require.NotEqual(t, 0, len(statusHistory))
				assert.Equal(t, 3, statusHistory[0].Status)
			})
		}
	})

	t.Run("when_succeeds_marks_update_as_complete", func(t *testing.T) {
		u, err := New(Config{
			OmahaURL:        "http://localhost:8000",
			AppID:           appID,
			Channel:         group.Track,
			InstanceID:      "instance001",
			InstanceVersion: oldVersion,
			OmahaReqHandler: newTestHandler(api),
			Debug:           false,
		})
		require.NoError(t, err)

		assert.Equal(t, oldVersion, u.InstanceVersion())
		err = u.TryUpdate(context.TODO(), &updateTestHandler{
			fetchUpdateResult: nil,
			applyUpdateResult: nil,
		})

		// Check if version is changed.
		require.NoError(t, err)
		assert.Equal(t, pkgVersion, u.InstanceVersion())

		// Check if no new update is available for instance.
		err = u.TryUpdate(context.TODO(), &updateTestHandler{
			fetchUpdateResult: nil,
			applyUpdateResult: nil,
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, NoUpdateError{})
	})
}

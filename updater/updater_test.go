package updater

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"log"
	"os"
	"testing"

	"github.com/google/uuid"
	omahaSpec "github.com/kinvolk/go-omaha/omaha"
	"github.com/kinvolk/nebraska/backend/pkg/api"
	"github.com/kinvolk/nebraska/backend/pkg/omaha"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

const (
	defaultTestDbURL string = "postgres://postgres:nebraska@127.0.0.1:5432/nebraska_tests?sslmode=disable&connect_timeout=10"
)

type TestOmahaHandler struct {
	handler *omaha.Handler
}

func NewTestHandler(a *api.API) *TestOmahaHandler {
	return &TestOmahaHandler{
		handler: omaha.NewHandler(a),
	}
}

func (h *TestOmahaHandler) Handle(req *omahaSpec.Request) (*omahaSpec.Response, error) {
	omahaReqXML, err := xml.Marshal(req)
	if err != nil {
		return nil, err
	}

	omahaRespXML := new(bytes.Buffer)
	if err = h.handler.Handle(bytes.NewReader(omahaReqXML), omahaRespXML, "0.1.0.0"); err != nil {
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
	a, err := api.New(api.OptionInitDB)

	require.NoError(t, err)
	require.NotNil(t, a)

	return a
}

func newUpdater(a *api.API, appID string, channel string, instanceID string, instanceVersion string) *Updater {
	return &Updater{
		omahaURL:        "http://localhost:8000",
		clientVersion:   "go-omaha",
		instanceID:      instanceID,
		sessionID:       uuid.New().String(),
		appID:           appID,
		instanceVersion: instanceVersion,
		channel:         channel,
		omahaReqHandler: NewTestHandler(a),
	}
}

func TestNewUpdater(t *testing.T) {
	_, err := New("http://localhost:8000", "io.phony.App", "stable", "instance001", "0.1.0")
	assert.NoError(t, err)
}

func TestCheckForUpdates(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&api.Team{Name: "test_team"})
	tApp, _ := a.AddApp(&api.Application{Name: "io.phony.App", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&api.Package{Type: api.PkgTypeOther, URL: "http://sample.url/pkg", Version: "0.1.0", ApplicationID: tApp.ID, Arch: api.ArchAMD64})
	tChannel, _ := a.AddChannel(&api.Channel{Name: "channel1", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID), Arch: api.ArchAMD64})
	tGroup, err := a.AddGroup(&api.Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes", Track: "stable"})
	assert.NoError(t, err)

	u := newUpdater(a, tApp.ID, tGroup.Track, "instance-0001", "0.2.0")
	info, err := u.CheckForUpdates(context.TODO())
	assert.NoError(t, err)
	assert.False(t, info.HasUpdate)
	assert.Equal(t, "", info.GetVersion())

	newPkg, _ := a.AddPackage(&api.Package{Type: api.PkgTypeOther, URL: "http://sample.url/pkg", Version: "0.3.0", ApplicationID: tApp.ID, Arch: api.ArchAMD64})
	tChannel.PackageID = null.StringFrom(newPkg.ID)
	err = a.UpdateChannel(tChannel)
	assert.NoError(t, err)

	info, err = u.CheckForUpdates(context.TODO())
	assert.NoError(t, err)
	assert.True(t, info.HasUpdate)
}

type updateTestHandler struct {
	fetchUpdateResult error
	applyUpdateResult error
}

func (u updateTestHandler) FetchUpdate(ctx context.Context, info *UpdateInfo) error {
	// Nothing to do
	return u.fetchUpdateResult
}

func (u updateTestHandler) ApplyUpdate(ctx context.Context, info *UpdateInfo) error {
	// Nothing to do
	return u.applyUpdateResult
}

func TestTryUpdate(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&api.Team{Name: "test_team"})
	tApp, _ := a.AddApp(&api.Application{Name: "io.phony.App", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&api.Package{Type: api.PkgTypeOther, URL: "http://sample.url/pkg", Version: "0.4.0", ApplicationID: tApp.ID, Arch: api.ArchAMD64})
	tChannel, _ := a.AddChannel(&api.Channel{Name: "channel1", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID), Arch: api.ArchAMD64})
	tGroup, err := a.AddGroup(&api.Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: false, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 10, PolicyUpdateTimeout: "60 minutes", Track: "stable"})
	assert.NoError(t, err)

	oldVersion := "0.2.0"

	u := newUpdater(a, tApp.ID, tGroup.Track, "instance-0001", "0.2.0")
	assert.Equal(t, oldVersion, u.instanceVersion)

	// Error when fetching update
	err = u.TryUpdate(context.TODO(), &updateTestHandler{
		errors.New("something went wrong fetching the update"),
		nil,
	})
	assert.Error(t, err)
	assert.Equal(t, oldVersion, u.instanceVersion)

	// Error when applying update
	err = u.TryUpdate(context.TODO(), &updateTestHandler{
		nil,
		errors.New("something went wrong applying the update"),
	})
	assert.Error(t, err)
	assert.Equal(t, oldVersion, u.instanceVersion)

	err = u.TryUpdate(context.TODO(), updateTestHandler{
		nil,
		nil,
	})
	assert.NoError(t, err)
	assert.Equal(t, tPkg.Version, u.instanceVersion)
}

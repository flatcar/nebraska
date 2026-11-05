package api

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api/runtime"
	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

// GetUpdatePackage registers the instance once and decides on the instance as
// registered. It must give the same answer, and leave the same rows behind, as
// registering the instance first, which is what a ping plus an update check
// used to do.
func TestGetUpdatePackageMatchesRegisteringFirst(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)
	rs := runtimeSvc(a)

	tTeam, err := as.AddTeam(&types.Team{Name: "test_team"})
	require.NoError(t, err)
	tApp, err := as.AddApp(&types.Application{Name: "test_app", TeamID: tTeam.ID})
	require.NoError(t, err)
	tPkg, err := as.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	require.NoError(t, err)
	tChannel, err := as.AddChannel(&types.Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	require.NoError(t, err)
	newGroup := func(name string) *types.Group {
		g, err := as.AddGroup(&types.Group{Name: name, ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true,
			PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 1000, PolicyUpdateTimeout: "60 minutes"})
		require.NoError(t, err)
		return g
	}
	group, otherGroup := newGroup("group"), newGroup("other_group")

	type stored struct {
		ip         string
		appRow     bool
		groupID    string
		version    string
		status     null.Int
		lastUpdate null.String
		inProgress bool
		checkedAgo time.Duration
	}
	granted, downloading, complete := null.IntFrom(int64(types.InstanceStatusUpdateGranted)), null.IntFrom(int64(types.InstanceStatusDownloading)), null.IntFrom(int64(types.InstanceStatusComplete))
	scenarios := map[string]*stored{
		"new instance":                nil,
		"checked recently":            {ip: "10.0.0.1", appRow: true, groupID: group.ID, version: "12.0.0", status: complete, checkedAgo: time.Minute},
		"checked 10 minutes ago":      {ip: "10.0.0.1", appRow: true, groupID: group.ID, version: "12.0.0", status: complete, checkedAgo: 10 * time.Minute},
		"version changed":             {ip: "10.0.0.1", appRow: true, groupID: group.ID, version: "11.0.0", status: complete, checkedAgo: time.Minute},
		"group changed":               {ip: "10.0.0.1", appRow: true, groupID: otherGroup.ID, version: "12.0.0", status: complete, checkedAgo: time.Minute},
		"ip changed":                  {ip: "10.0.0.9", appRow: true, groupID: group.ID, version: "12.0.0", status: complete, checkedAgo: time.Minute},
		"ip changed and stale":        {ip: "10.0.0.9", appRow: true, groupID: group.ID, version: "12.0.0", status: complete, checkedAgo: 10 * time.Minute},
		"no application row":          {ip: "10.0.0.1"},
		"offline 2 days, complete":    {ip: "10.0.0.1", appRow: true, groupID: group.ID, version: "12.0.0", status: complete, lastUpdate: null.StringFrom("12.0.0"), checkedAgo: 48 * time.Hour},
		"offline 2 days, downloading": {ip: "10.0.0.1", appRow: true, groupID: group.ID, version: "12.0.0", status: downloading, lastUpdate: null.StringFrom("12.1.0"), inProgress: true, checkedAgo: 48 * time.Hour},
		"offline 2 days, granted":     {ip: "10.0.0.1", appRow: true, groupID: group.ID, version: "12.0.0", status: granted, lastUpdate: null.StringFrom("12.1.0"), inProgress: true, checkedAgo: 48 * time.Hour},
		"granted recently":            {ip: "10.0.0.1", appRow: true, groupID: group.ID, version: "12.0.0", status: granted, lastUpdate: null.StringFrom("12.1.0"), inProgress: true, checkedAgo: time.Minute},
		"downloading recently":        {ip: "10.0.0.1", appRow: true, groupID: group.ID, version: "12.0.0", status: downloading, lastUpdate: null.StringFrom("12.1.0"), inProgress: true, checkedAgo: time.Minute},
	}

	seed := func(id string, s *stored) {
		if s == nil {
			return
		}
		_, err := a.db().Exec("insert into instance (id, ip, alias) values ($1, $2, 'alias')", id, s.ip)
		require.NoError(t, err)
		if s.appRow {
			_, err = a.db().Exec(`insert into instance_application (instance_id, application_id, group_id, version, status,
				last_check_for_updates, last_update_version, update_in_progress) values ($1, $2, $3, $4, $5, $6, $7, $8)`,
				id, tApp.ID, s.groupID, s.version, s.status, time.Now().Add(-s.checkedAgo), s.lastUpdate, s.inProgress)
			require.NoError(t, err)
		}
	}
	state := func(id string) string {
		var out string
		err := a.db().QueryRow(`select i.ip || '|' || i.alias || '|' || ia.group_id || '|' || ia.version || '|' || coalesce(ia.status, -1)
			|| '|' || coalesce(ia.last_update_version, '-') || '|' || ia.update_in_progress
			|| '|' || (ia.last_check_for_updates > now() - interval '1 minute') || '|' || (ia.last_update_granted_ts is not null)
			|| '|' || (select count(*) from instance_status_history h where h.instance_id = i.id)
			from instance i join instance_application ia on ia.instance_id = i.id where i.id = $1`, id).Scan(&out)
		require.NoError(t, err)
		return out
	}
	outcome := func(pkg *types.Package, err error) string {
		if err != nil {
			return "error: " + err.Error()
		}
		return "package " + pkg.Version
	}

	for name, s := range scenarios {
		for _, version := range []string{"12.0.0", "12.1.0"} {
			t.Run(fmt.Sprintf("%s, reports %s", name, version), func(t *testing.T) {
				twice, once := uuid.NewString(), uuid.NewString()
				seed(twice, s)
				seed(once, s)
				inst := func(id string) types.Instance { return types.Instance{ID: id, IP: "10.0.0.1"} }
				instApp := runtime.NewInstanceApplication(tApp.ID, group.ID, version)

				_, _ = rs.RegisterInstance(inst(twice), instApp)
				want := outcome(rs.GetUpdatePackage(inst(twice), instApp))
				got := outcome(rs.GetUpdatePackage(inst(once), instApp))

				assert.Equal(t, want, got)
				assert.Equal(t, state(twice), state(once))
			})
		}
	}
}

// When RegisterInstance writes only one of the instance's two rows, it builds
// the instance it returns from the read it made before writing. On every path,
// it must return what reading the instance right after returns.
func TestRegisterInstanceMatchesReadAfterWrite(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)
	rs := runtimeSvc(a)

	tTeam, err := as.AddTeam(&types.Team{Name: "test_team"})
	require.NoError(t, err)
	tApp, err := as.AddApp(&types.Application{Name: "test_app", TeamID: tTeam.ID})
	require.NoError(t, err)
	group, err := as.AddGroup(&types.Group{Name: "group", ApplicationID: tApp.ID, PolicyUpdatesEnabled: true,
		PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 1000, PolicyUpdateTimeout: "60 minutes"})
	require.NoError(t, err)

	inst := types.Instance{ID: uuid.NewString(), IP: "10.0.0.1"}
	version := "12.0.0"
	checkedAgo := func(d time.Duration) func() {
		return func() {
			_, err := a.db().Exec("update instance_application set last_check_for_updates = $1 where instance_id = $2", time.Now().Add(-d), inst.ID)
			require.NoError(t, err)
		}
	}
	steps := []struct {
		name   string
		before func()
	}{
		{"new instance: both rows written", func() {}},
		{"nothing changed: nothing written", func() {}},
		{"new IP: instance row written", func() { inst.IP = "10.0.0.2" }},
		{"new version: application row written", func() { version = "12.1.0" }},
		{"checked 10 minutes ago: application row written", checkedAgo(10 * time.Minute)},
		{"checked 2 days ago: application row written, not in the first read", checkedAgo(48 * time.Hour)},
	}
	for _, step := range steps {
		step.before()
		got, err := rs.RegisterInstance(inst, runtime.NewInstanceApplication(tApp.ID, group.ID, version))
		require.NoError(t, err, step.name)
		want, err := a.GetInstance(inst.ID, tApp.ID)
		require.NoError(t, err, step.name)
		assert.Equal(t, want, got, step.name)
	}
}

package runtime

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/api"
	"github.com/flatcar/nebraska/backend/pkg/api/admin"
	"github.com/flatcar/nebraska/backend/pkg/api/internal/dbconn"
	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

func setupPruneGroup(t *testing.T, a *api.API, name string) (*admin.Service, *Service, *types.Application, *types.Group) {
	t.Helper()
	as := adminSvc(a)
	rs := runtimeSvc(a)
	tTeam, err := as.AddTeam(&types.Team{Name: name + "_team"})
	require.NoError(t, err)
	tApp, err := as.AddApp(&types.Application{Name: name + "_app", TeamID: tTeam.ID})
	require.NoError(t, err)
	tGroup, err := as.AddGroup(&types.Group{Name: name + "_group", ApplicationID: tApp.ID, PolicyUpdatesEnabled: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	require.NoError(t, err)
	return as, rs, tApp, tGroup
}

func backdateCheckIn(t *testing.T, a *api.API, instanceID string, age time.Duration) {
	t.Helper()
	_, err := dbconn.DB(a.Conn()).Exec(
		`UPDATE instance_application SET last_check_for_updates = $1 WHERE instance_id = $2`,
		time.Now().UTC().Add(-age),
		instanceID,
	)
	require.NoError(t, err)
}

func TestPruneStaleInstancesDeletesOldAndKeepsFresh(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	_, rs, tApp, tGroup := setupPruneGroup(t, a, "prune")

	stale, err := rs.RegisterInstance(types.Instance{ID: uuid.New().String(), IP: "10.0.0.1"}, NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
	require.NoError(t, err)
	fresh, err := rs.RegisterInstance(types.Instance{ID: uuid.New().String(), IP: "10.0.0.2"}, NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
	require.NoError(t, err)

	require.NoError(t, rs.grantUpdate(stale, "1.0.1"))
	require.NoError(t, rs.updateInstanceStatus(stale.ID, tApp.ID, types.InstanceStatusComplete))
	backdateCheckIn(t, a, stale.ID, 2*time.Hour)

	result, err := rs.PruneStaleInstances(time.Now().UTC().Add(-time.Hour), 500, false)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Deleted)
	assert.False(t, result.MoreRemain)

	_, err = a.GetInstance(stale.ID, tApp.ID)
	assert.Error(t, err, "stale instance should be gone")

	stillThere, err := a.GetInstance(fresh.ID, tApp.ID)
	require.NoError(t, err)
	assert.Equal(t, fresh.ID, stillThere.ID)

	history, err := a.GetInstanceStatusHistory(stale.ID, tApp.ID, tGroup.ID, 100)
	require.NoError(t, err)
	assert.Empty(t, history, "status history should cascade-delete with the instance")
}

func TestPruneStaleInstancesDryRunDoesNotDelete(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	_, rs, tApp, tGroup := setupPruneGroup(t, a, "prune_dry")

	stale, err := rs.RegisterInstance(types.Instance{ID: uuid.New().String(), IP: "10.0.0.3"}, NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
	require.NoError(t, err)
	backdateCheckIn(t, a, stale.ID, 2*time.Hour)

	result, err := rs.PruneStaleInstances(time.Now().UTC().Add(-time.Hour), 500, true)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Candidates)
	assert.Equal(t, int64(0), result.Deleted)

	_, err = a.GetInstance(stale.ID, tApp.ID)
	require.NoError(t, err, "dry-run must not delete the instance")
}

func TestPruneStaleInstancesKeepsInstanceWithAnyFreshCheckIn(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as, rs, tApp1, tGroup1 := setupPruneGroup(t, a, "prune_multi")
	tApp2, err := as.AddApp(&types.Application{Name: "prune_multi_app2", TeamID: tApp1.TeamID})
	require.NoError(t, err)
	tGroup2, err := as.AddGroup(&types.Group{Name: "prune_multi_g2", ApplicationID: tApp2.ID, PolicyUpdatesEnabled: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	require.NoError(t, err)

	id := uuid.New().String()
	_, err = rs.RegisterInstance(types.Instance{ID: id, IP: "10.0.0.4"}, NewInstanceApplication(tApp1.ID, tGroup1.ID, "1.0.0"))
	require.NoError(t, err)
	_, err = rs.RegisterInstance(types.Instance{ID: id, IP: "10.0.0.4"}, NewInstanceApplication(tApp2.ID, tGroup2.ID, "1.0.0"))
	require.NoError(t, err)

	_, err = dbconn.DB(a.Conn()).Exec(
		`UPDATE instance_application SET last_check_for_updates = $1 WHERE instance_id = $2 AND application_id = $3`,
		time.Now().UTC().Add(-24*time.Hour),
		id,
		tApp1.ID,
	)
	require.NoError(t, err)

	result, err := rs.PruneStaleInstances(time.Now().UTC().Add(-time.Hour), 500, false)
	require.NoError(t, err)

	_, err = a.GetInstance(id, tApp2.ID)
	require.NoError(t, err, "instance must stay if any application still checks in")
	assert.Equal(t, int64(0), result.Deleted)
}

func TestPruneStaleInstancesBatches(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	_, rs, tApp, tGroup := setupPruneGroup(t, a, "prune_batch")

	ids := make([]string, 3)
	for i := range ids {
		inst, err := rs.RegisterInstance(types.Instance{ID: uuid.New().String(), IP: "10.0.0.10"}, NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
		require.NoError(t, err)
		ids[i] = inst.ID
		backdateCheckIn(t, a, inst.ID, 2*time.Hour)
	}

	result, err := rs.PruneStaleInstances(time.Now().UTC().Add(-time.Hour), 2, false)
	require.NoError(t, err)
	assert.Equal(t, int64(3), result.Deleted)

	for _, id := range ids {
		_, err := a.GetInstance(id, tApp.ID)
		assert.Error(t, err, "batched prune should remove all stale instances")
	}
}

func TestPruneStaleInstancesRespectsMaxBatches(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	_, rs, tApp, tGroup := setupPruneGroup(t, a, "prune_cap")

	ids := make([]string, 3)
	for i := range ids {
		inst, err := rs.RegisterInstance(types.Instance{ID: uuid.New().String(), IP: "10.0.0.11"}, NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
		require.NoError(t, err)
		ids[i] = inst.ID
		backdateCheckIn(t, a, inst.ID, 2*time.Hour)
	}

	result, err := rs.pruneStaleInstances(time.Now().UTC().Add(-time.Hour), 1, 2, false)
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.Deleted)
	assert.True(t, result.MoreRemain)

	remaining := 0
	for _, id := range ids {
		if _, err := a.GetInstance(id, tApp.ID); err == nil {
			remaining++
		}
	}
	assert.Equal(t, 1, remaining)
}

func TestPruneStaleInstancesOrphanCreatedTs(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	rs := runtimeSvc(a)
	db := dbconn.DB(a.Conn())

	orphanID := uuid.New().String()
	_, err := db.Exec(`INSERT INTO instance (id, ip, created_ts) VALUES ($1, $2, $3)`, orphanID, "10.0.0.9", time.Now().UTC().Add(-3*time.Hour))
	require.NoError(t, err)

	result, err := rs.PruneStaleInstances(time.Now().UTC().Add(-time.Hour), 500, false)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Deleted)

	var n int
	err = db.QueryRowx(`SELECT COUNT(*) FROM instance WHERE id = $1`, orphanID).Scan(&n)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}

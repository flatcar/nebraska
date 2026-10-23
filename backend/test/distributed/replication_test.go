package distributed_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/flatcar/go-omaha/omaha"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplicatedDeletesClearEdgeRows(t *testing.T) {
	control := controlURL()
	edge := openDB(t, edgeDBURL())

	suffix := fmt.Sprintf("n%d", time.Now().UnixNano())
	app := control + "/api/apps"
	appID := createID(t, app, fmt.Sprintf(`{"name":"replicated-%[1]s","product_id":"io.test.replicated.%[1]s"}`, suffix))
	t.Cleanup(func() { request(t, "DELETE", app+"/"+appID, nil) })

	groups := fmt.Sprintf("%s/%s/groups", app, appID)
	group := func(name string) string {
		return createID(t, groups, fmt.Sprintf(`{"name":"%s-%s","policy_max_updates_per_period":1,`+
			`"policy_period_interval":"1 hours","policy_timezone":"UTC","policy_update_timeout":"1 days"}`, name, suffix))
	}
	deletedGroup, keptGroup := group("deleted"), group("kept")

	waitCount(t, edge, 2, "select count(*) from groups where application_id = $1", appID)

	// What the edge records while serving an instance in each group.
	for _, groupID := range []string{deletedGroup, keptGroup} {
		instanceID := "replicated-" + groupID
		for _, row := range []struct {
			stmt string
			args []any
		}{
			{"insert into instance (id, ip) values ($1, '10.0.0.1')", []any{instanceID}},
			{"insert into instance_application (instance_id, application_id, group_id, version, last_check_for_updates) values ($1, $2, $3, '1.0.0', now())", []any{instanceID, appID, groupID}},
			{"insert into activity (id, created_ts, class, severity, version, application_id, group_id, instance_id) values (gen_random_uuid(), now(), 1, 1, '1.0.0', $2, $3, $1)", []any{instanceID, appID, groupID}},
			{"insert into event (created_ts, instance_id, application_id, event_type_id) values (now(), $1, $2, 1)", []any{instanceID, appID}},
			{"insert into instance_status_history (status, version, created_ts, instance_id, application_id, group_id) values (1, '1.0.0', now(), $1, $2, $3)", []any{instanceID, appID, groupID}},
			// No group, so only the application's delete can remove these.
			{"insert into activity (id, created_ts, class, severity, version, application_id, instance_id) values (gen_random_uuid(), now(), 1, 1, '1.0.0', $2, $1)", []any{instanceID, appID}},
			{"insert into instance_status_history (status, version, created_ts, instance_id, application_id) values (1, '1.0.0', now(), $1, $2)", []any{instanceID, appID}},
		} {
			_, err := edge.Exec(row.stmt, row.args...)
			require.NoError(t, err, row.stmt)
		}
		t.Cleanup(func() { _, _ = edge.Exec("delete from instance where id = $1", instanceID) })
	}

	t.Run("group", func(t *testing.T) {
		code, body := request(t, "DELETE", groups+"/"+deletedGroup, nil)
		require.Equal(t, http.StatusNoContent, code, body)
		waitCount(t, edge, 0, "select count(*) from groups where id = $1", deletedGroup)

		for _, table := range []string{"group_local", "activity", "instance_status_history", "instance_application"} {
			assert.Zero(t, count(t, edge, "select count(*) from "+table+" where group_id = $1", deletedGroup), table)
		}
		// The foreign key sets the group to null rather than deleting the row.
		assert.Equal(t, 1, count(t, edge, "select count(*) from instance_application where instance_id = $1 and group_id is null", "replicated-"+deletedGroup))
		assert.Equal(t, 1, count(t, edge, "select count(*) from group_local where group_id = $1", keptGroup))
	})

	t.Run("application", func(t *testing.T) {
		code, body := request(t, "DELETE", app+"/"+appID, nil)
		require.Equal(t, http.StatusNoContent, code, body)
		waitCount(t, edge, 0, "select count(*) from application where id = $1", appID)

		for _, table := range []string{"activity", "event", "instance_application", "instance_status_history"} {
			assert.Zero(t, count(t, edge, "select count(*) from "+table+" where application_id = $1", appID), table)
		}
		// Control cascades to the groups and replicates each delete separately.
		assert.Zero(t, count(t, edge, "select count(*) from group_local where group_id = $1", keptGroup))
	})
}

// The edge must answer update checks for applications and groups that reach it
// from control after it has already answered others.
func TestEdgeServesReplicatedAppsAndGroups(t *testing.T) {
	control, edge := controlURL(), edgeURL()
	edgeDB := openDB(t, edgeDBURL())

	suffix := fmt.Sprintf("n%d", time.Now().UnixNano())
	t.Cleanup(func() { _, _ = edgeDB.Exec("delete from instance where id like $1", "lookup-"+suffix+"-%") })

	newGroup := func(t *testing.T, appID, channelID, track string) {
		t.Helper()
		createID(t, fmt.Sprintf("%s/api/apps/%s/groups", control, appID), fmt.Sprintf(`{"name":"%[1]s","track":"%[1]s","channel_id":"%[2]s",`+
			`"policy_updates_enabled":true,"policy_max_updates_per_period":100,"policy_period_interval":"1 hours",`+
			`"policy_update_timeout":"1 days","policy_timezone":"UTC","application_id":"%[3]s"}`, track, channelID, appID))
	}
	// newApp creates an application on control with a package, a channel and a
	// group on the "stable" track, and waits for them to reach the edge.
	newApp := func(t *testing.T, name string) (appID, channelID string) {
		t.Helper()
		appID = createID(t, control+"/api/apps", fmt.Sprintf(`{"name":"%[1]s-%[2]s","product_id":"io.test.%[1]s.%[2]s"}`, name, suffix))
		t.Cleanup(func() { request(t, "DELETE", control+"/api/apps/"+appID, nil) })
		base := control + "/api/apps/" + appID
		pkgID := createID(t, base+"/packages", fmt.Sprintf(`{"arch":1,"filename":"update.gz","description":"","url":"http://example.com/",`+
			`"version":"1.0.0","type":4,"size":"10","hash":"","application_id":"%s","channels_blacklist":[]}`, appID))
		channelID = createID(t, base+"/channels", fmt.Sprintf(`{"name":"stable","arch":1,"color":"","application_id":"%s","package_id":"%s"}`, appID, pkgID))
		newGroup(t, appID, channelID, "stable")
		waitCount(t, edgeDB, 1, "select count(*) from groups where application_id = $1", appID)
		return appID, channelID
	}
	machines := 0
	requireUpdate := func(t *testing.T, appID, track string) {
		t.Helper()
		machines++
		app := updateCheck(t, edge, appID, track, fmt.Sprintf("lookup-%s-%d", suffix, machines))
		assert.Equal(t, omaha.AppOK, app.Status)
		require.NotNil(t, app.UpdateCheck)
		assert.Equal(t, omaha.UpdateOK, app.UpdateCheck.Status)
	}

	firstApp, firstChannel := newApp(t, "first")
	requireUpdate(t, firstApp, "stable")

	t.Run("new_group", func(t *testing.T) {
		newGroup(t, firstApp, firstChannel, "beta")
		waitCount(t, edgeDB, 2, "select count(*) from groups where application_id = $1", firstApp)
		requireUpdate(t, firstApp, "beta")
	})

	t.Run("new_application", func(t *testing.T) {
		secondApp, _ := newApp(t, "second")
		requireUpdate(t, secondApp, "stable")
	})
}

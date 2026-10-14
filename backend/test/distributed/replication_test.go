package distributed_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

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

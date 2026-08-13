package dbreads

import (
	"database/sql"
	"fmt"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

var (
	// appInstancesPerChannelMetricSQL only counts instances that are still
	// considered "active" (i.e. within validityInterval of their last
	// check-in), matching the activity window used everywhere else
	// (GetApp, GetApps, GetGroupInstancesStats, GetGroupVersionBreakdown,
	// GetInstances). It uses LEFT JOINs for groups/channel so instances
	// whose group has no channel assigned (groups.channel_id can be NULL)
	// are still counted instead of being silently dropped, and it also
	// reports the channel's arch so that amd64/arm64 channels sharing the
	// same name are not folded into a single row.
	appInstancesPerChannelMetricSQL = fmt.Sprintf(`
SELECT a.name AS app_name, ia.version AS version, COALESCE(c.name, '') AS channel_name, COALESCE(c.arch, -1) AS arch, count(ia.version) AS instances_count
FROM instance_application ia
JOIN application a ON a.id = ia.application_id
LEFT JOIN groups g ON ia.group_id = g.id
LEFT JOIN channel c ON g.channel_id = c.id
WHERE ia.last_check_for_updates > now() at time zone 'utc' - interval '%[1]s' AND %[2]s
GROUP BY app_name, version, channel_name, arch
ORDER BY app_name, version, channel_name, arch
`, validityInterval, ignoreFakeInstanceCondition("ia.instance_id"))

	failedUpdatesSQL = fmt.Sprintf(`
SELECT a.name AS app_name, count(*) as fail_count
FROM application a, event e, event_type et
WHERE a.id = e.application_id AND e.event_type_id = et.id AND et.result = 0 AND et.type = 3 AND %s
GROUP BY app_name
ORDER BY app_name
`, ignoreFakeInstanceCondition("e.instance_id"))
)

func (q *Queries) GetAppInstancesPerChannelMetrics() ([]types.AppInstancesPerChannelMetric, error) {
	var metrics []types.AppInstancesPerChannelMetric
	rows, err := q.db.Queryx(appInstancesPerChannelMetricSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var metric types.AppInstancesPerChannelMetric
		err := rows.StructScan(&metric)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return metrics, nil
}

func (q *Queries) GetFailedUpdatesMetrics() ([]types.FailedUpdatesMetric, error) {
	var metrics []types.FailedUpdatesMetric
	rows, err := q.db.Queryx(failedUpdatesSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var metric types.FailedUpdatesMetric
		err := rows.StructScan(&metric)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return metrics, nil
}

func (q *Queries) DbStats() sql.DBStats {
	return q.db.Stats()
}

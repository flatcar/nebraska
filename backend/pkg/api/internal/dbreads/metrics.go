package dbreads

import (
	"database/sql"
	"fmt"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/types"
)

var (
	appInstancesPerChannelMetricSQL = fmt.Sprintf(`
SELECT a.name AS app_name, ia.version AS version, c.name AS channel_name, count(ia.version) AS instances_count
FROM instance_application ia, application a, channel c, groups g
WHERE a.id = ia.application_id AND ia.group_id = g.id AND g.channel_id = c.id AND %s
GROUP BY app_name, version, channel_name
ORDER BY app_name, version, channel_name
`, ignoreFakeInstanceCondition("ia.instance_id"))

	failedUpdatesSQL = fmt.Sprintf(`
SELECT a.name AS app_name, count(*) as fail_count
FROM application a, event e, event_type et
WHERE a.id = e.application_id AND e.event_type_id = et.id AND et.result = 0 AND et.type = 3 AND %s
GROUP BY app_name
ORDER BY app_name
`, ignoreFakeInstanceCondition("e.instance_id"))

	// instancesPerOEMMetricSQL counts active instances grouped by application
	// and OEM platform. It joins instance (for the oem column) with
	// instance_application and application, and excludes synthetic test
	// instances using the same fake-instance filter applied elsewhere.
	instancesPerOEMMetricSQL = fmt.Sprintf(`
SELECT a.name AS app_name, i.oem AS oem, count(*) AS instances_count
FROM instance i, instance_application ia, application a
WHERE i.id = ia.instance_id AND a.id = ia.application_id AND %s
GROUP BY app_name, oem
ORDER BY app_name, oem
`, ignoreFakeInstanceCondition("i.id"))
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

// GetInstancesPerOEMMetrics returns the count of active instances for each
// (application, OEM platform) pair. Instances with an empty OEM string are
// included under the empty-string key so operators can see how many instances
// have not yet reported their platform.
func (q *Queries) GetInstancesPerOEMMetrics() ([]types.InstancesPerOEMMetric, error) {
	var metrics []types.InstancesPerOEMMetric
	rows, err := q.db.Queryx(instancesPerOEMMetricSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var metric types.InstancesPerOEMMetric
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

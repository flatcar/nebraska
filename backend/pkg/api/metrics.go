package api

import (
	"database/sql"
	"fmt"
)

var (
	appInstancesPerChannelMetricSQL string = fmt.Sprintf(`
SELECT a.name AS app_name, ia.version AS version, c.name AS channel_name, count(ia.version) AS instances_count
FROM instance_application ia, application a, channel c, groups g
WHERE a.id = ia.application_id AND ia.group_id = g.id AND g.channel_id = c.id AND %s
GROUP BY app_name, version, channel_name
ORDER BY app_name, version, channel_name
`, ignoreFakeInstanceCondition("ia.instance_id"))

	failedUpdatesSQL string = fmt.Sprintf(`
SELECT a.name AS app_name, count(*) as fail_count
FROM application a, event e, event_type et
WHERE a.id = e.application_id AND e.event_type_id = et.id AND et.result = 0 AND et.type = 3 AND %s
GROUP BY app_name
ORDER BY app_name
`, ignoreFakeInstanceCondition("e.instance_id"))
)

type AppInstancesPerChannelMetric struct {
	ApplicationName string `db:"app_name" json:"app_name"`
	Version         string `db:"version" json:"version"`
	ChannelName     string `db:"channel_name" json:"channel_name"`
	InstancesCount  int    `db:"instances_count" json:"instances_count"`
}

func (api *API) GetAppInstancesPerChannelMetrics() ([]AppInstancesPerChannelMetric, error) {
	var metrics []AppInstancesPerChannelMetric
	rows, err := api.db.Queryx(appInstancesPerChannelMetricSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var metric AppInstancesPerChannelMetric
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

type FailedUpdatesMetric struct {
	ApplicationName string `db:"app_name" json:"app_name"`
	FailureCount    int    `db:"fail_count" json:"fail_count"`
}

func (api *API) GetFailedUpdatesMetrics() ([]FailedUpdatesMetric, error) {
	var metrics []FailedUpdatesMetric
	rows, err := api.db.Queryx(failedUpdatesSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var metric FailedUpdatesMetric
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

func (api *API) DbStats() sql.DBStats {
	return api.db.Stats()
}

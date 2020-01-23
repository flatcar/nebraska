package api

import (
	"database/sql"
	"fmt"
)

var (
	appInstancesPerChannelMetricSQL string = fmt.Sprintf(`
SELECT a.name AS app_name, ia.version AS version, c.name AS channel_name, count(ia.version) AS instances_count
FROM instance_application ia, application a, channel c, groups g
WHERE a.team_id = $1 AND a.id = ia.application_id AND ia.group_id = g.id AND g.channel_id = c.id AND %s
GROUP BY app_name, version, channel_name
ORDER BY app_name, version, channel_name
`, ignoreFakeInstanceCondition("ia.instance_id"))

	failedUpdatesSQL string = fmt.Sprintf(`
SELECT a.name AS app_name, count(*) as fail_count
FROM application a, event e, event_type et
WHERE a.team_id = $1 AND a.id = e.application_id AND e.event_type_id = et.id AND et.result = 0 AND et.type = 3 AND %s
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

func (api *API) GetAppInstancesPerChannelMetrics(teamID string) ([]AppInstancesPerChannelMetric, error) {
	var metrics []AppInstancesPerChannelMetric

	switch err := api.dbR.SQL(appInstancesPerChannelMetricSQL, teamID).QueryStructs(&metrics); err {
	case nil:
		return metrics, nil
	case sql.ErrNoRows:
		return metrics, nil
	default:
		return nil, err
	}
}

type FailedUpdatesMetric struct {
	ApplicationName string `db:"app_name" json:"app_name"`
	FailureCount    int    `db:"fail_count" json:"fail_count"`
}

func (api *API) GetFailedUpdatesMetrics(teamID string) ([]FailedUpdatesMetric, error) {
	var metrics []FailedUpdatesMetric

	switch err := api.dbR.SQL(failedUpdatesSQL, teamID).QueryStructs(&metrics); err {
	case nil:
		return metrics, nil
	case sql.ErrNoRows:
		return metrics, nil
	default:
		return nil, err
	}
}

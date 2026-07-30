package api

import (
	"github.com/doug-martin/goqu/v9"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/types"
)

const (
	activityPackageNotFound       = types.ActivityPackageNotFound
	activityRolloutStarted        = types.ActivityRolloutStarted
	activityRolloutFinished       = types.ActivityRolloutFinished
	activityRolloutFailed         = types.ActivityRolloutFailed
	activityInstanceUpdateFailed  = types.ActivityInstanceUpdateFailed
	activityChannelPackageUpdated = types.ActivityChannelPackageUpdated
)

const (
	activitySuccess = types.ActivitySuccess
	activityInfo    = types.ActivityInfo
	activityWarning = types.ActivityWarning
	activityError   = types.ActivityError
)

type (
	Activity            = types.Activity
	ActivityQueryParams = types.ActivityQueryParams
)

// newGroupActivityEntry creates a new activity entry related to a specific
// group.
func (api *API) newGroupActivityEntry(class int, severity int, version, appID, groupID string) error {
	query, _, err := goqu.Insert("activity").
		Cols("class", "severity", "version", "application_id", "group_id").
		Vals(goqu.Vals{class, severity, version, appID, groupID}).
		ToSQL()
	if err != nil {
		return err
	}
	_, err = api.db.Exec(query)

	if err != nil {
		return err
	}

	return nil
}

// newInstanceActivityEntry creates a new activity entry related to a specific
// instance.
func (api *API) newInstanceActivityEntry(class int, severity int, version, appID, groupID, instanceID string) error {
	query, _, err := goqu.Insert("activity").
		Cols("class", "severity", "version", "application_id", "group_id", "instance_id").
		Vals(goqu.Vals{class, severity, version, appID, groupID, instanceID}).
		ToSQL()
	if err != nil {
		return err
	}
	_, err = api.db.Exec(query)

	if err != nil {
		return err
	}

	return nil
}

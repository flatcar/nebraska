package api

import (
	"github.com/doug-martin/goqu/v9"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/types"
)

var (
	// ErrInvalidChannel error indicates that a channel doesn't belong to the
	// application it was supposed to belong to.
	ErrInvalidChannel = types.ErrInvalidChannel

	// ErrExpectingValidTimezone error indicates that a valid timezone wasn't
	// provided when enabling the flag PolicyOfficeHours.
	ErrExpectingValidTimezone = types.ErrExpectingValidTimezone
)

type (
	GroupDescriptor                 = types.GroupDescriptor
	Group                           = types.Group
	VersionBreakdownEntry           = types.VersionBreakdownEntry
	VersionCountTimelineEntry       = types.VersionCountTimelineEntry
	StatusVersionCountTimelineEntry = types.StatusVersionCountTimelineEntry
	VersionCountMap                 = types.VersionCountMap
	InstancesStatusStats            = types.InstancesStatusStats
	UpdatesStats                    = types.UpdatesStats
)

// ClearUpdatesEnabledOverride clears the local policy_updates_enabled override
// on the group_local row so the admin default on groups takes effect again on
// this node.
func (api *API) ClearUpdatesEnabledOverride(groupID string) error {
	query, _, err := goqu.Update("group_local").
		Set(goqu.Record{"policy_updates_enabled_override": nil}).
		Where(goqu.C("group_id").Eq(groupID)).
		ToSQL()
	if err != nil {
		return err
	}
	result, err := api.db.Exec(query)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNoRowsAffected
	}
	return nil
}

// disableUpdates trips the safe-mode brake by setting the
// policy_updates_enabled override on the group_local row. The override
// lives on the node-local table, so it stops update grants on this node only.
func (api *API) disableUpdates(groupID string) error {
	query, _, err := goqu.Update("group_local").
		Set(goqu.Record{"policy_updates_enabled_override": false}).
		Where(goqu.C("group_id").Eq(groupID)).
		ToSQL()
	if err != nil {
		return err
	}
	_, err = api.db.Exec(query)

	return err
}

// setGroupRolloutInProgress updates the value of the rollout_in_progress flag
// for a given group, indicating if a rollout is taking place now or not.
func (api *API) setGroupRolloutInProgress(groupID string, inProgress bool) error {
	query, _, err := goqu.Update("group_local").
		Set(goqu.Record{"rollout_in_progress": inProgress}).
		Where(goqu.C("group_id").Eq(groupID)).
		ToSQL()
	if err != nil {
		return err
	}
	_, err = api.db.Exec(query)

	return err
}

package runtime

import (
	"github.com/doug-martin/goqu/v9"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

// ClearUpdatesEnabledOverride clears the local policy_updates_enabled override
// on the group_local row so the admin default on groups takes effect again on
// this node.
func (s *Service) ClearUpdatesEnabledOverride(groupID string) error {
	query, _, err := goqu.Update("group_local").
		Set(goqu.Record{"policy_updates_enabled_override": nil}).
		Where(goqu.C("group_id").Eq(groupID)).
		ToSQL()
	if err != nil {
		return err
	}
	result, err := s.db.Exec(query)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return types.ErrNoRowsAffected
	}
	return nil
}

// disableUpdates trips the safe-mode brake by setting the
// policy_updates_enabled override on the group_local row. The override
// lives on the node-local table, so it stops update grants on this node only.
func (s *Service) disableUpdates(groupID string) error {
	query, _, err := goqu.Update("group_local").
		Set(goqu.Record{"policy_updates_enabled_override": false}).
		Where(goqu.C("group_id").Eq(groupID)).
		ToSQL()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(query)

	return err
}

// setGroupRolloutInProgress updates the value of the rollout_in_progress flag
// for a given group, indicating if a rollout is taking place now or not.
func (s *Service) setGroupRolloutInProgress(groupID string, inProgress bool) error {
	query, _, err := goqu.Update("group_local").
		Set(goqu.Record{"rollout_in_progress": inProgress}).
		Where(goqu.C("group_id").Eq(groupID)).
		ToSQL()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(query)

	return err
}

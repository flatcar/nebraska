package runtime

import (
	"database/sql"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

// UpdateLocalOverrides updates this node's overrides of the group's policy.
// Overrides the patch leaves out are kept.
func (s *Service) UpdateLocalOverrides(groupID string, patch *types.GroupPolicyOverridesPatch) error {
	if err := patch.Validate(); err != nil {
		return err
	}

	query, _, err := goqu.From("group_local").
		Select("policy_office_hours_override", "policy_timezone_override").
		Where(goqu.C("group_id").Eq(groupID)).
		ToSQL()
	if err != nil {
		return err
	}
	var officeHours struct {
		Enabled  null.Bool   `db:"policy_office_hours_override"`
		Timezone null.String `db:"policy_timezone_override"`
	}
	if err := s.db.QueryRowx(query).StructScan(&officeHours); err != nil {
		if err == sql.ErrNoRows {
			return types.ErrNoRowsAffected
		}
		return err
	}
	// Office hours and timezone are overridden together. Only a patch that
	// changes one of them is checked, so a pause never fails on their account.
	if patch.PolicyOfficeHours != nil || patch.PolicyTimezone != nil {
		if patch.PolicyOfficeHours != nil {
			officeHours.Enabled = *patch.PolicyOfficeHours
		}
		if patch.PolicyTimezone != nil {
			officeHours.Timezone = *patch.PolicyTimezone
		}
		switch {
		case officeHours.Enabled.Valid && officeHours.Enabled.Bool && !officeHours.Timezone.Valid:
			return fmt.Errorf("%w: an office hours override needs a policy_timezone override", types.ErrInvalidLocalOverrides)
		case officeHours.Timezone.Valid && !officeHours.Enabled.Valid:
			return fmt.Errorf("%w: a policy_timezone override needs a policy_office_hours override", types.ErrInvalidLocalOverrides)
		}
	}

	if changes := patch.Changes(); len(changes) > 0 {
		record := goqu.Record{}
		for name, value := range changes {
			record[name+"_override"] = value
		}
		query, _, err := goqu.Update("group_local").
			Set(record).
			Where(goqu.C("group_id").Eq(groupID)).
			ToSQL()
		if err != nil {
			return err
		}
		if _, err := s.db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

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

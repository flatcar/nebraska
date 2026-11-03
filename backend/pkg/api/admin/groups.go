package admin

import (
	"database/sql"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

// addGroup validates and inserts the group using the transaction provided.
// The caller owns the transaction and is responsible for committing it and
// for invalidating the group cache afterwards.
func (s *Service) addGroup(group *types.Group, tx *sqlx.Tx) error {
	if group.PolicyOfficeHours && !isTimezoneValid(group.PolicyTimezone.String) {
		return types.ErrExpectingValidTimezone
	}

	if group.ChannelID.String != "" {
		// Validated through the transaction: when cloning an application the
		// channel this group points at was inserted by the same transaction
		// and isn't visible to reads made outside of it yet.
		if err := validateChannel(tx, group.ChannelID.String, group.ApplicationID); err != nil {
			return err
		}
	}
	// Instead of trying to solve this in the database, generate the ID beforehand to copy it to the track.
	if group.ID == "" {
		group.ID = uuid.New().String()
	}
	if group.Track == "" {
		group.Track = group.ID
	}
	query, _, err := goqu.Insert("groups").
		Cols("id", "name", "description", "application_id", "channel_id", "policy_updates_enabled", "policy_safe_mode", "policy_office_hours",
			"policy_timezone", "policy_period_interval", "policy_max_updates_per_period", "policy_update_timeout", "track").
		Vals(goqu.Vals{
			group.ID,
			group.Name,
			group.Description,
			group.ApplicationID,
			group.ChannelID,
			group.PolicyUpdatesEnabled,
			group.PolicySafeMode,
			group.PolicyOfficeHours,
			group.PolicyTimezone,
			group.PolicyPeriodInterval,
			group.PolicyMaxUpdatesPerPeriod,
			group.PolicyUpdateTimeout,
			group.Track,
		}).
		Returning(goqu.T("groups").All()).
		ToSQL()
	if err != nil {
		return err
	}
	return tx.QueryRowx(query).StructScan(group)
}

// AddGroup registers the provided group.
func (s *Service) AddGroup(group *types.Group) (*types.Group, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			l.Error().Err(err).Msg("AddGroup - could not roll back")
		}
	}()

	if err := s.addGroup(group, tx); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	s.UpdateCachedGroups()
	// Re-read through groupsQuery so the returned struct reflects the joined
	// group_local row.
	return s.GetGroup(group.ID)
}

// UpdateGroup updates an existing group using the context of the group
// provided.
func (s *Service) UpdateGroup(group *types.Group) error {
	if group.PolicyOfficeHours && !isTimezoneValid(group.PolicyTimezone.String) {
		return types.ErrExpectingValidTimezone
	}

	groupBeforeUpdate, err := s.GetGroup(group.ID)
	if err != nil {
		return err
	}

	if group.ChannelID.String != "" {
		if err := validateChannel(s.db, group.ChannelID.String, groupBeforeUpdate.ApplicationID); err != nil {
			return err
		}
	}
	if group.Track == "" {
		group.Track = group.ID
	}
	query, _, err := goqu.Update("groups").
		Set(
			goqu.Record{
				"name":                          group.Name,
				"description":                   group.Description,
				"channel_id":                    group.ChannelID,
				"policy_updates_enabled":        group.PolicyUpdatesEnabled,
				"policy_safe_mode":              group.PolicySafeMode,
				"policy_office_hours":           group.PolicyOfficeHours,
				"policy_timezone":               group.PolicyTimezone,
				"policy_period_interval":        group.PolicyPeriodInterval,
				"policy_max_updates_per_period": group.PolicyMaxUpdatesPerPeriod,
				"policy_update_timeout":         group.PolicyUpdateTimeout,
				"track":                         group.Track,
			},
		).
		Where(goqu.C("id").Eq(group.ID)).
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
	s.UpdateCachedGroups()
	return nil
}

// DeleteGroup removes the group identified by the id provided.
func (s *Service) DeleteGroup(groupID string) error {
	query, _, err := goqu.Delete("groups").Where(goqu.C("id").Eq(groupID)).ToSQL()
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
	s.UpdateCachedGroups()
	return nil
}

// validateChannel checks if a channel belongs to the application provided.
// It takes the queryer to run against so that it can be called either on the
// shared connection or on an open transaction, the latter being required when
// the channel was inserted by that same, not yet committed, transaction.
func validateChannel(q sqlx.Queryer, channelID, appID string) error {
	query, _, err := goqu.From("channel").
		Select("application_id").
		Where(goqu.C("id").Eq(channelID)).
		ToSQL()
	if err != nil {
		return err
	}
	var channelAppID string
	if err := q.QueryRowx(query).Scan(&channelAppID); err != nil {
		return err
	}
	if channelAppID != appID {
		return types.ErrInvalidChannel
	}
	return nil
}

// isTimezoneValid checks if the provided timezone is valid.
func isTimezoneValid(tz string) bool {
	if tz == "" {
		return false
	}

	if _, err := time.LoadLocation(tz); err != nil {
		return false
	}

	return true
}

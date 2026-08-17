package admin

import (
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

// AddGroup registers the provided group.
func (s *Service) AddGroup(group *types.Group) (*types.Group, error) {
	if group.PolicyOfficeHours && !isTimezoneValid(group.PolicyTimezone.String) {
		return nil, types.ErrExpectingValidTimezone
	}

	if group.ChannelID.String != "" {
		if err := s.validateChannel(group.ChannelID.String, group.ApplicationID); err != nil {
			return nil, err
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
		return nil, err
	}
	err = s.db.QueryRowx(query).StructScan(group)
	if err != nil {
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
		if err := s.validateChannel(group.ChannelID.String, groupBeforeUpdate.ApplicationID); err != nil {
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
func (s *Service) validateChannel(channelID, appID string) error {
	channel, err := s.GetChannel(channelID)
	if err != nil {
		return err
	}
	if channel.ApplicationID != appID {
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

package api

import (
	"errors"
	"fmt"
	"time"

	"gopkg.in/mgutz/dat.v1"
)

var (
	// ErrInvalidChannel error indicates that a channel doesn't belong to the
	// application it was supposed to belong to.
	ErrInvalidChannel = errors.New("nebraska: invalid channel")

	// ErrExpectingValidTimezone error indicates that a valid timezone wasn't
	// provided when enabling the flag PolicyOfficeHours.
	ErrExpectingValidTimezone = errors.New("nebraska: expecting valid timezone")
)

// Group represents a Nebraska application's group.
type Group struct {
	ID                        string                   `db:"id" json:"id"`
	Name                      string                   `db:"name" json:"name"`
	Description               string                   `db:"description" json:"description"`
	CreatedTs                 time.Time                `db:"created_ts" json:"created_ts"`
	RolloutInProgress         bool                     `db:"rollout_in_progress" json:"rollout_in_progress"`
	ApplicationID             string                   `db:"application_id" json:"application_id"`
	ChannelID                 dat.NullString           `db:"channel_id" json:"channel_id"`
	PolicyUpdatesEnabled      bool                     `db:"policy_updates_enabled" json:"policy_updates_enabled"`
	PolicySafeMode            bool                     `db:"policy_safe_mode" json:"policy_safe_mode"`
	PolicyOfficeHours         bool                     `db:"policy_office_hours" json:"policy_office_hours"`
	PolicyTimezone            dat.NullString           `db:"policy_timezone" json:"policy_timezone"`
	PolicyPeriodInterval      string                   `db:"policy_period_interval" json:"policy_period_interval"`
	PolicyMaxUpdatesPerPeriod int                      `db:"policy_max_updates_per_period" json:"policy_max_updates_per_period"`
	PolicyUpdateTimeout       string                   `db:"policy_update_timeout" json:"policy_update_timeout"`
	VersionBreakdown          []*VersionBreakdownEntry `db:"version_breakdown" json:"version_breakdown,omitempty"`
	Channel                   *Channel                 `db:"channel" json:"channel,omitempty"`
	InstancesStats            InstancesStatusStats     `db:"instances_stats" json:"instances_stats,omitempty"`
}

// VersionBreakdownEntry represents the distribution of the versions currently
// installed in the instances belonging to a given group.
type VersionBreakdownEntry struct {
	Version    string  `db:"version" json:"version"`
	Instances  int     `db:"instances" json:"instances"`
	Percentage float64 `db:"percentage" json:"percentage"`
}

type VersionCountTimelineEntry struct {
	Time    time.Time `db:"ts" json:"time"`
	Version string    `db:"version" json:"version"`
	Total   uint64    `db:"total" json:"total"`
}

type StatusVersionCountTimelineEntry struct {
	Time    time.Time `db:"ts" json:"time"`
	Status  int       `db:"status" json:"status"`
	Version string    `db:"version" json:"version"`
	Total   uint64    `db:"total" json:"total"`
}

type VersionCountMap = map[string]uint64

// InstancesStatusStats represents a set of statistics about the status of the
// instances that belong to a given group.
type InstancesStatusStats struct {
	Total         int `db:"total" json:"total"`
	Undefined     int `db:"undefined" json:"undefined"`
	UpdateGranted int `db:"update_granted" json:"update_granted"`
	Error         int `db:"error" json:"error"`
	Complete      int `db:"complete" json:"complete"`
	Installed     int `db:"installed" json:"installed"`
	Downloaded    int `db:"downloaded" json:"downloaded"`
	Downloading   int `db:"downloading" json:"downloading"`
	OnHold        int `db:"onhold" json:"onhold"`
}

// UpdatesStats represents a set of statistics about the status of the updates
// that may be taking place in the instaces belonging to a given group.
type UpdatesStats struct {
	TotalInstances                   int `db:"total_instances"`
	UpdatesToCurrentVersionGranted   int `db:"updates_to_current_version_granted"`
	UpdatesToCurrentVersionAttempted int `db:"updates_to_current_version_attempted"`
	UpdatesToCurrentVersionSucceeded int `db:"updates_to_current_version_succeeded"`
	UpdatesToCurrentVersionFailed    int `db:"updates_to_current_version_failed"`
	UpdatesGrantedInLastPeriod       int `db:"updates_granted_in_last_period"`
	UpdatesInProgress                int `db:"updates_in_progress"`
	UpdatesTimedOut                  int `db:"updates_timed_out"`
}

// AddGroup registers the provided group.
func (api *API) AddGroup(group *Group) (*Group, error) {
	if group.PolicyOfficeHours && !isTimezoneValid(group.PolicyTimezone.String) {
		return nil, ErrExpectingValidTimezone
	}

	if group.ChannelID.String != "" {
		if err := api.validateChannel(group.ChannelID.String, group.ApplicationID); err != nil {
			return nil, err
		}
	}

	err := api.dbR.
		InsertInto("groups").
		Whitelist("name", "description", "application_id", "channel_id", "policy_updates_enabled", "policy_safe_mode", "policy_office_hours",
			"policy_timezone", "policy_period_interval", "policy_max_updates_per_period", "policy_update_timeout").
		Record(group).
		Returning("*").
		QueryStruct(group)

	return group, err
}

// UpdateGroup updates an existing group using the context of the group
// provided.
func (api *API) UpdateGroup(group *Group) error {
	if group.PolicyOfficeHours && !isTimezoneValid(group.PolicyTimezone.String) {
		return ErrExpectingValidTimezone
	}

	groupBeforeUpdate, err := api.GetGroup(group.ID)
	if err != nil {
		return err
	}

	if group.ChannelID.String != "" {
		if err := api.validateChannel(group.ChannelID.String, groupBeforeUpdate.ApplicationID); err != nil {
			return err
		}
	}

	result, err := api.dbR.
		Update("groups").
		SetWhitelist(group, "name", "description", "channel_id", "policy_updates_enabled", "policy_safe_mode", "policy_office_hours",
			"policy_timezone", "policy_period_interval", "policy_max_updates_per_period", "policy_update_timeout").
		Where("id = $1", group.ID).
		Exec()

	if err == nil && result.RowsAffected == 0 {
		return ErrNoRowsAffected
	}

	return err
}

// DeleteGroup removes the group identified by the id provided.
func (api *API) DeleteGroup(groupID string) error {
	result, err := api.dbR.
		DeleteFrom("groups").
		Where("id = $1", groupID).
		Exec()

	if err == nil && result.RowsAffected == 0 {
		return ErrNoRowsAffected
	}

	return err
}

// GetGroup returns the group identified by the id provided.
func (api *API) GetGroup(groupID string) (*Group, error) {
	var group Group

	err := api.groupsQuery().
		Where("id = $1", groupID).
		QueryStruct(&group)

	if err != nil {
		return nil, err
	}

	return &group, nil
}

// GetGroups returns all groups that belong to the application provided.
func (api *API) GetGroups(appID string, page, perPage uint64) ([]*Group, error) {
	page, perPage = validatePaginationParams(page, perPage)

	var groups []*Group

	err := api.groupsQuery().
		Where("application_id = $1", appID).
		Paginate(page, perPage).
		QueryStructs(&groups)

	return groups, err
}

// validateChannel checks if a channel belongs to the application provided.
func (api *API) validateChannel(channelID, appID string) error {
	channel, err := api.GetChannel(channelID)
	if err == nil {
		if channel.ApplicationID != appID {
			return ErrInvalidChannel
		}
	}

	return nil
}

// getGroupUpdatesStats returns a set of statistics about the distribution of
// updates and their status in the group provided.
func (api *API) getGroupUpdatesStats(group *Group) (*UpdatesStats, error) {
	var updatesStats UpdatesStats

	packageVersion := ""
	if group.Channel.Package != nil {
		packageVersion = group.Channel.Package.Version
	}

	query := fmt.Sprintf(`
	SELECT
		count(*) total_instances,
		sum(case when last_update_version = $1 then 1 else 0 end) updates_to_current_version_granted,
		sum(case when update_in_progress = 'false' and last_update_version = $1 then 1 else 0 end) updates_to_current_version_attempted,
		sum(case when update_in_progress = 'false' and last_update_version = $1 and last_update_version = version then 1 else 0 end) updates_to_current_version_succeeded,
		sum(case when update_in_progress = 'false' and last_update_version = $1 and last_update_version != version then 1 else 0 end) updates_to_current_version_failed,
		sum(case when last_update_granted_ts > now() at time zone 'utc' - interval $2 then 1 else 0 end) updates_granted_in_last_period,
		sum(case when update_in_progress = 'true' and now() at time zone 'utc' - last_update_granted_ts <= interval $3 then 1 else 0 end) updates_in_progress,
		sum(case when update_in_progress = 'true' and now() at time zone 'utc' - last_update_granted_ts > interval $4 then 1 else 0 end) updates_timed_out
	FROM instance_application
	WHERE group_id=$5 AND last_check_for_updates > now() at time zone 'utc' - interval '%s'
	`, validityInterval)

	err := api.dbR.SQL(query, packageVersion, group.PolicyPeriodInterval, group.PolicyUpdateTimeout, group.PolicyUpdateTimeout, group.ID).
		QueryStruct(&updatesStats)
	if err != nil {
		return nil, err
	}

	return &updatesStats, nil
}

// disableUpdates updates the group provided setting the policy_updates_enabled
// field to false. This usually happens when the first instance in a group
// processing an update to a specific version fails if safe mode is enabled.
func (api *API) disableUpdates(groupID string) error {
	_, err := api.dbR.
		Update("groups").
		Set("policy_updates_enabled", false).
		Where("id = $1", groupID).
		Exec()

	return err
}

// setGroupRolloutInProgress updates the value of the rollout_in_progress flag
// for a given group, indicating if a rollout is taking place now or not.
func (api *API) setGroupRolloutInProgress(groupID string, inProgress bool) error {
	_, err := api.dbR.
		Update("groups").
		Set("rollout_in_progress", inProgress).
		Where("id = $1", groupID).
		Exec()

	return err
}

// groupsQuery returns a SelectDocBuilder prepared to return all groups. This
// query is meant to be extended later in the methods using it to filter by a
// specific group id, all groups of a given app, specify how to query the rows
// or their destination.
func (api *API) groupsQuery() *dat.SelectDocBuilder {
	return api.dbR.
		SelectDoc("*").
		One("instances_stats", api.groupInstancesStatusQuery()).
		One("channel", api.channelsQuery().Where("id = groups.channel_id")).
		Many("version_breakdown", api.groupVersionBreakdownQuery()).
		From("groups").
		OrderBy("created_ts DESC")
}

// groupVersionBreakdownQuery returns a SQL query prepared to return the version
// breakdown of all instances running on a given group.
func (api *API) groupVersionBreakdownQuery() string {
	return fmt.Sprintf(`
	SELECT version, count(*) as instances, (count(*) * 100.0 / total) as percentage
	FROM instance_application, (
		SELECT count(*) as total 
		FROM instance_application 
		WHERE group_id=groups.id AND last_check_for_updates > now() at time zone 'utc' - interval '%s'
		) totals
	WHERE group_id=groups.id AND last_check_for_updates > now() at time zone 'utc' - interval '%s'
	GROUP BY version, total
	ORDER BY regexp_matches(version, '(\d+)\.(\d+)\.(\d+)')::int[] DESC
	`, validityInterval, validityInterval)
}

// groupInstancesStatusQuery returns a SQL query prepared to return a summary
// of the status of the instances that belong to a given group.
func (api *API) groupInstancesStatusQuery() string {
	return fmt.Sprintf(`
	SELECT
		count(*) total,
		sum(case when status IS NULL then 1 else 0 end) undefined,
		sum(case when status = %d then 1 else 0 end) error,
		sum(case when status = %d then 1 else 0 end) update_granted,
		sum(case when status = %d then 1 else 0 end) complete,
		sum(case when status = %d then 1 else 0 end) installed,
		sum(case when status = %d then 1 else 0 end) downloaded,
		sum(case when status = %d then 1 else 0 end) downloading,
		sum(case when status = %d then 1 else 0 end) onhold
	FROM instance_application
	WHERE group_id=groups.id AND last_check_for_updates > now() at time zone 'utc' - interval '%s'`,
		InstanceStatusError, InstanceStatusUpdateGranted, InstanceStatusComplete, InstanceStatusInstalled,
		InstanceStatusDownloaded, InstanceStatusDownloading, InstanceStatusOnHold, validityInterval)
}

func (api *API) GetGroupVersionCountTimeline(groupID string) (map[time.Time](VersionCountMap), error) {
	var timelineEntry []VersionCountTimelineEntry
	// Get the number of instances per version until each of the time-interval
	// divisions. This is done only for the instances that pinged the server in
	// the last time-interval.
	query := fmt.Sprintf(`
	WITH time_series AS (SELECT * FROM generate_series(now() - interval '%[1]s', now(), INTERVAL '1 hour') AS ts),
		 recent_instances AS (SELECT instance_id, (CASE WHEN last_update_granted_ts IS NOT NULL THEN last_update_granted_ts ELSE created_ts END), version, 4 status FROM instance_application WHERE group_id=$1 AND last_check_for_updates >= now() - interval '%[1]s' ORDER BY last_update_granted_ts DESC),
		 instance_versions AS (SELECT instance_id, created_ts, version, status FROM instance_status_history WHERE instance_id IN (SELECT instance_id FROM recent_instances) AND status = 4 UNION (SELECT * FROM recent_instances) ORDER BY created_ts DESC)
	SELECT ts, (CASE WHEN version IS NULL THEN '' ELSE version END), sum(CASE WHEN version IS NOT null THEN 1 ELSE 0 END) total FROM (SELECT * FROM time_series LEFT JOIN LATERAL(SELECT distinct ON (instance_id) instance_Id, version, created_ts FROM instance_versions WHERE created_ts <= time_series.ts ORDER BY instance_Id, created_ts DESC) _ ON true) AS _
	GROUP BY 1,2
	ORDER BY ts DESC;
	`, validityInterval)

	if err := api.dbR.SQL(query, groupID).QueryStructs(&timelineEntry); err != nil {
		return nil, err
	}

	allVersions := make(map[string]struct{})
	timelineCount := make(map[time.Time]VersionCountMap)

	// Create the timeline map, and gather all the versions found.
	for _, entry := range timelineEntry {
		value, ok := timelineCount[entry.Time]
		if !ok {
			value = make(VersionCountMap)
			timelineCount[entry.Time] = value
		}

		allVersions[entry.Version] = struct{}{}
		versionCount, ok := value[entry.Version]
		if !ok {
			versionCount = entry.Total
		}

		value[entry.Version] = versionCount
	}

	// We want to return all the versions count per time-interval, i.e. we
	// don't want some time-intervals to have 3 versions accounted, and others
	// just 1, so this assigns the missing versions per interval.
	for version := range allVersions {
		for timestamp := range timelineCount {
			if _, ok := timelineCount[timestamp][version]; !ok {
				timelineCount[timestamp][version] = 0
			}
		}
	}

	return timelineCount, nil
}

func (api *API) GetGroupStatusCountTimeline(groupID string) (map[time.Time](map[int](VersionCountMap)), error) {
	var timelineEntry []StatusVersionCountTimelineEntry
	// Get the versions and their number of instances per status within each of the given time intervals.
	query := fmt.Sprintf(`
	WITH time_series AS (SELECT * FROM generate_series(now() - interval '%[1]s', now(), INTERVAL '1 hour') AS ts)
	SELECT ts, (CASE WHEN status IS NULL THEN 0 ELSE status END), (CASE WHEN version IS NULL THEN '' ELSE version END), count(instance_id) as total FROM (SELECT * FROM time_series
		LEFT JOIN LATERAL(SELECT * FROM instance_status_history WHERE group_id=$1 AND created_ts BETWEEN time_series.ts - INTERVAL '1 hour' + INTERVAL '1 sec' AND time_series.ts) _ ON TRUE) AS _
	GROUP BY 1,2,3
	ORDER BY ts DESC;
	`, validityInterval)

	if err := api.dbR.SQL(query, groupID).QueryStructs(&timelineEntry); err != nil {
		return nil, err
	}

	allStatuses := make(map[int]struct{})
	timelineCount := make(map[time.Time](map[int](VersionCountMap)))

	// Create the timeline map, and gather all the statuses found.
	for _, entry := range timelineEntry {
		value, ok := timelineCount[entry.Time]
		if !ok {
			value = make(map[int](VersionCountMap))
			timelineCount[entry.Time] = value
		}

		allStatuses[entry.Status] = struct{}{}
		versionCount, ok := value[entry.Status]
		if !ok {
			versionCount = make(VersionCountMap)
		}

		versionCount[entry.Version] = entry.Total

		value[entry.Status] = versionCount
	}

	// We want to return all the status per time-interval, i.e. we don't want
	// some time-intervals to have 2 statuses accounted, and others just 1, so
	// this assigns the missing statuses per interval.
	for status := range allStatuses {
		for timestamp := range timelineCount {
			if _, ok := timelineCount[timestamp][status]; !ok {
				timelineCount[timestamp][status] = make(VersionCountMap)
			}
		}
	}

	return timelineCount, nil
}

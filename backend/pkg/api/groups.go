package api

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gopkg.in/guregu/null.v4"
)

type (
	durationParam    string
	durationCode     int
	postgresDuration string
	postgresInterval string
)

const (
	oneHour durationCode = iota
	oneDay
	sevenDays
	thirtyDays
)

const (
	// If an instance doesn't update its status for deadInstanceTimeSpan then the instance
	// is considered dead.
	deadInstanceTimeSpan = "6 months"
)

var durationParamToCode = map[durationParam]durationCode{
	"1h":  oneHour,
	"1d":  oneDay,
	"7d":  sevenDays,
	"30d": thirtyDays,
}
var (
	// ErrInvalidChannel error indicates that a channel doesn't belong to the
	// application it was supposed to belong to.
	ErrInvalidChannel = errors.New("nebraska: invalid channel")

	// ErrExpectingValidTimezone error indicates that a valid timezone wasn't
	// provided when enabling the flag PolicyOfficeHours.
	ErrExpectingValidTimezone = errors.New("nebraska: expecting valid timezone")

	// cachedGroups caches the mapping of group track names and
	// architectures to groups. It must not be modified directly but
	// replaced (atomically or via lock) by a new map to prevent data races.
	// An update must be triggered through updateCachedGroups() each time
	// a group entry changes (channel architectures are not modified after
	// creation, thus changes to the channel don't need to trigger an
	// update). A RW lock was chosen to prevent data races over the pointer
	// itself. An alternative is to use atomic loads instead of RLock()
	// and using atomic stores inside Lock() of a normal Mutex to serialize
	// the writes (or use channel handshakes instead of a mutex).
	cachedGroups                    map[GroupDescriptor]string
	cachedGroupsLock                sync.RWMutex
	cachedGroupVersionCount         = make(map[groupDurationCacheKey]groupVersionCountCache)
	cachedGroupVersionCountLock     sync.RWMutex
	cachedGroupVersionCountLifespan = time.Minute
)

type groupDurationCacheKey struct {
	GroupID  string
	Duration string
}

type groupVersionCountCache struct {
	data     map[time.Time](VersionCountMap)
	storedAt time.Time
}

type GroupDescriptor struct {
	Track string
	Arch  Arch
}

// Group represents a Nebraska application's group.
type Group struct {
	ID                        string      `db:"id" json:"id"`
	Name                      string      `db:"name" json:"name"`
	Description               string      `db:"description" json:"description"`
	CreatedTs                 time.Time   `db:"created_ts" json:"created_ts"`
	RolloutInProgress         bool        `db:"rollout_in_progress" json:"rollout_in_progress"`
	ApplicationID             string      `db:"application_id" json:"application_id"`
	ChannelID                 null.String `db:"channel_id" json:"channel_id"`
	PolicyUpdatesEnabled      bool        `db:"policy_updates_enabled" json:"policy_updates_enabled"`
	PolicySafeMode            bool        `db:"policy_safe_mode" json:"policy_safe_mode"`
	PolicyOfficeHours         bool        `db:"policy_office_hours" json:"policy_office_hours"`
	PolicyTimezone            null.String `db:"policy_timezone" json:"policy_timezone"`
	PolicyPeriodInterval      string      `db:"policy_period_interval" json:"policy_period_interval"`
	PolicyMaxUpdatesPerPeriod int         `db:"policy_max_updates_per_period" json:"policy_max_updates_per_period"`
	PolicyUpdateTimeout       string      `db:"policy_update_timeout" json:"policy_update_timeout"`
	Channel                   *Channel    `db:"channel" json:"channel,omitempty"`
	Track                     string      `db:"track" json:"track"`
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
	Total         int      `db:"total" json:"total"`
	Undefined     null.Int `db:"undefined" json:"undefined"`
	UpdateGranted null.Int `db:"update_granted" json:"update_granted"`
	Error         null.Int `db:"error" json:"error"`
	Complete      null.Int `db:"complete" json:"complete"`
	Installed     null.Int `db:"installed" json:"installed"`
	Downloaded    null.Int `db:"downloaded" json:"downloaded"`
	Downloading   null.Int `db:"downloading" json:"downloading"`
	OnHold        null.Int `db:"onhold" json:"onhold"`
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
	err = api.db.QueryRowx(query).StructScan(group)
	if err != nil {
		return nil, err
	}
	api.updateCachedGroups()
	return group, nil
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
	api.updateCachedGroups()
	return nil
}

// DeleteGroup removes the group identified by the id provided.
func (api *API) DeleteGroup(groupID string) error {
	query, _, err := goqu.Delete("groups").Where(goqu.C("id").Eq(groupID)).ToSQL()
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
	api.updateCachedGroups()
	return nil
}

// GetGroup returns the group identified by the id provided.
func (api *API) GetGroup(groupID string) (*Group, error) {
	var group Group

	query, _, err := goqu.From("groups").
		Where(goqu.C("id").Eq(groupID)).
		ToSQL()
	if err != nil {
		return nil, err
	}
	err = api.db.QueryRowx(query).StructScan(&group)
	if err != nil {
		return nil, err
	}
	if group.ChannelID.String == "" {
		group.Channel = nil
	} else {
		channel, err := api.GetChannel(group.ChannelID.String)
		switch err {
		case nil:
			group.Channel = channel
		case sql.ErrNoRows:
			group.Channel = nil
		default:
			return nil, err
		}
	}
	return &group, nil
}

// GetGroupID returns the ID of the first group identified by the track name and the channel architecture.
// The track names should be unique in combination with the group's channel architecture but this is not
// enforced on the DB level and the newest entry wins.
func (api *API) GetGroupID(trackName string, arch Arch) (string, error) {
	var cachedGroupsRef map[GroupDescriptor]string
	cachedGroupsLock.RLock()
	if cachedGroups != nil {
		// Keep a reference to the map that we found.
		cachedGroupsRef = cachedGroups
	}
	cachedGroupsLock.RUnlock()
	// Generate map on startup or if invalidated.
	if cachedGroupsRef == nil {
		cachedGroupsLock.Lock()
		cachedGroupsRef = cachedGroups
		// If a concurrent execution generated it inbetween our RUnlock() and Lock(),
		// we can use this because any invalidation inbetween must have happened
		// before the generation because all writes are sequential.
		if cachedGroupsRef == nil {
			cachedGroups = make(map[GroupDescriptor]string)
			query, _, err := goqu.From("groups").ToSQL()
			var groups []*Group
			if err == nil {
				groups, err = api.getGroupsFromQuery(query)
			}
			// Checks boths errors above.
			if err != nil {
				logger.Error().Err(err).Msg("GetGroupID error")
			} else {
				for _, group := range groups {
					if group.Channel != nil {
						descriptor := GroupDescriptor{Track: group.Track, Arch: group.Channel.Arch}
						// The groups are sorted descendingly by the creation time.
						// The newest group with the track name and arch wins.
						if otherID, ok := cachedGroups[descriptor]; ok {
							// Log a warning for others.
							logger.Warn().Str("group", group.ID).Str("group2", otherID).Str("track", group.Track).Msg("GetGroupID - another group already uses the same track name and architecture")
						}
						cachedGroups[descriptor] = group.ID
					} else {
						logger.Warn().Str("group", group.ID).Msg("GetGroupID - no channel found for")
					}
				}
			}
			// Keep a reference to the map we created.
			cachedGroupsRef = cachedGroups
		}
		cachedGroupsLock.Unlock()
	}

	cachedGroupID, ok := cachedGroupsRef[GroupDescriptor{Track: trackName, Arch: arch}]
	if !ok {
		return "", fmt.Errorf("no group found for track %v and architecture %v", trackName, arch)
	}
	return cachedGroupID, nil
}

// updateCachedGroups invalidates the cached track names in cachedGroups and
// must be called whenever the group entries are modified.
func (api *API) updateCachedGroups() {
	cachedGroupsLock.Lock()
	cachedGroups = nil
	// Generating the map is not always possible here because the database
	// can be closed.
	cachedGroupsLock.Unlock()
}

// GetGroups returns all groups that belong to the application provided.
func (api *API) GetGroups(appID string, page, perPage uint64) ([]*Group, error) {
	page, perPage = validatePaginationParams(page, perPage)
	limit, offset := sqlPaginate(page, perPage)
	query, _, err := api.groupsQuery().Where(goqu.C("application_id").Eq(appID)).
		Limit(limit).
		Offset(offset).
		ToSQL()
	if err != nil {
		return nil, err
	}
	return api.getGroupsFromQuery(query)
}

func (api *API) getGroups(appID string) ([]*Group, error) {
	query, _, err := api.groupsQuery().Where(goqu.C("application_id").Eq(appID)).ToSQL()
	if err != nil {
		return nil, err
	}
	return api.getGroupsFromQuery(query)
}

func (api *API) getGroupsFromQuery(query string) ([]*Group, error) {
	var groups []*Group
	rows, err := api.db.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		group := Group{}
		if err := rows.StructScan(&group); err != nil {
			return nil, err
		}
		if group.ChannelID.String == "" {
			group.Channel = nil
		} else {
			channel, err := api.GetChannel(group.ChannelID.String)
			switch err {
			case nil:
				group.Channel = channel
			case sql.ErrNoRows:
				group.Channel = nil
			default:
				return nil, err
			}
		}
		groups = append(groups, &group)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return groups, nil
}

// validateChannel checks if a channel belongs to the application provided.
func (api *API) validateChannel(channelID, appID string) error {
	channel, err := api.GetChannel(channelID)
	if err != nil {
		return err
	}
	if channel.ApplicationID != appID {
		return ErrInvalidChannel
	}
	return nil
}

// getGroupUpdatesStats returns a set of statistics about the distribution of
// updates and their status in the group provided.
func (api *API) getGroupUpdatesStats(group *Group) (*UpdatesStats, error) {
	var updatesStats UpdatesStats

	packageVersion := ""
	if group.Channel != nil && group.Channel.Package != nil {
		packageVersion = group.Channel.Package.Version
	}
	query, _, err := goqu.From("instance_application").Select(
		goqu.COUNT("*").As("total_instances"),
		goqu.COALESCE(goqu.SUM(goqu.L("case when last_update_version = ? then 1 else 0 end", packageVersion)), 0).As("updates_to_current_version_granted"),
		goqu.COALESCE(goqu.SUM(goqu.L("case when update_in_progress = 'false' and last_update_version = ? then 1 else 0 end", packageVersion)), 0).As("updates_to_current_version_attempted"),
		goqu.COALESCE(goqu.SUM(goqu.L("case when update_in_progress = 'false' and last_update_version = ? and last_update_version = version then 1 else 0 end", packageVersion)), 0).As("updates_to_current_version_succeeded"),
		goqu.COALESCE(goqu.SUM(goqu.L("case when update_in_progress = 'false' and last_update_version = ? and last_update_version != version then 1 else 0 end", packageVersion)), 0).As("updates_to_current_version_failed"),
		goqu.COALESCE(goqu.SUM(goqu.L("case when last_update_granted_ts > now() at time zone 'utc' - interval ? then 1 else 0 end", group.PolicyPeriodInterval)), 0).As("updates_granted_in_last_period"),
		goqu.COALESCE(goqu.SUM(goqu.L("case when update_in_progress = 'true' and now() at time zone 'utc' - last_update_granted_ts <= interval ? then 1 else 0 end", group.PolicyUpdateTimeout)), 0).As("updates_in_progress"),
		goqu.COALESCE(goqu.SUM(goqu.L("case when update_in_progress = 'true' and now() at time zone 'utc' - last_update_granted_ts > interval ? then 1 else 0 end", group.PolicyUpdateTimeout)), 0).As("updates_timed_out"),
	).Where(goqu.C("group_id").Eq(group.ID), goqu.L("last_check_for_updates > now() at time zone 'utc' - interval ?", validityInterval),
		goqu.L(ignoreFakeInstanceCondition("instance_id")),
	).ToSQL()
	if err != nil {
		return nil, err
	}
	err = api.db.QueryRowx(query).StructScan(&updatesStats)
	if err != nil {
		return nil, err
	}
	return &updatesStats, nil
}

// disableUpdates updates the group provided setting the policy_updates_enabled
// field to false. This usually happens when the first instance in a group
// processing an update to a specific version fails if safe mode is enabled.
func (api *API) disableUpdates(groupID string) error {
	query, _, err := goqu.Update("groups").
		Set(goqu.Record{"policy_updates_enabled": false}).
		Where(goqu.C("id").Eq(groupID)).
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
	query, _, err := goqu.Update("groups").
		Set(goqu.Record{"rollout_in_progress": inProgress}).
		Where(goqu.C("id").Eq(groupID)).
		ToSQL()
	if err != nil {
		return err
	}
	_, err = api.db.Exec(query)

	return err
}

// groupsQuery returns a SelectDataset prepared to return all groups. This
// query is meant to be extended later in the methods using it to filter by a
// specific group id, all groups of a given app, specify how to query the rows
// or their destination.
func (api *API) groupsQuery() *goqu.SelectDataset {
	query := goqu.From("groups").Order(goqu.I("created_ts").Desc())

	return query
}

// GetGroupVersionBreakdown returns a version breakdown of all instances running on a given group.
func (api *API) GetGroupVersionBreakdown(groupID string) ([]*VersionBreakdownEntry, error) {
	var entryList []*VersionBreakdownEntry
	query := fmt.Sprintf(`
	SELECT version, count(*) as instances, (count(*) * 100.0 / total) as percentage
	FROM instance_application, (
		SELECT count(*) as total
		FROM instance_application
		WHERE group_id=$1 AND last_check_for_updates > now() at time zone 'utc' - interval '%[1]s'
		) totals
	WHERE group_id=$1 AND last_check_for_updates > now() at time zone 'utc' - interval '%[1]s' AND %[2]s
	GROUP BY version, total
	ORDER BY regexp_matches(version, '(\d+)\.(\d+)\.(\d+)')::int[] DESC
	`, validityInterval, ignoreFakeInstanceCondition("instance_id"))
	rows, err := api.db.Queryx(query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var entry VersionBreakdownEntry
		err := rows.StructScan(&entry)
		if err != nil {
			return nil, err
		}
		entryList = append(entryList, &entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entryList, nil
}

// getGroupInstancesStats returns a summary of the status of the
// instances that belong to a given group.
func (api *API) GetGroupInstancesStats(groupID, duration string) (*InstancesStatusStats, error) {
	var instancesStats InstancesStatusStats
	durationString, _, err := durationParamToPostgresTimings(durationParam(duration))
	if err != nil {
		return nil, err
	}

	query, _, err := goqu.From("instance_application").Select(
		goqu.COUNT("*").As("total"),
		goqu.COALESCE(goqu.SUM(goqu.L("case when status IS NULL then 1 else 0 end")), 0).As("undefined"),
		goqu.COALESCE(goqu.SUM(goqu.L("case when status = ? then 1 else 0 end", InstanceStatusError)), 0).As("error"),
		goqu.COALESCE(goqu.SUM(goqu.L("case when status = ? then 1 else 0 end", InstanceStatusUpdateGranted)), 0).As("update_granted"),
		goqu.COALESCE(goqu.SUM(goqu.L("case when status = ? then 1 else 0 end", InstanceStatusComplete)), 0).As("complete"),
		goqu.COALESCE(goqu.SUM(goqu.L("case when status = ? then 1 else 0 end", InstanceStatusInstalled)), 0).As("installed"),
		goqu.COALESCE(goqu.SUM(goqu.L("case when status = ? then 1 else 0 end", InstanceStatusDownloaded)), 0).As("downloaded"),
		goqu.COALESCE(goqu.SUM(goqu.L("case when status = ? then 1 else 0 end", InstanceStatusDownloading)), 0).As("downloading"),
		goqu.COALESCE(goqu.SUM(goqu.L("case when status = ? then 1 else 0 end", InstanceStatusOnHold)), 0).As("onhold"),
	).Where(goqu.C("group_id").Eq(groupID), goqu.L("last_check_for_updates > now() at time zone 'utc' - interval ?", durationString),
		goqu.L(ignoreFakeInstanceCondition("instance_id")),
	).ToSQL()
	if err != nil {
		return nil, err
	}
	err = api.db.QueryRowx(query).StructScan(&instancesStats)
	if err != nil {
		return nil, err
	}

	return &instancesStats, nil
}
func durationCodeToPostgresTimings(code durationCode) (postgresDuration, postgresInterval, error) {
	switch code {
	case thirtyDays:
		return "30 days", "3 days", nil
	case sevenDays:
		return "7 days", "1 days", nil
	case oneDay:
		return "1days", "1 hour", nil
	case oneHour:
		return "1hour", "15 minute", nil
	default:
		return "", "", fmt.Errorf("invalid duration enumeration value %d", code)
	}
}

func durationParamToPostgresTimings(duration durationParam) (postgresDuration, postgresInterval, error) {
	code, ok := durationParamToCode[duration]
	if !ok {
		return "", "", fmt.Errorf("invalid duration param %s", duration)
	}
	return durationCodeToPostgresTimings(code)
}

func (api *API) GetGroupVersionCountTimeline(groupID string, duration string) (map[time.Time](VersionCountMap), bool, error) {
	cacheKey := groupDurationCacheKey{GroupID: groupID, Duration: duration}

	cachedGroupVersionCountLock.RLock()
	val, ok := cachedGroupVersionCount[cacheKey]
	cachedGroupVersionCountLock.RUnlock()
	if ok {
		if time.Since(val.storedAt) < cachedGroupVersionCountLifespan {
			logger.Debug().Str("cacheStatus", "HIT").Str("groupID", groupID).Str("duration", duration).Msg("GetGroupVersionCountTimeline")
			return val.data, true, nil
		}
		logger.Debug().Str("cacheStatus", "STALE").Str("groupID", groupID).Str("duration", duration).Msg("GetGroupVersionCountTimeline")
	}

	var timelineEntry []VersionCountTimelineEntry
	durationString, interval, err := durationParamToPostgresTimings(durationParam(duration))
	// Get the number of instances per version until each of the time-interval
	// divisions. This is done only for the instances that pinged the server in
	// the last time-interval.
	if err != nil {
		return nil, false, err
	}

	instanceQuery, _, err := goqu.From("instance_application").Select(goqu.L("array_agg(DISTINCT (\"instance_id\")) \"instance_id\"")).Where(goqu.C("group_id").Eq(groupID),
		goqu.L(fmt.Sprintf("last_check_for_updates >= now() - interval '%s'", durationString)),
		goqu.L(ignoreFakeInstanceCondition("instance_id"))).ToSQL()

	if err != nil {
		return nil, false, err
	}
	ids := []string{}
	err = api.db.QueryRowx(instanceQuery).Scan(pq.Array(&ids))
	if err != nil {
		return nil, false, err
	}
	values := "VALUES "
	for _, id := range ids {
		values = values + fmt.Sprintf("('%s'),", id)
	}
	values = strings.TrimRight(values, ",")

	if len(ids) == 0 {
		values = values + "('')"
	}

	query := fmt.Sprintf(`
	WITH time_series AS (SELECT * FROM generate_series(now() - interval '%[1]s', now(), INTERVAL '%[2]s') AS ts),
		 recent_instances AS (SELECT instance_id, (CASE WHEN last_update_granted_ts 
			IS NOT NULL THEN last_update_granted_ts ELSE created_ts END), version, 4 status FROM 
			instance_application WHERE group_id=$1 AND 
			last_check_for_updates >= now() - interval '%[1]s' AND %[3]s ORDER BY last_update_granted_ts DESC),
		 instance_versions AS (SELECT instance_id, created_ts, version, status 
			FROM instance_status_history WHERE instance_id = ANY (%[5]s) 
			AND status = 4 AND created_ts >= now() - interval '%[6]s' UNION (SELECT * FROM recent_instances) ORDER BY created_ts DESC)
	SELECT ts, (CASE WHEN version IS NULL THEN '' ELSE version END), 
	  sum(CASE WHEN version IS NOT null THEN 1 ELSE 0 END) total 
	FROM (SELECT * FROM time_series 
		LEFT JOIN (SELECT distinct ON (instance_id) instance_Id, version, 
		  created_ts FROM instance_versions WHERE %[4]s ORDER BY instance_Id, 
		  created_ts DESC) _ 
		ON created_ts <= time_series.ts) AS _
	GROUP BY 1,2
	ORDER BY ts DESC;
	`, durationString, interval, ignoreFakeInstanceCondition("instance_id"),
		ignoreFakeInstanceCondition("instance_Id"), values, deadInstanceTimeSpan)

	rows, err := api.db.Queryx(query, groupID)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()
	for rows.Next() {
		var timelineEntryEntity VersionCountTimelineEntry
		err := rows.StructScan(&timelineEntryEntity)
		if err != nil {
			return nil, false, err
		}
		timelineEntry = append(timelineEntry, timelineEntryEntity)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
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

		// The query may produce a time entry with an empty string for the version when there are
		// no instances for that time interval, so we skip adding those to the result.
		if entry.Version == "" {
			continue
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

	go func() {
		cachedGroupVersionCountLock.Lock()
		defer cachedGroupVersionCountLock.Unlock()
		val, ok := cachedGroupVersionCount[cacheKey]
		if !ok || time.Since(val.storedAt) >= cachedGroupVersionCountLifespan {
			logger.Debug().Str("cacheStatus", "SET").Str("groupID", groupID).Str("duration", duration).Msg("GetGroupVersionCountTimeline")
			cachedGroupVersionCount[cacheKey] = groupVersionCountCache{timelineCount, time.Now()}
		}
	}()

	return timelineCount, false, nil
}

func (api *API) GetGroupStatusCountTimeline(groupID string, duration string) (map[time.Time](map[int](VersionCountMap)), error) {
	var timelineEntry []StatusVersionCountTimelineEntry
	durationString, interval, err := durationParamToPostgresTimings(durationParam(duration))
	if err != nil {
		return nil, err
	}
	// Get the versions and their number of instances per status within each of the given time intervals.
	query := fmt.Sprintf(`
	WITH time_series AS (SELECT * FROM generate_series(now() - interval '%[1]s', now(), INTERVAL '%[2]s') AS ts),
	min_time AS (SELECT min(ts) AS min_ts FROM time_series),
	filtered_status_history AS (SELECT instance_status_history.* FROM instance_status_history,
		 min_time WHERE group_id=$1 AND %[3]s AND created_ts >= min_time.min_ts - INTERVAL '1 hour')
	SELECT ts, (CASE WHEN status IS NULL THEN 0 ELSE status END), 
	  (CASE WHEN version IS NULL THEN '' ELSE version END), count(instance_id) as total 
	FROM 
	(
		  SELECT * FROM time_series
		  LEFT JOIN(SELECT * FROM filtered_status_history) 
		  _ ON created_ts >= time_series.ts - INTERVAL '%[2]s' AND created_ts < time_series.ts
	) AS _
	GROUP BY 1,2,3
	ORDER BY ts DESC;
	`, durationString, interval, ignoreFakeInstanceCondition("instance_id"))
	rows, err := api.db.Queryx(query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var timelineEntryEntity StatusVersionCountTimelineEntry
		err := rows.StructScan(&timelineEntryEntity)
		if err != nil {
			return nil, err
		}
		timelineEntry = append(timelineEntry, timelineEntryEntity)
	}
	if err := rows.Err(); err != nil {
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

		// The query may produce a time entry with a 0 value for the status when there are
		// no instances for that time interval, so we skip adding those to the result.
		if entry.Status == 0 {
			continue
		}

		allStatuses[entry.Status] = struct{}{}
		versionCount, ok := value[entry.Status]
		if !ok {
			versionCount = make(VersionCountMap)
		}

		// The query may produce a time entry with an empty string for the version when there are
		// no instances for that time interval, so we skip adding those to the result.
		if entry.Version == "" {
			continue
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

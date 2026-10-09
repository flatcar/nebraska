package types

import (
	"errors"
	"time"

	"gopkg.in/guregu/null.v4"
)

// ErrExpectingValidTimezone error indicates that a valid timezone wasn't
// provided when enabling the flag PolicyOfficeHours.
var ErrExpectingValidTimezone = errors.New("nebraska: expecting valid timezone")

type GroupDescriptor struct {
	AppID string
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

	// AdminPolicy is the policy as stored on groups, before this node's
	// overrides apply.
	AdminPolicy GroupPolicy `db:"admin_policy" json:"-"`
	// LocalOverrides are this node's overrides, null where the admin default
	// applies.
	LocalOverrides *GroupPolicyOverrides `db:"local_overrides" json:"local_overrides,omitempty"`
}

// GroupPolicy is the update policy of a group.
type GroupPolicy struct {
	PolicyUpdatesEnabled      bool        `db:"policy_updates_enabled" json:"policy_updates_enabled"`
	PolicySafeMode            bool        `db:"policy_safe_mode" json:"policy_safe_mode"`
	PolicyOfficeHours         bool        `db:"policy_office_hours" json:"policy_office_hours"`
	PolicyTimezone            null.String `db:"policy_timezone" json:"policy_timezone"`
	PolicyPeriodInterval      string      `db:"policy_period_interval" json:"policy_period_interval"`
	PolicyMaxUpdatesPerPeriod int         `db:"policy_max_updates_per_period" json:"policy_max_updates_per_period"`
	PolicyUpdateTimeout       string      `db:"policy_update_timeout" json:"policy_update_timeout"`
}

// GroupPolicyOverrides are a node's overrides of a group's update policy.
type GroupPolicyOverrides struct {
	PolicyUpdatesEnabled      null.Bool   `db:"policy_updates_enabled" json:"policy_updates_enabled"`
	PolicySafeMode            null.Bool   `db:"policy_safe_mode" json:"policy_safe_mode"`
	PolicyOfficeHours         null.Bool   `db:"policy_office_hours" json:"policy_office_hours"`
	PolicyTimezone            null.String `db:"policy_timezone" json:"policy_timezone"`
	PolicyPeriodInterval      null.String `db:"policy_period_interval" json:"policy_period_interval"`
	PolicyMaxUpdatesPerPeriod null.Int    `db:"policy_max_updates_per_period" json:"policy_max_updates_per_period"`
	PolicyUpdateTimeout       null.String `db:"policy_update_timeout" json:"policy_update_timeout"`
}

// ShowAdminPolicy replaces the effective policy with the admin policy, the
// values UpdateGroup writes.
func (g *Group) ShowAdminPolicy() {
	g.PolicyUpdatesEnabled = g.AdminPolicy.PolicyUpdatesEnabled
	g.PolicySafeMode = g.AdminPolicy.PolicySafeMode
	g.PolicyOfficeHours = g.AdminPolicy.PolicyOfficeHours
	g.PolicyTimezone = g.AdminPolicy.PolicyTimezone
	g.PolicyPeriodInterval = g.AdminPolicy.PolicyPeriodInterval
	g.PolicyMaxUpdatesPerPeriod = g.AdminPolicy.PolicyMaxUpdatesPerPeriod
	g.PolicyUpdateTimeout = g.AdminPolicy.PolicyUpdateTimeout
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
// that may be taking place in the instances belonging to a given group.
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

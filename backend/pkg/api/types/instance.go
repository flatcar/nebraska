package types

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

const (
	// InstanceStatusUndefined indicates that the instance hasn't sent yet an
	// event to Nebraska so it doesn't know in which state it is.
	InstanceStatusUndefined int = 1 + iota

	// InstanceStatusUpdateGranted indicates that the instance has been granted
	// an update (it should be reporting soon through events how is it going).
	InstanceStatusUpdateGranted

	// InstanceStatusError indicates that the instance reported an error while
	// processing the update.
	InstanceStatusError

	// InstanceStatusComplete indicates that the instance completed the update
	// process successfully.
	InstanceStatusComplete

	// InstanceStatusInstalled indicates that the instance has installed the
	// downloaded packages, but it hasn't applied it or restarted yet.
	InstanceStatusInstalled

	// InstanceStatusDownloaded indicates that the instance downloaded
	// successfully the update package.
	InstanceStatusDownloaded

	// InstanceStatusDownloading indicates that the instance started
	// downloading the update package.
	InstanceStatusDownloading

	// InstanceStatusOnHold indicates that the instance hasn't been granted an
	// update because one of the rollout policy limits has been reached.
	InstanceStatusOnHold
)

// Instance represents an instance running one or more applications for which
// Nebraska can provide updates.
type Instance struct {
	ID           string              `db:"id" json:"id"`
	IP           string              `db:"ip" json:"ip"`
	OEM          string              `db:"oem" json:"oem,omitempty"`
	AlephVersion string              `db:"aleph_version" json:"aleph_version,omitempty"`
	CreatedTs    time.Time           `db:"created_ts" json:"created_ts"`
	Application  InstanceApplication `db:"application" json:"application,omitempty"`
	Alias        string              `db:"alias" json:"alias,omitempty"`
}

type InstancesWithTotal struct {
	TotalInstances uint64      `json:"total"`
	Instances      []*Instance `json:"instances"`
}

// InstanceApplication represents some details about an application running on
// a given instance: current version of the app, last time the instance checked
// for updates for this app, etc.
type InstanceApplication struct {
	InstanceID          string      `db:"instance_id" json:"instance_id,omitempty"`
	ApplicationID       string      `db:"application_id" json:"application_id"`
	GroupID             null.String `db:"group_id" json:"group_id"`
	Version             string      `db:"version" json:"version"`
	CreatedTs           time.Time   `db:"created_ts" json:"created_ts"`
	Status              null.Int    `db:"status" json:"status"`
	LastCheckForUpdates time.Time   `db:"last_check_for_updates" json:"last_check_for_updates"`
	LastUpdateGrantedTs null.Time   `db:"last_update_granted_ts" json:"last_update_granted_ts"`
	LastUpdateVersion   null.String `db:"last_update_version" json:"last_update_version"`
	UpdateInProgress    bool        `db:"update_in_progress" json:"update_in_progress"`
}

// InstanceStatusHistoryEntry represents an entry in the instance status
// history.
type InstanceStatusHistoryEntry struct {
	ID            int         `db:"id" json:"-"`
	Status        int         `db:"status" json:"status"`
	Version       string      `db:"version" json:"version"`
	CreatedTs     time.Time   `db:"created_ts" json:"created_ts"`
	InstanceID    string      `db:"instance_id" json:"-"`
	ApplicationID string      `db:"application_id" json:"-"`
	GroupID       string      `db:"group_id" json:"-"`
	ErrorCode     null.String `db:"error_code" json:"error_code"`
}

// InstancesQueryParams represents a helper structure used to pass a set of
// parameters when querying instances.
type InstancesQueryParams struct {
	ApplicationID string `json:"application_id"`
	GroupID       string `json:"group_id"`
	Status        int    `json:"status"`
	Version       string `json:"version"`
	Page          uint64 `json:"page"`
	PerPage       uint64 `json:"perpage"`
	SortFilter    string `json:"sort_filter"`
	SortOrder     string `json:"sort_order"`
	SearchFilter  string `json:"search_filter"`
	SearchValue   string `json:"search_value"`
}

type InstanceStats struct {
	Timestamp   time.Time `db:"timestamp" json:"timestamp"`
	ChannelName string    `db:"channel_name" json:"channel_name"`
	Arch        string    `db:"arch" json:"arch"`
	Version     string    `db:"version" json:"version"`
	Instances   int       `db:"instances" json:"instances"`
}

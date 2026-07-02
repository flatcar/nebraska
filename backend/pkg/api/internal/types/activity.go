package types

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

// Activity represents a Nebraska activity entry.
type Activity struct {
	ID              string      `db:"id" json:"id"`
	AppID           null.String `db:"application_id" json:"app_id"`
	GroupID         null.String `db:"group_id" json:"group_id"`
	CreatedTs       time.Time   `db:"created_ts" json:"created_ts"`
	Class           int         `db:"class" json:"class"`
	Severity        int         `db:"severity" json:"severity"`
	Version         string      `db:"version" json:"version"`
	ApplicationName string      `db:"application_name" json:"application_name"`
	GroupName       null.String `db:"group_name" json:"group_name"`
	ChannelName     null.String `db:"channel_name" json:"channel_name"`
	InstanceID      null.String `db:"instance_id" json:"instance_id"`
}

// ActivityQueryParams represents a helper structure used to pass a set of
// parameters when querying activity entries.
type ActivityQueryParams struct {
	AppID      string    `db:"application_id"`
	GroupID    string    `db:"group_id"`
	ChannelID  string    `db:"channel_id"`
	InstanceID string    `db:"instance_id"`
	Version    string    `db:"version"`
	Severity   int       `db:"severity"`
	Start      time.Time `db:"start"`
	End        time.Time `db:"end"`
	Page       uint64    `json:"page"`
	PerPage    uint64    `json:"perpage"`
}

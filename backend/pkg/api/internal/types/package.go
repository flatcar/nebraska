package types

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type File struct {
	ID        int64       `db:"id" json:"id"`
	PackageID string      `db:"package_id" json:"package_id"`
	Name      null.String `db:"name" json:"name"`
	Size      null.String `db:"size" json:"size"`
	Hash      null.String `db:"hash" json:"hash"`
	Hash256   null.String `db:"hash256" json:"hash256"`
	CreatedTs time.Time   `db:"created_ts" json:"created_ts"`
}

func (f File) Equals(otherFile File) bool {
	return f.Name.String == otherFile.Name.String && f.Size.String == otherFile.Size.String && f.Hash.String == otherFile.Hash.String && f.Hash256.String == otherFile.Hash256.String
}

// Package represents a Nebraska application's package.
type Package struct {
	ID                string         `db:"id" json:"id"`
	Type              int            `db:"type" json:"type"`
	Version           string         `db:"version" json:"version"`
	URL               string         `db:"url" json:"url"`
	Filename          null.String    `db:"filename" json:"filename"`
	Description       null.String    `db:"description" json:"description"`
	Size              null.String    `db:"size" json:"size"`
	Hash              null.String    `db:"hash" json:"hash"`
	CreatedTs         time.Time      `db:"created_ts" json:"created_ts"`
	ChannelsBlacklist StringArray    `db:"channels_blacklist" json:"channels_blacklist"`
	ApplicationID     string         `db:"application_id" json:"application_id"`
	FlatcarAction     *FlatcarAction `db:"flatcar_action" json:"flatcar_action"`
	Arch              Arch           `db:"arch" json:"arch"`
	ExtraFiles        []File         `db:"extra_files" json:"extra_files"`

	// Floor metadata (populated when querying floor packages)
	IsFloor     bool        `db:"is_floor" json:"is_floor,omitempty"`
	FloorReason null.String `db:"floor_reason" json:"floor_reason"`
}

// ChannelPackageFloor represents a floor package for a specific channel
type ChannelPackageFloor struct {
	ChannelID   string      `db:"channel_id" json:"channel_id"`
	PackageID   string      `db:"package_id" json:"package_id"`
	FloorReason null.String `db:"floor_reason" json:"floor_reason"`
	CreatedTs   time.Time   `db:"created_ts" json:"created_ts"`
}

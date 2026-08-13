package types

import (
	"errors"
	"time"

	"gopkg.in/guregu/null.v4"
)

// ErrInvalidPackage error indicates that a package doesn't belong to the
// application it was supposed to belong to.
var ErrInvalidPackage = errors.New("nebraska: invalid package")

// ErrPackageIsFloor error indicates that the package cannot be deleted
// because it is marked as a floor version for one or more channels.
var ErrPackageIsFloor = errors.New("nebraska: cannot delete package marked as floor version")

// ErrBlacklistingChannel error indicates that the channel the package is
// trying to blacklist is already pointing to the package.
var ErrBlacklistingChannel = errors.New("nebraska: channel trying to blacklist is already pointing to the package")

// ErrBlacklistingFloor error indicates that the package cannot be blacklisted
// because it is marked as a floor version for the channel.
var ErrBlacklistingFloor = errors.New("nebraska: cannot blacklist package marked as floor version for this channel")

// ErrPackageBlacklisted indicates that the package is blacklisted for this channel
var ErrPackageBlacklisted = errors.New("nebraska: cannot mark blacklisted package as floor")

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

const (
	// PkgTypeFlatcar indicates that the package is a Flatcar update package
	PkgTypeFlatcar int = 1 + iota

	// PkgTypeDocker indicates that the package is a Docker container
	PkgTypeDocker

	// PkgTypeRocket indicates that the package is a Rocket container
	PkgTypeRocket

	// PkgTypeOther is the generic package type.
	PkgTypeOther
)

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

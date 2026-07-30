package types

import (
	"errors"
	"time"

	"gopkg.in/guregu/null.v4"
)

// ErrInvalidChannel error indicates that a channel doesn't belong to the
// application it was supposed to belong to.
var ErrInvalidChannel = errors.New("nebraska: invalid channel")

// ErrBlacklistedChannel error indicates an attempt of creating/updating a
// channel using a package that has blacklisted the channel.
var ErrBlacklistedChannel = errors.New("nebraska: blacklisted channel")

// Channel represents a Nebraska application's channel.
type Channel struct {
	ID            string      `db:"id" json:"id"`
	Name          string      `db:"name" json:"name"`
	Color         string      `db:"color" json:"color"`
	CreatedTs     time.Time   `db:"created_ts" json:"created_ts"`
	ApplicationID string      `db:"application_id" json:"application_id"`
	PackageID     null.String `db:"package_id" json:"package_id"`
	Package       *Package    `db:"package" json:"package"`
	Arch          Arch        `db:"arch" json:"arch"`
}

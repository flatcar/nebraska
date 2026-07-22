package types

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

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

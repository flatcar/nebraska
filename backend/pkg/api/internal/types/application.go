package types

import (
	"errors"
	"time"

	"gopkg.in/guregu/null.v4"
)

// FlatcarAppID is the well-known Flatcar application UUID seeded into every
// Nebraska database.
const FlatcarAppID = "e96281a6-d1af-4bde-9a0a-97b76e56dc57"

// ErrInvalidApplicationOrGroup indicates that the application or group id
// provided are not valid or related to each other.
var ErrInvalidApplicationOrGroup = errors.New("nebraska: invalid application or group")

// Application represents a Nebraska application instance.
type Application struct {
	ID          string      `db:"id" json:"id"`
	ProductID   null.String `db:"product_id" json:"product_id"`
	Name        string      `db:"name" json:"name"`
	Description string      `db:"description" json:"description"`
	CreatedTs   time.Time   `db:"created_ts" json:"created_ts"`
	TeamID      string      `db:"team_id" json:"-"`
	Groups      []*Group    `db:"groups" json:"groups"`
	Channels    []*Channel  `db:"channels" json:"channels"`

	Instances struct {
		Count int `db:"count" json:"count"`
	} `db:"instances" json:"instances,omitempty"`
}

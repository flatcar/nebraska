package types

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

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

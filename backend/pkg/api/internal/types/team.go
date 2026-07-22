package types

import "time"

// Team represents a Nebraska team.
type Team struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	CreatedTs time.Time `db:"created_ts"`
}

package types

import "time"

// User represents a Nebraska user.
type User struct {
	ID        string    `db:"id" json:"id"`             // UUID v4 unique, created automatically
	Username  string    `db:"username" json:"username"` // unique username
	Secret    string    `db:"secret" json:"secret"`     // md5 hash from (username:realm:password)
	CreatedTs time.Time `db:"created_ts" json:"-"`      // Created automatically
	TeamID    string    `db:"team_id" json:"team_id"`   // User can be in single team
}

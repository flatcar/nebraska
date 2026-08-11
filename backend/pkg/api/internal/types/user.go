package types

import (
	"errors"
	"time"
)

// ErrUpdatingPassword indicates that something went wrong while updating
// the user's password.
var ErrUpdatingPassword = errors.New("nebraska: error updating password")

// ErrInvalidCredentials indicates that a username/password pair didn't
// match the stored secret.
var ErrInvalidCredentials = errors.New("nebraska: invalid credentials")

// User represents a Nebraska user.
type User struct {
	ID        string    `db:"id" json:"id"`             // UUID v4 unique, created automatically
	Username  string    `db:"username" json:"username"` // unique username
	Secret    string    `db:"secret" json:"secret"`     // bcrypt hash of the password, or a legacy md5 hash of (username:realm:password)
	CreatedTs time.Time `db:"created_ts" json:"-"`      // Created automatically
	TeamID    string    `db:"team_id" json:"team_id"`   // User can be in single team
}

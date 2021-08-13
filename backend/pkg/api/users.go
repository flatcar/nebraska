package api

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/doug-martin/goqu/v9"
)

const (
	// Realm used for basic authentication.
	Realm = "nebraska"
)

var (
	// ErrUpdatingPassword indicates that something went wrong while updating
	// the user's password.
	ErrUpdatingPassword = errors.New("nebraska: error updating password")
)

// User represents a Nebraska user.
type User struct {
	ID        string    `db:"id" json:"id"`             // UUID v4 unique, created automatically
	Username  string    `db:"username" json:"username"` // unique username
	Secret    string    `db:"secret" json:"secret"`     // md5 hash from (username:realm:password)
	CreatedTs time.Time `db:"created_ts" json:"-"`      // Created automatically
	TeamID    string    `db:"team_id" json:"team_id"`   // User can be in single team
}

// AddTeam registers a team.
func (api *API) AddUser(user *User) (*User, error) {
	query, _, err := goqu.Insert("users").
		Cols("username", "team_id", "secret").
		Vals(goqu.Vals{user.Username, user.TeamID, user.Secret}).
		Returning(goqu.T("users").All()).
		ToSQL()
	if err != nil {
		return nil, err
	}
	err = api.db.QueryRowx(query).StructScan(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUser returns the user identified by the username provided.
func (api *API) GetUser(username string) (*User, error) {
	var user User
	query, _, err := goqu.From("users").
		Where(goqu.C("username").Eq(username)).
		ToSQL()
	if err != nil {
		return nil, err
	}
	err = api.db.QueryRowx(query).StructScan(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (api *API) GetUsersInTeam(teamID string) ([]*User, error) {
	var users []*User
	query, _, err := goqu.From("users").
		Where(goqu.C("team_id").Eq(teamID)).
		ToSQL()
	if err != nil {
		return nil, err
	}
	rows, err := api.db.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var user User
		err := rows.StructScan(&user)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

// UpdateUserPassword updates the password of the provided user.
func (api *API) UpdateUserPassword(username, newPassword string) error {
	secret, err := api.GenerateUserSecret(username, newPassword)
	if err != nil {
		return err
	}
	query, _, err := goqu.Update("users").
		Set(goqu.Record{"secret": secret}).
		Where(goqu.C("username").Eq(username)).
		ToSQL()
	if err != nil {
		return err
	}
	result, err := api.db.Exec(query)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUpdatingPassword
	}

	return nil
}

// GenerateUserSecret generates a md5 hash from the username and password
// provided (username:realm:password).
func (api *API) GenerateUserSecret(username, password string) (string, error) {
	h := md5.New()
	if _, err := io.WriteString(h, username+":"+Realm+":"+password); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

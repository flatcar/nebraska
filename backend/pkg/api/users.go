package api

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"

	"github.com/doug-martin/goqu/v9"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/types"
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

type User = types.User

// AddUser registers a user.
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

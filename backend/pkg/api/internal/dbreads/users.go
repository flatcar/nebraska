package dbreads

import (
	"github.com/doug-martin/goqu/v9"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/types"
)

// GetUser returns the user identified by the username provided.
func (q *Queries) GetUser(username string) (*types.User, error) {
	var user types.User
	query, _, err := goqu.From("users").
		Where(goqu.C("username").Eq(username)).
		ToSQL()
	if err != nil {
		return nil, err
	}
	err = q.db.QueryRowx(query).StructScan(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (q *Queries) GetUsersInTeam(teamID string) ([]*types.User, error) {
	var users []*types.User
	query, _, err := goqu.From("users").
		Where(goqu.C("team_id").Eq(teamID)).
		ToSQL()
	if err != nil {
		return nil, err
	}
	rows, err := q.db.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var user types.User
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

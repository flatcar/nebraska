package api

import (
	"time"

	"github.com/doug-martin/goqu/v9"
)

// Team represents a Nebraska team.
type Team struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	CreatedTs time.Time `db:"created_ts"`
}

func (api *API) GetTeams() ([]*Team, error) {
	var teams []*Team
	query, _, err := goqu.From("team").
		Select("id", "name", "created_ts").
		Order(goqu.C("name").Asc()).
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
		var team Team
		err := rows.StructScan(&team)
		if err != nil {
			return nil, err
		}
		teams = append(teams, &team)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return teams, nil
}

func (api *API) GetTeam() (*Team, error) {
	var team = &Team{}
	query, _, err := goqu.From("team").
		Select("id", "name", "created_ts").
		Limit(1).
		ToSQL()
	if err != nil {
		return nil, err
	}
	err = api.db.QueryRowx(query).StructScan(team)
	if err != nil {
		return nil, err
	}
	return team, nil
}

func (api *API) UpdateTeam(team *Team) error {
	query, _, err := goqu.Update("team").
		Set(goqu.Record{"name": team.Name}).
		Where(goqu.C("id").Eq(team.ID)).
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
		return err
	}

	return nil
}

// AddTeam registers a team.
func (api *API) AddTeam(team *Team) (*Team, error) {
	var query *goqu.InsertDataset
	if team.ID != "" {
		query = goqu.Insert("team").
			Cols("id", "name").
			Vals(goqu.Vals{team.ID, team.Name}).
			Returning(goqu.T("team").All())
	} else {
		query = goqu.Insert("team").
			Cols("name").
			Vals(goqu.Vals{team.Name}).
			Returning(goqu.T("team").All())
	}
	insertQuery, _, err := query.ToSQL()
	if err != nil {
		return nil, err
	}
	err = api.db.QueryRowx(insertQuery).StructScan(team)

	if err != nil {
		return nil, err
	}
	return team, err
}

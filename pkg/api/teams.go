package api

import "time"

// Team represents a Nebraska team.
type Team struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

func (api *API) GetTeams() ([]*Team, error) {
	var teams []*Team

	err := api.dbR.
		SelectDoc("id, name, created_at").
		From("team").
		OrderBy("name").
		QueryStructs(&teams)

	if err != nil {
		return nil, err
	}
	return teams, nil
}

func (api *API) GetTeam() (*Team, error) {
	var team *Team

	err := api.dbR.
		SelectDoc("id, name, created_at").
		From("team").
		Limit(1).
		QueryStruct(&team)

	if err != nil {
		return nil, err
	}

	return team, nil
}

func (api *API) UpdateTeam(team *Team) error {
	result, err := api.dbR.
		Update("team").
		SetWhitelist(team, "name").
		Where("id = $1", team.ID).
		Exec()

	if err == nil && result.RowsAffected == 0 {
		return ErrNoRowsAffected
	}

	return err
}

// AddTeam registers a team.
func (api *API) AddTeam(team *Team) (*Team, error) {
	var err error

	if team.ID != "" {
		err = api.dbR.InsertInto("team").Whitelist("id", "name").Record(team).Returning("*").QueryStruct(team)
	} else {
		err = api.dbR.InsertInto("team").Whitelist("name").Record(team).Returning("*").QueryStruct(team)
	}

	return team, err
}

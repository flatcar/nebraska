package api

import "time"

// Team represents a Nebraska team.
type Team struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	Users     []User    `db:"-"`
}

// TableName returns a table name for Team struct. It's for GORM.
func (Team) TableName() string {
	return "team"
}

func (api *API) GetTeams() ([]*Team, error) {
	if api.useGORM() {
		var teams []*Team
		result := api.gormDB.
			Order("name").
			Find(&teams)
		if result.Error != nil {
			return nil, result.Error
		}
		return teams, nil
	}

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
	if api.useGORM() {
		var team Team
		result := api.gormDB.
			Take(&team)
		if result.Error != nil {
			return nil, result.Error
		}
		return &team, nil
	}

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
	if api.useGORM() {
		result := api.gormDB.
			Model(team).
			Select("name").
			Update(team)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNoRowsAffected
		}
		return nil
	}
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
	if api.useGORM() {
		result := api.gormDB.
			Create(team)
		if result.Error != nil {
			return nil, result.Error
		}
		return team, nil
	}
	var err error

	if team.ID != "" {
		err = api.dbR.InsertInto("team").Whitelist("id", "name").Record(team).Returning("*").QueryStruct(team)
	} else {
		err = api.dbR.InsertInto("team").Whitelist("name").Record(team).Returning("*").QueryStruct(team)
	}

	return team, err
}

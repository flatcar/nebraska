package api

import (
	"database/sql"
	"time"

	"github.com/doug-martin/goqu/v9"
	"gopkg.in/guregu/null.v4"
)

const (
	flatcarAppID = "e96281a6-d1af-4bde-9a0a-97b76e56dc57"
)

// Application represents a Nebraska application instance.
type Application struct {
	ID          string     `db:"id" json:"id"`
	Name        string     `db:"name" json:"name"`
	Description string     `db:"description" json:"description"`
	CreatedTs   time.Time  `db:"created_ts" json:"created_ts"`
	TeamID      string     `db:"team_id" json:"-"`
	Groups      []*Group   `db:"groups" json:"groups"`
	Channels    []*Channel `db:"channels" json:"channels"`

	Instances struct {
		Count int `db:"count" json:"count"`
	} `db:"instances" json:"instances,omitempty"`
}

// AddApp registers the provided application.
func (api *API) AddApp(app *Application) (*Application, error) {
	query, _, err := goqu.Insert("application").
		Cols("name", "description", "team_id").
		Vals(goqu.Vals{app.Name, app.Description, app.TeamID}).
		Returning(goqu.T("application").All()).
		ToSQL()
	if err != nil {
		return nil, err
	}
	err = api.readDb.QueryRowx(query).StructScan(app)
	if err != nil {
		return nil, err
	}

	return app, nil
}

// AddAppCloning registers the provided application, cloning the groups and
// channels from an existing application. Channels' packages will be set to null
// as packages won't be cloned.
func (api *API) AddAppCloning(app *Application, sourceAppID string) (*Application, error) {
	app, err := api.AddApp(app)
	if err != nil {
		return nil, err
	}

	// NOTE: cloning operation is not transactional and something could go wrong

	if sourceAppID != "" {
		sourceApp, err := api.GetApp(sourceAppID)
		if err != nil {
			logger.Error("AddAppCloning - could not get source app", err)
			return app, nil
		}

		channelsIDsMappings := make(map[string]null.String)

		for _, channel := range sourceApp.Channels {
			originalChannelID := channel.ID
			channel.ApplicationID = app.ID
			channel.PackageID = null.String{}
			channelCopy, err := api.AddChannel(channel)
			if err != nil {
				logger.Error("AddAppCloning - could not add channel", err)
				return app, nil // FIXME - think about what we should return to the caller
			}
			channelsIDsMappings[originalChannelID] = null.StringFrom(channelCopy.ID)
		}

		for _, group := range sourceApp.Groups {
			group.ApplicationID = app.ID
			if group.ChannelID.String != "" {
				group.ChannelID = channelsIDsMappings[group.ChannelID.String]
			}
			group.PolicyUpdatesEnabled = true
			if _, err := api.AddGroup(group); err != nil {
				logger.Error("AddAppCloning - could not add group", err)
				return app, nil // FIXME - think about what we should return to the caller
			}
		}
	}

	return app, nil
}

// UpdateApp updates an existing application using the content of the
// application provided.
func (api *API) UpdateApp(app *Application) error {
	query, _, err := goqu.Update("application").
		Set(
			goqu.Record{
				"name":        app.Name,
				"description": app.Description,
			},
		).
		Where(goqu.C("id").Eq(app.ID)).
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
		return ErrNoRowsAffected
	}
	return nil
}

// DeleteApp removes the application identified by the id provided.
func (api *API) DeleteApp(appID string) error {
	query, _, err := goqu.Delete("application").Where(goqu.C("id").Eq(appID)).ToSQL()
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
		return ErrNoRowsAffected
	}

	return nil
}

// GetApp returns the application identified by the id provided.
func (api *API) GetApp(appID string) (*Application, error) {
	var app Application
	query, _, err := api.appsQuery().
		Where(goqu.C("id").Eq(appID)).ToSQL()
	if err != nil {
		return nil, err
	}
	if err := api.readDb.QueryRowx(query).StructScan(&app); err != nil {
		return nil, err
	}
	groups, err := api.getGroups(app.ID)
	if err == nil || err == sql.ErrNoRows {
		app.Groups = groups
	} else {
		return nil, err
	}
	channels, err := api.getChannels(app.ID)
	if err == nil || err == sql.ErrNoRows {
		app.Channels = channels
	} else {
		return nil, err
	}
	app.Instances.Count, err = api.getInstanceCount(app.ID, "", validityInterval)
	if err != nil {
		return nil, err
	}
	return &app, nil
}

// GetApps returns all applications that belong to the team id provided.
func (api *API) GetApps(teamID string, page, perPage uint64) ([]*Application, error) {
	page, perPage = validatePaginationParams(page, perPage)
	var apps []*Application
	limit, offset := sqlPaginate(page, perPage)
	query, _, err := api.appsQuery().
		Where(goqu.C("team_id").Eq(teamID)).
		Limit(limit).
		Offset(offset).
		ToSQL()
	if err != nil {
		return nil, err
	}
	rows, err := api.readDb.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		app := Application{}
		err := rows.StructScan(&app)
		if err != nil {
			return nil, err
		}
		groups, err := api.getGroups(app.ID)
		if err == nil || err == sql.ErrNoRows {
			app.Groups = groups
		} else {
			return nil, err
		}
		channels, err := api.getChannels(app.ID)
		if err == nil || err == sql.ErrNoRows {
			app.Channels = channels
		} else {
			return nil, err
		}
		app.Instances.Count, err = api.getInstanceCount(app.ID, "", validityInterval)
		if err != nil {
			return nil, err
		}
		apps = append(apps, &app)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return apps, nil
}

// appsQuery returns a SelectDataset prepared to return all applications.
// This query is meant to be extended later in the methods using it to filter
// by a specific application id, all applications that belong to a given team,
// specify how to query the rows or their destination.
func (api *API) appsQuery() *goqu.SelectDataset {
	query := goqu.From("application").
		Select("id", "name", "description", "created_ts").
		Order(goqu.I("created_ts").Desc())
	return query
}

func (api *API) getInstanceCount(appID, groupID string, duration postgresDuration) (int, error) {
	query, _, err := api.appInstancesCountQuery(appID, groupID, duration).ToSQL()
	if err != nil {
		return 0, err
	}
	count := 0
	if err := api.readDb.QueryRow(query).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// appInstancesCountQuery returns a SelectDataset prepared to return the number of
// instances running a given application.
func (api *API) appInstancesCountQuery(appID, groupID string, duration postgresDuration) *goqu.SelectDataset {
	query := goqu.From("instance_application").
		Select(goqu.COUNT("*")).
		Where(
			goqu.L("last_check_for_updates > now() at time zone 'utc' - interval ?", duration),
			goqu.L(ignoreFakeInstanceCondition("instance_id")),
		)
	if appID != "" {
		query = query.Where(goqu.C("application_id").Eq(appID))
	}
	if groupID != "" {
		query = query.Where(goqu.C("group_id").Eq(groupID))
	}
	return query
}

package api

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"gopkg.in/mgutz/dat.v1"
)

const (
	flatcarAppID = "e96281a6-d1af-4bde-9a0a-97b76e56dc57"
)

// Application represents a Nebraska application instance.
type Application struct {
	ID          string     `db:"id" json:"id"`
	Name        string     `db:"name" json:"name"`
	Description string     `db:"description" json:"description"`
	CreatedAt   time.Time  `db:"created_at" json:"created_ts"`
	TeamID      string     `db:"team_id" json:"-"`
	Groups      []*Group   `db:"groups" json:"groups"`
	Channels    []*Channel `db:"channels" json:"channels"`
	Packages    []*Package `db:"packages" json:"packages"`

	Instances struct {
		Count int `db:"count" json:"count"`
	} `db:"instances" json:"instances,omitempty"`
}

// TableName returns a table name for Application struct. It's for
// GORM.
func (Application) TableName() string {
	return "application"
}

// AddApp registers the provided application.
func (api *API) AddApp(app *Application) (*Application, error) {
	if api.useGORM() {
		result, app := addAppWithGORM(api.gormDB, app)
		if result.Error != nil {
			return nil, result.Error
		}
		return app, nil
	}
	err := api.dbR.
		InsertInto("application").
		Whitelist("name", "description", "team_id").
		Record(app).
		Returning("*").
		QueryStruct(app)

	if err != nil {
		return nil, err
	}
	return app, nil
}

func addAppWithGORM(db *gorm.DB, app *Application) (*gorm.DB, *Application) {
	result := db.
		Select("name", "description", "team_id").
		Create(app)
	return result, app
}

// AddAppCloning registers the provided application, cloning the groups and
// channels from an existing application. Channels' packages will be set to null
// as packages won't be cloned.
func (api *API) AddAppCloning(app *Application, sourceAppID string) (*Application, error) {
	if api.useGORM() {
		tx := api.gormDB.Begin()
		var result *gorm.DB
		result, app = addAppWithGORM(tx, app)
		if err := result.Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		/*
			if sourceAppID != "" {
				var sourceApp *Application
				result, sourceApp = getAppWithGORM(tx, sourceAppID)
				if err != result.Error {
					tx.Rollback()
					return app, nil
				}
				channelsIDsMappings := make(map[string]dat.NullString)
				// TODO
			}
		*/
		if err := tx.Commit().Error; err != nil {
			return nil, err
		}
		return app, nil
	}
	app, err := api.AddApp(app)
	if err != nil {
		return nil, err
	}

	// NOTE: cloning operation is not transactional and something could go wrong

	if sourceAppID != "" {
		sourceApp, err := api.GetApp(sourceAppID)
		if err != nil {
			return app, nil
		}

		channelsIDsMappings := make(map[string]dat.NullString)

		for _, channel := range sourceApp.Channels {
			originalChannelID := channel.ID
			channel.ApplicationID = app.ID
			channel.PackageID = dat.NullString{}
			channelCopy, err := api.AddChannel(channel)
			if err != nil {
				return app, nil // FIXME - think about what we should return to the caller
			}
			channelsIDsMappings[originalChannelID] = dat.NullStringFrom(channelCopy.ID)
		}

		for _, group := range sourceApp.Groups {
			group.ApplicationID = app.ID
			if group.ChannelID.String != "" {
				group.ChannelID = channelsIDsMappings[group.ChannelID.String]
			}
			// TODO: why override that?
			group.PolicyUpdatesEnabled = true
			if _, err := api.AddGroup(group); err != nil {
				return app, nil // FIXME - think about what we should return to the caller
			}
		}
	}

	return app, nil
}

// UpdateApp updates an existing application using the content of the
// application provided.
func (api *API) UpdateApp(app *Application) error {
	if api.useGORM() {
		result := api.gormDB.
			Model(app).
			Select("name", "description").
			Update(app)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNoRowsAffected
		}
		return nil
	}
	result, err := api.dbR.
		Update("application").
		SetWhitelist(app, "name", "description").
		Where("id = $1", app.ID).
		Exec()

	if err == nil && result.RowsAffected == 0 {
		return ErrNoRowsAffected
	}

	return err
}

// DeleteApp removes the application identified by the id provided.
func (api *API) DeleteApp(appID string) error {
	if api.useGORM() {
		result := api.gormDB.
			Where(Application{ID: appID}).
			Delete(Application{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNoRowsAffected
		}
		return nil
	}
	result, err := api.dbR.
		DeleteFrom("application").
		Where("id = $1", appID).
		Exec()

	if err == nil && result.RowsAffected == 0 {
		return ErrNoRowsAffected
	}

	return err
}

// GetApp returns the application identified by the id provided.
func (api *API) GetApp(appID string) (*Application, error) {
	if api.useGORM() {
		result, app := getAppWithGORM(api.gormDB, appID)
		if result.Error != nil {
			return nil, result.Error
		}
		return app, nil
	}
	var app Application

	err := api.appsQuery().
		Where("id = $1", appID).
		QueryStruct(&app)

	if err != nil {
		return nil, err
	}

	return &app, nil
}

func getAppWithGORM(db *gorm.DB, appID string) (*gorm.DB, *Application) {
	var app Application
	result := db.
		Where(Application{ID: appID}).
		First(&app)
	return result, &app
}

// GetApps returns all applications that belong to the team id provided.
func (api *API) GetApps(teamID string, page, perPage uint64) ([]*Application, error) {
	page, perPage = validatePaginationParams(page, perPage)
	if api.useGORM() {
		var apps []*Application
		result := api.appsQueryGORM().Where(Application{TeamID: teamID}).Limit(perPage).Offset((page - 1) * perPage).Find(&apps)
		if result.Error != nil {
			return nil, result.Error
		}
		return apps, nil
	}

	var apps []*Application

	err := api.appsQuery().
		Where("team_id = $1", teamID).
		Paginate(page, perPage).
		QueryStructs(&apps)

	return apps, err
}

func (api *API) appsQueryGORM() *gorm.DB {
	return api.gormDB
}

// appsQuery returns a SelectDocBuilder prepared to return all applications.
// This query is meant to be extended later in the methods using it to filter
// by a specific application id, all applications that belong to a given team,
// specify how to query the rows or their destination.
func (api *API) appsQuery() *dat.SelectDocBuilder {
	return api.dbR.
		SelectDoc("id, name, description, created_at").
		One("instances", api.appInstancesCountQuery()).
		Many("groups", api.groupsQuery().Where("application_id = application.id")).
		Many("channels", api.channelsQuery().Where("application_id = application.id")).
		Many("packages", api.packagesQuery().Where("application_id = application.id")).
		From("application").
		OrderBy("created_at DESC")
}

// appInstancesCountQuery returns a SQL query prepared to return the number of
// instances running a given application.
func (api *API) appInstancesCountQuery() string {
	return fmt.Sprintf(`
	SELECT count(*)
	FROM instance_application
	WHERE application_id = application.id AND
	      last_check_for_updates > now() at time zone 'utc' - interval '%s'
	`, validityInterval)
}

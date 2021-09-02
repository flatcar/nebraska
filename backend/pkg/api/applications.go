package api

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v4"
)

type appsCache map[string]string

var (
	// cachedApps caches the mapping of apps' product-ids and UUIDs to apps' UUIDs.
	// The UUID -> UUID seeming redundancy is because this way we can use it to
	// validate also the UUIDs against existing apps.
	// It must not be modified directly but replaced (atomically or via lock)
	// by a new map to prevent data races.
	// An update must be triggered through clearCachedAppIDs() each time
	// an apps change. A RW lock was chosen to prevent data
	// races over the pointer itself.
	cachedAppIDs      appsCache
	cachedAppsIDsLock sync.RWMutex
)

const (
	flatcarAppID = "e96281a6-d1af-4bde-9a0a-97b76e56dc57"
)

// Application represents a Nebraska application instance.
type Application struct {
	ID          string      `db:"id" json:"id"`
	ProductID   null.String `db:"product_id" json:"product_id"`
	Name        string      `db:"name" json:"name"`
	Description string      `db:"description" json:"description"`
	CreatedTs   time.Time   `db:"created_ts" json:"created_ts"`
	TeamID      string      `db:"team_id" json:"-"`
	Groups      []*Group    `db:"groups" json:"groups"`
	Channels    []*Channel  `db:"channels" json:"channels"`

	Instances struct {
		Count int `db:"count" json:"count"`
	} `db:"instances" json:"instances,omitempty"`
}

// AddApp registers the provided application.
func (api *API) AddApp(app *Application) (*Application, error) {
	query, _, err := goqu.Insert("application").
		Cols("name", "product_id", "description", "team_id").
		Vals(goqu.Vals{app.Name, app.ProductID, app.Description, app.TeamID}).
		Returning(goqu.T("application").All()).
		ToSQL()
	if err != nil {
		return nil, err
	}
	err = api.db.QueryRowx(query).StructScan(app)
	if err != nil {
		return nil, err
	}

	api.clearCachedAppIDs()
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
			logger.Error().Err(err).Msg("AddAppCloning - could not get source app")
			return app, nil
		}

		channelsIDsMappings := make(map[string]null.String)

		for _, channel := range sourceApp.Channels {
			originalChannelID := channel.ID
			channel.ApplicationID = app.ID
			channel.PackageID = null.String{}
			channelCopy, err := api.AddChannel(channel)
			if err != nil {
				logger.Error().Err(err).Msg("AddAppCloning - could not add channel")
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
			group.ID = ""
			if _, err := api.AddGroup(group); err != nil {
				logger.Error().Err(err).Msg("AddAppCloning - could not add group")
				return app, nil // FIXME - think about what we should return to the caller
			}
		}
	}
	// Even though AddApp will invalidate the cache, we need to do it again here
	// to prevent eventual race issues.
	api.clearCachedAppIDs()
	return app, nil
}

// UpdateApp updates an existing application using the content of the
// application provided.
func (api *API) UpdateApp(app *Application) error {
	query, _, err := goqu.Update("application").
		Set(
			goqu.Record{
				"name":        app.Name,
				"product_id":  app.ProductID,
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

	api.clearCachedAppIDs()
	return nil
}

// DeleteApp removes the application identified by the id provided.
func (api *API) DeleteApp(appID string) error {
	realAppID, err := api.GetAppID(appID)
	if err != nil {
		return err
	}

	query, _, err := goqu.Delete("application").Where(goqu.C("id").Eq(realAppID)).ToSQL()
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

	api.clearCachedAppIDs()
	return nil
}

// GetApp returns the application identified by the id provided.
func (api *API) GetApp(appID string) (*Application, error) {
	realAppID, err := api.GetAppID(appID)
	if err != nil {
		return nil, err
	}

	var app Application
	query, _, err := goqu.From("application").
		Where(goqu.C("id").Eq(realAppID)).ToSQL()
	if err != nil {
		return nil, err
	}
	if err := api.db.QueryRowx(query).StructScan(&app); err != nil {
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

func (api *API) GetAppsCount(teamID string) (int, error) {
	query, _, err := goqu.From("application").Where(goqu.C("team_id").Eq(teamID)).Select(goqu.L("count(*)")).ToSQL()
	if err != nil {
		return 0, err
	}
	count := 0
	err = api.db.QueryRow(query).Scan(&count)

	if err != nil {
		return 0, err
	}
	return count, nil
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
	rows, err := api.db.Queryx(query)
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

// clearCachedAppIDs invalidates the cached app IDs in cachedApps and
// must be called whenever the apps entries are modified.
func (api *API) clearCachedAppIDs() {
	cachedAppsIDsLock.Lock()
	cachedAppIDs = nil
	// Generating the map is not always possible here because the database
	// can be closed.
	cachedAppsIDsLock.Unlock()
}

func (api *API) GetAppID(appOrProductID string) (string, error) {
	var cachedAppsRef appsCache
	cachedAppsIDsLock.RLock()
	if cachedAppIDs != nil {
		// Keep a reference to the map that we found.
		cachedAppsRef = cachedAppIDs
	}
	cachedAppsIDsLock.RUnlock()

	// Generate map on startup or if invalidated.
	if cachedAppsRef == nil {
		cachedAppsIDsLock.Lock()
		cachedAppsRef = cachedAppIDs

		if cachedAppsRef == nil {
			cachedAppIDs = make(appsCache)

			query, _, err := goqu.From("application").ToSQL()

			var rows *sqlx.Rows
			if err == nil {
				rows, err = api.db.Queryx(query)
			}

			if err == nil {
				defer rows.Close()
				for rows.Next() {
					app := Application{}
					err := rows.StructScan(&app)
					if err != nil {
						logger.Warn().Err(err).Msg("Failed to read app from DB")
					}

					if prodIDPtr := app.ProductID.Ptr(); prodIDPtr != nil {
						cachedAppIDs[*prodIDPtr] = app.ID
					}

					// So we can quickly validate the UUID based IDs
					cachedAppIDs[app.ID] = app.ID
				}
			} else {
				logger.Error().Err(err).Msg("Failed to get apps")
			}

			cachedAppsRef = cachedAppIDs
		}
		cachedAppsIDsLock.Unlock()
	}

	// Trim space and the {} that may surround the ID
	appIDNoBrackets := strings.TrimSpace(appOrProductID)
	if len(appIDNoBrackets) > 1 && appIDNoBrackets[0] == '{' {
		appIDNoBrackets = strings.TrimSpace(appIDNoBrackets[1 : len(appIDNoBrackets)-1])
	}

	cachedAppID, ok := cachedAppsRef[appIDNoBrackets]
	if !ok {
		return "", fmt.Errorf("no app found for ID %v", appOrProductID)
	}
	return cachedAppID, nil
}

// appsQuery returns a SelectDataset prepared to return all applications.
// This query is meant to be extended later in the methods using it to filter
// by a specific application id, all applications that belong to a given team,
// specify how to query the rows or their destination.
func (api *API) appsQuery() *goqu.SelectDataset {
	query := goqu.From("application").
		Select("id", "product_id", "name", "description", "created_ts").
		Order(goqu.I("created_ts").Desc())
	return query
}

func (api *API) getInstanceCount(appID, groupID string, duration postgresDuration) (int, error) {
	query, _, err := api.appInstancesCountQuery(appID, groupID, duration).ToSQL()
	if err != nil {
		return 0, err
	}
	count := 0
	if err := api.db.QueryRow(query).Scan(&count); err != nil {
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

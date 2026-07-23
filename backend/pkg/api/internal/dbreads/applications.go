package dbreads

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/types"
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

// GetApp returns the application identified by the id provided.
func (q *Queries) GetApp(appID string) (*types.Application, error) {
	var app types.Application
	query, _, err := goqu.From("application").
		Where(goqu.C("id").Eq(appID)).ToSQL()
	if err != nil {
		return nil, err
	}
	if err := q.db.QueryRowx(query).StructScan(&app); err != nil {
		return nil, err
	}
	groups, err := q.getGroups(app.ID)
	if err == nil || err == sql.ErrNoRows {
		app.Groups = groups
	} else {
		return nil, err
	}
	channels, err := q.getChannels(app.ID)
	if err == nil || err == sql.ErrNoRows {
		app.Channels = channels
	} else {
		return nil, err
	}
	app.Instances.Count, err = q.getInstanceCount(app.ID, "", validityInterval)
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (q *Queries) GetAppsCount(teamID string) (int, error) {
	query := goqu.From("application").Where(goqu.C("team_id").Eq(teamID)).Select(goqu.L("count(*)"))
	return q.GetCountQuery(query)
}

// GetApps returns all applications that belong to the team id provided.
func (q *Queries) GetApps(teamID string, page, perPage uint64) ([]*types.Application, error) {
	page, perPage = validatePaginationParams(page, perPage)
	var apps []*types.Application
	limit, offset := sqlPaginate(page, perPage)
	query, _, err := q.appsQuery().
		Where(goqu.C("team_id").Eq(teamID)).
		Limit(limit).
		Offset(offset).
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
		app := types.Application{}
		err := rows.StructScan(&app)
		if err != nil {
			return nil, err
		}
		groups, err := q.getGroups(app.ID)
		if err == nil || err == sql.ErrNoRows {
			app.Groups = groups
		} else {
			return nil, err
		}
		channels, err := q.getChannels(app.ID)
		if err == nil || err == sql.ErrNoRows {
			app.Channels = channels
		} else {
			return nil, err
		}
		app.Instances.Count, err = q.getInstanceCount(app.ID, "", validityInterval)
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

// ClearCachedAppIDs invalidates the cached app IDs in cachedApps and
// must be called whenever the apps entries are modified.
func (q *Queries) ClearCachedAppIDs() {
	cachedAppsIDsLock.Lock()
	cachedAppIDs = nil
	// Generating the map is not always possible here because the database
	// can be closed.
	cachedAppsIDsLock.Unlock()
}

func (q *Queries) GetAppID(appOrProductID string) (string, error) {
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
				rows, err = q.db.Queryx(query)
			}

			if err == nil {
				defer rows.Close()
				for rows.Next() {
					app := types.Application{}
					err := rows.StructScan(&app)
					if err != nil {
						l.Warn().Err(err).Msg("Failed to read app from DB")
					}

					if prodIDPtr := app.ProductID.Ptr(); prodIDPtr != nil {
						// lower case so lookups are case insensitive
						prodIDLower := strings.ToLower(*prodIDPtr)
						cachedAppIDs[prodIDLower] = app.ID
					}

					// So we can quickly validate the UUID based IDs
					cachedAppIDs[app.ID] = app.ID
				}
			} else {
				l.Error().Err(err).Msg("Failed to get apps")
			}

			cachedAppsRef = cachedAppIDs
		}
		cachedAppsIDsLock.Unlock()
	}

	// Trim space and the {} that may surround the ID
	appIDNoBrackets := strings.TrimSpace(appOrProductID)
	lastIdx := len(appIDNoBrackets) - 1
	if len(appIDNoBrackets) > 2 && appIDNoBrackets[0] == '{' && appIDNoBrackets[lastIdx] == '}' {
		appIDNoBrackets = strings.TrimSpace(appIDNoBrackets[1:lastIdx])
	}

	// Case insensitive, so use lower case as key
	appIDNoBrackets = strings.ToLower(appIDNoBrackets)

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
func (q *Queries) appsQuery() *goqu.SelectDataset {
	query := goqu.From("application").
		Select("id", "product_id", "name", "description", "created_ts").
		Order(goqu.I("created_ts").Desc())
	return query
}

func (q *Queries) getInstanceCount(appID, groupID string, duration postgresDuration) (int, error) {
	query, _, err := q.appInstancesCountQuery(appID, groupID, duration).ToSQL()
	if err != nil {
		return 0, err
	}
	count := 0
	if err := q.db.QueryRow(query).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// appInstancesCountQuery returns a SelectDataset prepared to return the number of
// instances running a given application.
func (q *Queries) appInstancesCountQuery(appID, groupID string, duration postgresDuration) *goqu.SelectDataset {
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

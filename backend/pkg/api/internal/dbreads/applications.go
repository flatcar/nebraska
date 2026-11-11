package dbreads

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
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

// GetAppID returns the id of the application with the given id or product id.
// It always queries the database, so it sees changes made by anything else
// that shares the database.
func (q *Queries) GetAppID(appOrProductID string) (string, error) {
	// Trim space and the {} that may surround the ID
	key := strings.TrimSpace(appOrProductID)
	lastIdx := len(key) - 1
	if len(key) > 2 && key[0] == '{' && key[lastIdx] == '}' {
		key = strings.TrimSpace(key[1:lastIdx])
	}

	// Case insensitive, so use lower case as key
	key = strings.ToLower(key)

	// Product ids must contain a dot, so a UUID-shaped key can only be an id.
	match := goqu.L("lower(product_id)").Eq(key)
	if isCanonicalUUID(key) {
		match = goqu.C("id").Eq(key)
	}

	// Bind the key so that every lookup reuses one prepared statement.
	query, args, err := PGDialect.From("application").Prepared(true).Select("id").Where(match).ToSQL()
	if err != nil {
		return "", err
	}

	var appID string
	if err := q.db.QueryRow(query, args...).Scan(&appID); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no app found for ID %v", appOrProductID)
		}
		l.Error().Err(err).Str("appOrProductID", appOrProductID).Msg("GetAppID")
		return "", err
	}
	return appID, nil
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

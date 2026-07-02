package dbreads

import (
	"time"

	"github.com/doug-martin/goqu/v9"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/types"
)

// Gets the activity count using some ActivityQueryParams filters
// Page and PerPage are ignored.
// Start is nil, then it defaults -3 days.
// End is nil, then it defaults to Now.
func (q *Queries) GetActivityCount(teamID string, p types.ActivityQueryParams) (int, error) {
	return q.GetCountQuery(q.activityQuery(teamID, p, true))
}

// GetActivity returns a list of activity entries that match the specified
// criteria in the query parameters.
func (q *Queries) GetActivity(teamID string, p types.ActivityQueryParams) ([]*types.Activity, error) {
	var activityEntries []*types.Activity
	query, _, err := q.activityQuery(teamID, p, false).ToSQL()
	if err != nil {
		return nil, err
	}
	rows, err := q.db.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		activityEntry := &types.Activity{}
		err := rows.StructScan(activityEntry)
		if err != nil {
			return nil, err
		}
		activityEntries = append(activityEntries, activityEntry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return activityEntries, nil
}

// activityQuery returns a SelectDataset prepared to return all activity
// entries that match the criteria provided in ActivityQueryParams.
// countSelect true returns a count without pagination (Page and PerPage are ignored)
func (q *Queries) activityQuery(teamID string, p types.ActivityQueryParams, countSelect bool) *goqu.SelectDataset {
	p.Page, p.PerPage = validatePaginationParams(p.Page, p.PerPage)

	var start, end time.Time
	if !p.Start.IsZero() {
		start = p.Start.UTC()
	} else {
		start = time.Now().UTC().AddDate(0, 0, -3)
	}
	if !p.End.IsZero() {
		end = p.End.UTC()
	} else {
		end = time.Now().UTC()
	}
	query := goqu.From(goqu.L(`
		all_activity AS a 
		INNER JOIN application AS app ON (a.application_id = app.id)
		LEFT JOIN groups AS g ON (a.group_id = g.id)
		LEFT JOIN channel AS c ON (a.channel_id = c.id)
	`))

	if countSelect {
		query = query.Select(goqu.L(`count(a)`))
	} else {
		query = query.Select(
			"a.id", "a.application_id", "a.group_id", "a.created_ts", "a.class",
			"a.severity", "a.version", "a.instance_id",
			goqu.I("app.name").As("application_name"), goqu.I("g.name").
				As("group_name"), goqu.I("c.name").As("channel_name"))
	}
	query = query.Where(goqu.I("app.team_id").
		Eq(teamID), goqu.And(goqu.I("a.created_ts").
		Gte(start), goqu.I("a.created_ts").
		Lt(end)))

	if p.AppID != "" {
		query = query.Where(goqu.I("app.id").Eq(p.AppID))
	}

	if p.GroupID != "" {
		query = query.Where(goqu.I("g.id").Eq(p.GroupID))
	}

	if p.ChannelID != "" {
		query = query.Where(goqu.I("c.id").Eq(p.ChannelID))
	}

	if p.InstanceID != "" {
		query = query.Where(goqu.I("a.instance_id").Eq(p.InstanceID))
	} else {
		query = query.Where(goqu.L(ignoreFakeInstanceCondition("a.instance_id")))
	}

	if p.Version != "" {
		query = query.Where(goqu.I("a.version").Eq(p.Version))
	}

	if p.Severity != 0 {
		query = query.Where(goqu.I("a.severity").Eq(p.Severity))
	}

	if !countSelect {
		limit, offset := sqlPaginate(p.Page, p.PerPage)
		query = query.Limit(limit).
			Offset(offset).Order(goqu.I("a.created_ts").Desc())
	}

	return query
}

// HasRecentRuntimeActivity reports whether there is matching runtime activity
// entry in the last 24h. Only runtime classes (1-5) are meaningful here.
// Admin events live in admin_activity and are not returned here.
func (q *Queries) HasRecentRuntimeActivity(class int, p types.ActivityQueryParams) bool {
	recent := time.Now().UTC().Add(-24 * time.Hour)

	query := goqu.From("activity").
		Select("id").
		Where(goqu.C("class").Eq(class)).
		Where(goqu.C("created_ts").Gt(recent))

	if p.Severity != 0 {
		query = query.Where(goqu.C("severity").Eq(p.Severity))
	}

	if p.Version != "" {
		query = query.Where(goqu.C("version").Eq(p.Version))
	}

	if p.GroupID != "" {
		query = query.Where(goqu.C("group_id").Eq(p.GroupID))
	}

	if p.AppID != "" {
		query = query.Where(goqu.I("application_id").Eq(p.AppID))
	}

	if p.InstanceID != "" {
		query = query.Where(goqu.C("instance_id").Eq(p.InstanceID))
	} else {
		query = query.Where(goqu.L(ignoreFakeInstanceCondition("instance_id")))
	}

	query = query.Limit(1)

	sql, _, err := query.ToSQL()
	if err != nil {
		return false
	}

	var id string
	if err := q.db.QueryRow(sql).Scan(&id); err != nil {
		return false
	}
	return true
}

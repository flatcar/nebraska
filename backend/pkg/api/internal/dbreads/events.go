package dbreads

import (
	"time"

	"github.com/doug-martin/goqu/v9"
	"gopkg.in/guregu/null.v4"
)

func (q *Queries) GetEvent(instanceID string, appID string, timestamp time.Time) (null.String, error) {
	query, _, err := goqu.From("event").
		Select("error_code").
		Where(goqu.C("instance_id").Eq(instanceID)).
		Where(goqu.C("application_id").Eq(appID)).
		Where(goqu.C("created_ts").Lte(timestamp)).
		Order(goqu.C("created_ts").Desc()).
		Limit(1).
		ToSQL()
	if err != nil {
		return null.NewString("", true), err
	}
	var errCode null.String
	err = q.db.QueryRow(query).Scan(&errCode)
	if err != nil {
		return null.NewString("", true), err
	}
	return errCode, nil
}

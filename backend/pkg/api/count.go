package api

import (
	"github.com/doug-martin/goqu/v9"
)

func (api *API) GetCountQuery(query *goqu.SelectDataset) (int, error) {
	sql, _, err := query.ToSQL()

	if err != nil {
		return 0, err
	}
	count := 0
	err = api.db.QueryRow(sql).Scan(&count)

	if err != nil {
		return 0, err
	}
	return count, nil
}

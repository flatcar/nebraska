package api

func (api *API) GetCountSQL(sql string, err error) (int, error) {
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

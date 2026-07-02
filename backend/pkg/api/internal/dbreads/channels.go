package dbreads

import (
	"database/sql"

	"github.com/doug-martin/goqu/v9"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/types"
)

// GetChannel returns the channel identified by the id provided.
func (q *Queries) GetChannel(channelID string) (*types.Channel, error) {
	var channel types.Channel

	query, _, err := goqu.From("channel").
		Where(goqu.C("id").Eq(channelID)).
		ToSQL()
	if err != nil {
		return nil, err
	}
	err = q.db.QueryRowx(query).StructScan(&channel)
	if err != nil {
		return nil, err
	}
	packageEntity, err := q.getPackage(channel.PackageID)
	switch err {
	case nil:
		channel.Package = packageEntity
	case sql.ErrNoRows:
		channel.Package = nil
	default:
		return nil, err
	}
	return &channel, nil
}

// GetChannelsCount retuns the total number of channels in an app
func (q *Queries) GetChannelsCount(appID string) (int, error) {
	query := goqu.From("channel").Where(goqu.C("application_id").Eq(appID)).Select(goqu.L("count(*)"))
	return q.GetCountQuery(query)
}

// GetChannels returns all channels associated to the application provided.
func (q *Queries) GetChannels(appID string, page, perPage uint64) ([]*types.Channel, error) {
	page, perPage = validatePaginationParams(page, perPage)
	limit, offset := sqlPaginate(page, perPage)
	query, _, err := q.channelsQuery().
		Where(goqu.C("application_id").Eq(appID)).
		Limit(limit).
		Offset(offset).
		ToSQL()
	if err != nil {
		return nil, err
	}
	return q.getChannelsFromQuery(query)
}

func (q *Queries) getChannels(appID string) ([]*types.Channel, error) {
	query, _, err := q.channelsQuery().
		Where(goqu.C("application_id").Eq(appID)).
		ToSQL()
	if err != nil {
		return nil, err
	}
	return q.getChannelsFromQuery(query)
}

func (q *Queries) getChannelsFromQuery(query string) ([]*types.Channel, error) {
	var channels []*types.Channel
	rows, err := q.db.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		channel := types.Channel{}
		if err := rows.StructScan(&channel); err != nil {
			return nil, err
		}

		packageEntity, err := q.getPackage(channel.PackageID)
		switch err {
		case nil:
			channel.Package = packageEntity
		case sql.ErrNoRows:
			channel.Package = nil
		default:
			return nil, err
		}
		channels = append(channels, &channel)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return channels, nil
}

// channelsQuery returns a SelectDataset prepared to return all channels.
// This query is meant to be extended later in the methods using it to filter
// by a specific channel id, all channels that belong to a given application,
// specify how to query the rows or their destination.

func (q *Queries) channelsQuery() *goqu.SelectDataset {
	query := goqu.From("channel").Order(goqu.I("name").Asc())
	return query
}

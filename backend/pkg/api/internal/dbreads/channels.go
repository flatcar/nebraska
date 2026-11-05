package dbreads

import (
	"encoding/json"
	"time"

	"github.com/doug-martin/goqu/v9"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

// channelJSON is the channel row of the query it is used in, as JSON, with its
// package nested in it:
//
//	{<channel columns>, "package": {<package columns>,
//		"channels_blacklist": [<channel id>, ...],
//		"flatcar_action": {<action columns>},
//		"extra_files": [{<file columns>}, ...]}}
//
// The keys are the column names, which the types' JSON tags match. It takes
// whichever action row comes first.
const channelJSON = `to_jsonb(channel) || jsonb_build_object('package', (
	SELECT to_jsonb(package) || jsonb_build_object(
		'channels_blacklist', (SELECT array_agg(channel_id) FROM package_channel_blacklist WHERE package_id = package.id),
		'flatcar_action', (SELECT to_jsonb(flatcar_action) FROM flatcar_action WHERE package_id = package.id LIMIT 1),
		'extra_files', (SELECT jsonb_agg(package_file ORDER BY id) FROM package_file WHERE package_id = package.id))
	FROM package WHERE id = channel.package_id))`

// GetChannel returns the channel identified by the id provided.
func (q *Queries) GetChannel(channelID string) (*types.Channel, error) {
	query, _, err := goqu.From("channel").
		Select(goqu.L(channelJSON)).
		Where(goqu.C("id").Eq(channelID)).
		ToSQL()
	if err != nil {
		return nil, err
	}
	var data []byte
	if err := q.db.QueryRowx(query).Scan(&data); err != nil {
		return nil, err
	}
	return decodeChannel(data)
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
		var data []byte
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		channel, err := decodeChannel(data)
		if err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return channels, nil
}

// channelsQuery returns a SelectDataset prepared to return all channels, each
// as channelJSON.
// This query is meant to be extended later in the methods using it to filter
// by a specific channel id, all channels that belong to a given application,
// specify how to query the rows or their destination.

func (q *Queries) channelsQuery() *goqu.SelectDataset {
	query := goqu.From("channel").Select(goqu.L(channelJSON)).Order(goqu.I("name").Asc())
	return query
}

// decodeChannel decodes a channel read as channelJSON, with its times in the
// location the driver gives a timestamptz and its action's package id set.
func decodeChannel(data []byte) (*types.Channel, error) {
	var channel types.Channel
	if err := json.Unmarshal(data, &channel); err != nil {
		return nil, err
	}
	channel.CreatedTs = channel.CreatedTs.In(time.Local)
	if p := channel.Package; p != nil {
		p.CreatedTs = p.CreatedTs.In(time.Local)
		if a := p.FlatcarAction; a != nil {
			a.CreatedTs = a.CreatedTs.In(time.Local)
			// The API leaves an action's package id out of its JSON; the action
			// belongs to the package it is nested in.
			a.PackageID = p.ID
		}
		for i := range p.ExtraFiles {
			p.ExtraFiles[i].CreatedTs = p.ExtraFiles[i].CreatedTs.In(time.Local)
		}
	}
	return &channel, nil
}

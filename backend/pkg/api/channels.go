package api

import (
	"database/sql"
	"errors"
	"time"

	"github.com/doug-martin/goqu/v9"
	"gopkg.in/guregu/null.v4"
)

var (
	// ErrInvalidPackage error indicates that a package doesn't belong to the
	// application it was supposed to belong to.
	ErrInvalidPackage = errors.New("nebraska: invalid package")

	// ErrBlacklistedChannel error indicates an attempt of creating/updating a
	// channel using a package that has blacklisted the channel.
	ErrBlacklistedChannel = errors.New("nebraska: blacklisted channel")
)

// Channel represents a Nebraska application's channel.
type Channel struct {
	ID            string      `db:"id" json:"id"`
	Name          string      `db:"name" json:"name"`
	Color         string      `db:"color" json:"color"`
	CreatedTs     time.Time   `db:"created_ts" json:"created_ts"`
	ApplicationID string      `db:"application_id" json:"application_id"`
	PackageID     null.String `db:"package_id" json:"package_id"`
	Package       *Package    `db:"package" json:"package"`
	Arch          Arch        `db:"arch" json:"arch"`
}

// AddChannel registers the provided channel.
func (api *API) AddChannel(channel *Channel) (*Channel, error) {
	if !channel.Arch.IsValid() {
		return nil, ErrInvalidArch
	}
	if channel.PackageID.String != "" {
		if _, err := api.validatePackage(channel.PackageID.String, channel.ID, channel.ApplicationID, channel.Arch); err != nil {
			return nil, err
		}
	}
	query, _, err := goqu.Insert("channel").
		Cols("name", "color", "application_id", "package_id", "arch").
		Vals(goqu.Vals{
			channel.Name,
			channel.Color,
			channel.ApplicationID,
			channel.PackageID,
			channel.Arch}).
		Returning(goqu.T("channel").All()).
		ToSQL()
	if err != nil {
		return nil, err
	}
	err = api.db.QueryRowx(query).StructScan(channel)
	if err != nil {
		return nil, err
	}
	return channel, nil
}

// UpdateChannel updates an existing channel using the content of the channel
// provided.
func (api *API) UpdateChannel(channel *Channel) error {
	channelBeforeUpdate, err := api.GetChannel(channel.ID)
	if err != nil {
		return err
	}

	var pkg *Package
	if channel.PackageID.String != "" {
		if pkg, err = api.validatePackage(channel.PackageID.String, channel.ID, channelBeforeUpdate.ApplicationID, channelBeforeUpdate.Arch); err != nil {
			return err
		}
	}
	query, _, err := goqu.Update("channel").
		Set(goqu.Record{
			"name":       channel.Name,
			"color":      channel.Color,
			"package_id": channel.PackageID,
		}).
		Where(goqu.C("id").Eq(channel.ID)).
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

	if channelBeforeUpdate.PackageID.String != channel.PackageID.String && pkg != nil {
		if err := api.newChannelActivityEntry(activityChannelPackageUpdated, activityInfo, pkg.Version, pkg.ApplicationID, channel.ID); err != nil {
			l.Error().Err(err).Msg("UpdateChannel - could not add channel activity")
		}
	}

	return nil
}

// DeleteChannel removes the channel identified by the id provided.
func (api *API) DeleteChannel(channelID string) error {
	query, _, err := goqu.Delete("channel").
		Where(goqu.C("id").Eq(channelID)).
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

	return nil
}

// GetChannel returns the channel identified by the id provided.
func (api *API) GetChannel(channelID string) (*Channel, error) {
	var channel Channel

	query, _, err := goqu.From("channel").
		Where(goqu.C("id").Eq(channelID)).
		ToSQL()
	if err != nil {
		return nil, err
	}
	err = api.db.QueryRowx(query).StructScan(&channel)
	if err != nil {
		return nil, err
	}
	packageEntity, err := api.getPackage(channel.PackageID)
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
func (api *API) GetChannelsCount(appID string) (int, error) {
	query := goqu.From("channel").Where(goqu.C("application_id").Eq(appID)).Select(goqu.L("count(*)"))
	return api.GetCountQuery(query)
}

// GetChannels returns all channels associated to the application provided.
func (api *API) GetChannels(appID string, page, perPage uint64) ([]*Channel, error) {
	page, perPage = validatePaginationParams(page, perPage)
	limit, offset := sqlPaginate(page, perPage)
	query, _, err := api.channelsQuery().
		Where(goqu.C("application_id").Eq(appID)).
		Limit(limit).
		Offset(offset).
		ToSQL()
	if err != nil {
		return nil, err
	}
	return api.getChannelsFromQuery(query)
}

func (api *API) getChannels(appID string) ([]*Channel, error) {
	query, _, err := api.channelsQuery().
		Where(goqu.C("application_id").Eq(appID)).
		ToSQL()
	if err != nil {
		return nil, err
	}
	return api.getChannelsFromQuery(query)
}

func (api *API) getChannelsFromQuery(query string) ([]*Channel, error) {
	var channels []*Channel
	rows, err := api.db.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		channel := Channel{}
		if err := rows.StructScan(&channel); err != nil {
			return nil, err
		}

		packageEntity, err := api.getPackage(channel.PackageID)
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

// validatePackage checks if a package belongs to the application provided and
// that the channel is not in the package's channels blacklist. It returns the
// package if everything is ok.
func (api *API) validatePackage(packageID, channelID, appID string, channelArch Arch) (*Package, error) {
	pkg, err := api.GetPackage(packageID)
	if err == nil {
		if pkg.ApplicationID != appID {
			return nil, ErrInvalidPackage
		}
		if pkg.Arch != channelArch {
			return nil, ErrArchMismatch
		}

		for _, blacklistedChannelID := range pkg.ChannelsBlacklist {
			if channelID == blacklistedChannelID {
				return nil, ErrBlacklistedChannel
			}
		}
	}

	return pkg, err
}

// channelsQuery returns a SelectDataset prepared to return all channels.
// This query is meant to be extended later in the methods using it to filter
// by a specific channel id, all channels that belong to a given application,
// specify how to query the rows or their destination.

func (api *API) channelsQuery() *goqu.SelectDataset {
	query := goqu.From("channel").Order(goqu.I("name").Asc())
	return query
}

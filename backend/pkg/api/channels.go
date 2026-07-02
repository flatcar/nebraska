package api

import (
	"errors"

	"github.com/doug-martin/goqu/v9"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/types"
)

var (
	// ErrInvalidPackage error indicates that a package doesn't belong to the
	// application it was supposed to belong to.
	ErrInvalidPackage = errors.New("nebraska: invalid package")

	// ErrBlacklistedChannel error indicates an attempt of creating/updating a
	// channel using a package that has blacklisted the channel.
	ErrBlacklistedChannel = errors.New("nebraska: blacklisted channel")
)

type Channel = types.Channel

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

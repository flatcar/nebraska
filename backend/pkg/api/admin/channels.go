package admin

import (
	"github.com/doug-martin/goqu/v9"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

// AddChannel registers the provided channel.
func (s *Service) AddChannel(channel *types.Channel) (*types.Channel, error) {
	if !channel.Arch.IsValid() {
		return nil, types.ErrInvalidArch
	}
	if channel.PackageID.String != "" {
		if _, err := s.validatePackage(channel.PackageID.String, channel.ID, channel.ApplicationID, channel.Arch); err != nil {
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
	err = s.db.QueryRowx(query).StructScan(channel)
	if err != nil {
		return nil, err
	}
	return channel, nil
}

// UpdateChannel updates an existing channel using the content of the channel
// provided.
func (s *Service) UpdateChannel(channel *types.Channel) error {
	channelBeforeUpdate, err := s.GetChannel(channel.ID)
	if err != nil {
		return err
	}

	var pkg *types.Package
	if channel.PackageID.String != "" {
		if pkg, err = s.validatePackage(channel.PackageID.String, channel.ID, channelBeforeUpdate.ApplicationID, channelBeforeUpdate.Arch); err != nil {
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
	result, err := s.db.Exec(query)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return types.ErrNoRowsAffected
	}

	if channelBeforeUpdate.PackageID.String != channel.PackageID.String && pkg != nil {
		if err := s.newChannelActivityEntry(types.ActivityChannelPackageUpdated, types.ActivityInfo, pkg.Version, pkg.ApplicationID, channel.ID); err != nil {
			l.Error().Err(err).Msg("UpdateChannel - could not add channel activity")
		}
	}

	return nil
}

// DeleteChannel removes the channel identified by the id provided.
func (s *Service) DeleteChannel(channelID string) error {
	query, _, err := goqu.Delete("channel").
		Where(goqu.C("id").Eq(channelID)).
		ToSQL()
	if err != nil {
		return err
	}
	result, err := s.db.Exec(query)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return types.ErrNoRowsAffected
	}

	return nil
}

// validatePackage checks if a package belongs to the application provided and
// that the channel is not in the package's channels blacklist. It returns the
// package if everything is ok.
func (s *Service) validatePackage(packageID, channelID, appID string, channelArch types.Arch) (*types.Package, error) {
	pkg, err := s.GetPackage(packageID)
	if err == nil {
		if pkg.ApplicationID != appID {
			return nil, types.ErrInvalidPackage
		}
		if pkg.Arch != channelArch {
			return nil, types.ErrArchMismatch
		}

		for _, blacklistedChannelID := range pkg.ChannelsBlacklist {
			if channelID == blacklistedChannelID {
				return nil, types.ErrBlacklistedChannel
			}
		}
	}

	return pkg, err
}

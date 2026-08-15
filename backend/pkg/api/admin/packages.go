package admin

import (
	"database/sql"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/dbreads"
	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

// checkMatchingArch returns an error if the arch does not match the channels
func (s *Service) checkMatchingArch(channelIDs types.StringArray, arch types.Arch) error {
	if len(channelIDs) == 0 {
		return nil
	}

	query, _, err := goqu.From("channel").
		Select(goqu.COUNT("*")).
		Where(goqu.Ex{"id": channelIDs}).
		Where(goqu.C("arch").Neq(arch)).
		ToSQL()

	if err != nil {
		return err
	}
	count := 0
	if err := s.db.QueryRow(query).Scan(&count); err != nil {
		return err
	}

	if count > 0 {
		return types.ErrArchMismatch
	}

	return nil
}

// addPackage contains the common logic for adding a package.
// It handles validation, package insertion, blacklist, and files within a transaction.
// The caller is responsible for handling FlatcarAction and committing the transaction.
func (s *Service) addPackage(pkg *types.Package, tx *sqlx.Tx) error {
	if !dbreads.IsValidSemver(pkg.Version) {
		return types.ErrInvalidSemver
	}
	if !pkg.Arch.IsValid() {
		return types.ErrInvalidArch
	}
	if err := s.checkMatchingArch(pkg.ChannelsBlacklist, pkg.Arch); err != nil {
		return err
	}

	query, _, err := goqu.Insert("package").
		Cols("type", "filename", "description", "size", "hash", "url", "version", "application_id", "arch").
		Vals(goqu.Vals{
			pkg.Type,
			pkg.Filename,
			pkg.Description,
			pkg.Size,
			pkg.Hash,
			pkg.URL,
			pkg.Version,
			pkg.ApplicationID,
			pkg.Arch,
		}).
		Returning(goqu.T("package").All()).
		ToSQL()
	if err != nil {
		return err
	}
	if err = tx.QueryRowx(query).StructScan(pkg); err != nil {
		return err
	}

	if len(pkg.ChannelsBlacklist) > 0 {
		for _, channelID := range pkg.ChannelsBlacklist {
			query, _, err := goqu.Insert("package_channel_blacklist").
				Cols("package_id", "channel_id").
				Vals(goqu.Vals{pkg.ID, channelID}).
				ToSQL()
			if err != nil {
				return err
			}
			if _, err = tx.Exec(query); err != nil {
				return err
			}
		}
	}

	if err = s.updatePackageFiles(tx, pkg, nil); err != nil {
		return err
	}

	return nil
}

// AddPackage registers the provided package for manual/UI creation.
// For Flatcar packages, only minimal FlatcarAction data (sha256) is required.
func (s *Service) AddPackage(pkg *types.Package) (*types.Package, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			l.Error().Err(err).Msg("AddPackage - could not roll back")
		}
	}()

	if err := s.addPackage(pkg, tx); err != nil {
		return nil, err
	}

	// For manual packages creation only, insert minimal FlatcarAction (sha256 only)
	if pkg.Type == types.PkgTypeFlatcar && pkg.FlatcarAction != nil {
		query, _, err := goqu.Insert("flatcar_action").
			Cols("package_id", "sha256").
			Vals(goqu.Vals{pkg.ID, pkg.FlatcarAction.Sha256}).
			Returning(goqu.T("flatcar_action").All()).
			ToSQL()
		if err != nil {
			return nil, err
		}
		flatcarAction := &types.FlatcarAction{}
		if err = tx.QueryRowx(query).StructScan(flatcarAction); err != nil {
			return nil, err
		}
		pkg.FlatcarAction = flatcarAction
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return pkg, nil
}

// AddPackageWithMetadata registers a package with complete FlatcarAction metadata.
// This is typically used by automated systems (like syncer) that have full update metadata.
// For Flatcar packages, all critical FlatcarAction fields must be provided.
func (s *Service) AddPackageWithMetadata(pkg *types.Package) (*types.Package, error) {
	if pkg.Type == types.PkgTypeFlatcar {
		if pkg.FlatcarAction == nil {
			return nil, fmt.Errorf("flatcar packages require FlatcarAction metadata")
		}
		if pkg.FlatcarAction.Event == "" || pkg.FlatcarAction.Sha256 == "" {
			return nil, fmt.Errorf("FlatcarAction must have Event and Sha256 fields")
		}
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			l.Error().Err(err).Msg("AddPackageWithCompleteMetadata - could not roll back")
		}
	}()

	if err := s.addPackage(pkg, tx); err != nil {
		return nil, err
	}

	if pkg.Type == types.PkgTypeFlatcar && pkg.FlatcarAction != nil {
		query, _, err := goqu.Insert("flatcar_action").
			Cols("package_id", "event", "chromeos_version", "sha256", "needs_admin",
				"is_delta", "disable_payload_backoff", "metadata_signature_rsa",
				"metadata_size", "deadline").
			Vals(goqu.Vals{
				pkg.ID,
				pkg.FlatcarAction.Event,
				pkg.FlatcarAction.ChromeOSVersion,
				pkg.FlatcarAction.Sha256,
				pkg.FlatcarAction.NeedsAdmin,
				pkg.FlatcarAction.IsDelta,
				pkg.FlatcarAction.DisablePayloadBackoff,
				pkg.FlatcarAction.MetadataSignatureRsa,
				pkg.FlatcarAction.MetadataSize,
				pkg.FlatcarAction.Deadline,
			}).
			Returning(goqu.T("flatcar_action").All()).
			ToSQL()
		if err != nil {
			return nil, err
		}
		flatcarAction := &types.FlatcarAction{}
		if err = tx.QueryRowx(query).StructScan(flatcarAction); err != nil {
			return nil, err
		}
		pkg.FlatcarAction = flatcarAction
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return pkg, nil
}

// UpdatePackage updates an existing package using the content of the package
// provided.
func (s *Service) UpdatePackage(pkg *types.Package) error {
	if !dbreads.IsValidSemver(pkg.Version) {
		return types.ErrInvalidSemver
	}
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			l.Error().Err(err).Msg("UpdatePackage - could not roll back")
		}
	}()
	query, _, err := goqu.Update("package").
		Set(goqu.Record{
			"type":        pkg.Type,
			"filename":    pkg.Filename,
			"description": pkg.Description,
			"size":        pkg.Size,
			"hash":        pkg.Hash,
			"url":         pkg.URL,
			"version":     pkg.Version,
		}).
		Where(goqu.C("id").Eq(pkg.ID)).
		ToSQL()
	if err != nil {
		return err
	}
	result, err := tx.Exec(query)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	} else if rowsAffected == 0 {
		return types.ErrNoRowsAffected
	}

	oldPkg, err := s.GetPackage(pkg.ID)
	if err != nil {
		return err
	}

	if err = s.updatePackageBlacklistedChannels(tx, pkg, oldPkg); err != nil {
		return err
	}

	if err = s.updatePackageFiles(tx, pkg, oldPkg); err != nil {
		return err
	}

	if pkg.Type == types.PkgTypeFlatcar && pkg.FlatcarAction != nil {
		if pkg.FlatcarAction.ID == "" {
			pkg.FlatcarAction.ID = uuid.New().String()
		}
		query, _, err = goqu.Insert("flatcar_action").
			Cols("id", "package_id", "sha256").
			Vals(goqu.Vals{pkg.FlatcarAction.ID, pkg.ID, pkg.FlatcarAction.Sha256}).
			OnConflict(goqu.DoUpdate("id", goqu.Record{"sha256": pkg.FlatcarAction.Sha256, "package_id": pkg.ID})).
			Returning(goqu.T("flatcar_action").All()).
			ToSQL()
		if err != nil {
			return err
		}
		err = tx.QueryRowx(query).StructScan(pkg.FlatcarAction)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// DeletePackage removes the package identified by the id provided.
// It will fail if the package is marked as a floor for any channel.
func (s *Service) DeletePackage(pkgID string) error {
	// Use a single query that checks for floor status and deletes in one operation
	query, _, err := goqu.Delete("package").
		Where(goqu.And(
			goqu.C("id").Eq(pkgID),
			// Check that package is not a floor in any channel
			goqu.L("NOT EXISTS (SELECT 1 FROM channel_package_floors WHERE package_id = package.id)"),
		)).
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
		// Check if package exists and if it's a floor
		var exists bool
		var isFloor bool
		err = s.db.QueryRow(`
			SELECT 
				EXISTS(SELECT 1 FROM package WHERE id = $1),
				EXISTS(SELECT 1 FROM channel_package_floors WHERE package_id = $1)
		`, pkgID).Scan(&exists, &isFloor)
		if err != nil {
			return err
		}
		if !exists {
			return types.ErrNoRowsAffected
		}
		if isFloor {
			return types.ErrPackageIsFloor
		}
		return types.ErrNoRowsAffected
	}

	return nil
}

// updatePackageBlacklistedChannels adds or removes as needed channels to the
// package's channels blacklist based on the new entries provided in the updated
// package entry.
//
// This method is part of the transaction that updates a package and when it's
// called, the package has already been updated except for the channels
// blacklist, that may happen here if needed.
func (s *Service) updatePackageBlacklistedChannels(tx *sqlx.Tx, pkg *types.Package, oldPkg *types.Package) error {
	pkgUpdated := oldPkg

	newChannelsBlacklist := make(map[string]struct{}, len(pkg.ChannelsBlacklist))
	for _, channelID := range pkg.ChannelsBlacklist {
		newChannelsBlacklist[channelID] = struct{}{}
	}

	oldChannelsBlacklist := make(map[string]struct{}, len(pkgUpdated.ChannelsBlacklist))
	for _, channelID := range pkgUpdated.ChannelsBlacklist {
		oldChannelsBlacklist[channelID] = struct{}{}
	}

	for channelID := range newChannelsBlacklist {
		if _, ok := oldChannelsBlacklist[channelID]; ok {
			continue
		}
		channel, err := s.GetChannel(channelID)
		if err != nil {
			return err
		}
		if channel.PackageID.String == pkg.ID {
			return types.ErrBlacklistingChannel
		}
		// Check if package is a floor for this channel
		isFloor, err := s.isPackageFloorForChannel(pkg.ID, channelID)
		if err != nil {
			return err
		}
		if isFloor {
			return types.ErrBlacklistingFloor
		}
		query, _, err := goqu.Insert("package_channel_blacklist").
			Cols("package_id", "channel_id").
			Vals(goqu.Vals{pkg.ID, channelID}).
			ToSQL()
		if err != nil {
			return err
		}
		_, err = tx.Exec(query)

		if err != nil {
			return err
		}
	}

	for channelID := range oldChannelsBlacklist {
		if _, ok := newChannelsBlacklist[channelID]; ok {
			continue
		}
		query, _, err := goqu.Delete("package_channel_blacklist").
			Where(goqu.C("package_id").Eq(pkg.ID), goqu.C("channel_id").Eq(channelID)).
			ToSQL()
		if err != nil {
			return err
		}
		_, err = tx.Exec(query)

		if err != nil {
			return err
		}
	}

	return nil
}

// updatePackageFiles adds or removes as needed files to the
// package's extra files list based on the new entries provided in the updated
// package entry.
//
// This method is part of the transaction that updates a package and when it's
// called, the package has already been updated except for the files, that may
// happen here if needed.
func (s *Service) updatePackageFiles(tx *sqlx.Tx, pkg *types.Package, oldPkg *types.Package) error {
	var oldFiles map[int64]types.File

	if oldPkg != nil {
		oldFiles = make(map[int64]types.File, len(oldPkg.ExtraFiles))
		for _, fileInfo := range oldPkg.ExtraFiles {
			oldFiles[fileInfo.ID] = fileInfo
		}
	}

	for _, newFile := range pkg.ExtraFiles {
		isUpdate := false
		if fileInfo, ok := oldFiles[newFile.ID]; ok {
			if fileInfo.ID == newFile.ID {
				delete(oldFiles, newFile.ID)

				// If nothing changed, don't touch this file
				if fileInfo.Equals(newFile) {
					continue
				}

				isUpdate = true
			}
		}

		if isUpdate {
			query, _, err := goqu.Update("package_file").
				Set(goqu.Record{
					"name":    newFile.Name.String,
					"size":    newFile.Size.String,
					"hash":    newFile.Hash.String,
					"hash256": newFile.Hash256.String,
				}).
				Where(goqu.C("id").Eq(newFile.ID)).
				ToSQL()
			if err != nil {
				return err
			}
			_, err = tx.Exec(query)

			if err != nil {
				return err
			}

			continue
		}

		query, _, err := goqu.Insert("package_file").
			Cols("package_id", "name", "size", "hash", "hash256").
			Vals(goqu.Vals{pkg.ID, newFile.Name.String, newFile.Size.String, newFile.Hash.String, newFile.Hash256.String}).
			ToSQL()
		if err != nil {
			return err
		}
		_, err = tx.Exec(query)

		if err != nil {
			return err
		}
	}

	for fileName := range oldFiles {
		oldFile := oldFiles[fileName]
		query, _, err := goqu.Delete("package_file").
			Where(goqu.C("package_id").Eq(pkg.ID), goqu.C("id").Eq(oldFile.ID)).
			ToSQL()
		if err != nil {
			return err
		}
		_, err = tx.Exec(query)

		if err != nil {
			return err
		}
	}

	return nil
}

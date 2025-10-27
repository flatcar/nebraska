package api

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v4"
)

const (
	// PkgTypeFlatcar indicates that the package is a Flatcar update package
	PkgTypeFlatcar int = 1 + iota

	// PkgTypeDocker indicates that the package is a Docker container
	PkgTypeDocker

	// PkgTypeRocket indicates that the package is a Rocket container
	PkgTypeRocket

	// PkgTypeOther is the generic package type.
	PkgTypeOther
)

var (
	// ErrBlacklistingChannel error indicates that the channel the package is
	// trying to blacklist is already pointing to the package.
	ErrBlacklistingChannel = errors.New("nebraska: channel trying to blacklist is already pointing to the package")

	// ErrPackageIsFloor error indicates that the package cannot be deleted
	// because it is marked as a floor version for one or more channels.
	ErrPackageIsFloor = errors.New("nebraska: cannot delete package marked as floor version")

	// ErrBlacklistingFloor error indicates that the package cannot be blacklisted
	// because it is marked as a floor version for the channel.
	ErrBlacklistingFloor = errors.New("nebraska: cannot blacklist package marked as floor version for this channel")
)

type File struct {
	ID        int64       `db:"id" json:"id"`
	PackageID string      `db:"package_id" json:"package_id"`
	Name      null.String `db:"name" json:"name"`
	Size      null.String `db:"size" json:"size"`
	Hash      null.String `db:"hash" json:"hash"`
	Hash256   null.String `db:"hash256" json:"hash256"`
	CreatedTs time.Time   `db:"created_ts" json:"created_ts"`
}

func (f File) Equals(otherFile File) bool {
	return f.Name.String == otherFile.Name.String && f.Size.String == otherFile.Size.String && f.Hash.String == otherFile.Hash.String && f.Hash256.String == otherFile.Hash256.String
}

// Package represents a Nebraska application's package.
type Package struct {
	ID                string         `db:"id" json:"id"`
	Type              int            `db:"type" json:"type"`
	Version           string         `db:"version" json:"version"`
	URL               string         `db:"url" json:"url"`
	Filename          null.String    `db:"filename" json:"filename"`
	Description       null.String    `db:"description" json:"description"`
	Size              null.String    `db:"size" json:"size"`
	Hash              null.String    `db:"hash" json:"hash"`
	CreatedTs         time.Time      `db:"created_ts" json:"created_ts"`
	ChannelsBlacklist StringArray    `db:"channels_blacklist" json:"channels_blacklist"`
	ApplicationID     string         `db:"application_id" json:"application_id"`
	FlatcarAction     *FlatcarAction `db:"flatcar_action" json:"flatcar_action"`
	Arch              Arch           `db:"arch" json:"arch"`
	ExtraFiles        []File         `db:"extra_files" json:"extra_files"`

	// Floor metadata (populated when querying floor packages)
	IsFloor     bool        `db:"is_floor" json:"is_floor,omitempty"`
	FloorReason null.String `db:"floor_reason" json:"floor_reason"`
}

// ChannelPackageFloor represents a floor package for a specific channel
type ChannelPackageFloor struct {
	ChannelID   string      `db:"channel_id" json:"channel_id"`
	PackageID   string      `db:"package_id" json:"package_id"`
	FloorReason null.String `db:"floor_reason" json:"floor_reason"`
	CreatedTs   time.Time   `db:"created_ts" json:"created_ts"`
}

// checkMatchingArch returns an error if the arch does not match the channels
func (api *API) checkMatchingArch(channelIDs StringArray, arch Arch) error {
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
	if err := api.db.QueryRow(query).Scan(&count); err != nil {
		return err
	}

	if count > 0 {
		return ErrArchMismatch
	}

	return nil
}

// addPackage contains the common logic for adding a package.
// It handles validation, package insertion, blacklist, and files within a transaction.
// The caller is responsible for handling FlatcarAction and committing the transaction.
func (api *API) addPackage(pkg *Package, tx *sqlx.Tx) error {
	if !isValidSemver(pkg.Version) {
		return ErrInvalidSemver
	}
	if !pkg.Arch.IsValid() {
		return ErrInvalidArch
	}
	if err := api.checkMatchingArch(pkg.ChannelsBlacklist, pkg.Arch); err != nil {
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

	if err = api.updatePackageFiles(tx, pkg, nil); err != nil {
		return err
	}

	return nil
}

// AddPackage registers the provided package for manual/UI creation.
// For Flatcar packages, only minimal FlatcarAction data (sha256) is required.
func (api *API) AddPackage(pkg *Package) (*Package, error) {
	tx, err := api.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			l.Error().Err(err).Msg("AddPackage - could not roll back")
		}
	}()

	if err := api.addPackage(pkg, tx); err != nil {
		return nil, err
	}

	// For manual packages creation only, insert minimal FlatcarAction (sha256 only)
	if pkg.Type == PkgTypeFlatcar && pkg.FlatcarAction != nil {
		query, _, err := goqu.Insert("flatcar_action").
			Cols("package_id", "sha256").
			Vals(goqu.Vals{pkg.ID, pkg.FlatcarAction.Sha256}).
			Returning(goqu.T("flatcar_action").All()).
			ToSQL()
		if err != nil {
			return nil, err
		}
		flatcarAction := &FlatcarAction{}
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
func (api *API) AddPackageWithMetadata(pkg *Package) (*Package, error) {
	if pkg.Type == PkgTypeFlatcar {
		if pkg.FlatcarAction == nil {
			return nil, fmt.Errorf("flatcar packages require FlatcarAction metadata")
		}
		if pkg.FlatcarAction.Event == "" || pkg.FlatcarAction.Sha256 == "" {
			return nil, fmt.Errorf("FlatcarAction must have Event and Sha256 fields")
		}
	}

	tx, err := api.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			l.Error().Err(err).Msg("AddPackageWithCompleteMetadata - could not roll back")
		}
	}()

	if err := api.addPackage(pkg, tx); err != nil {
		return nil, err
	}

	if pkg.Type == PkgTypeFlatcar && pkg.FlatcarAction != nil {
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
		flatcarAction := &FlatcarAction{}
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
func (api *API) UpdatePackage(pkg *Package) error {
	if !isValidSemver(pkg.Version) {
		return ErrInvalidSemver
	}
	tx, err := api.db.Beginx()
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
		return ErrNoRowsAffected
	}

	oldPkg, err := api.GetPackage(pkg.ID)
	if err != nil {
		return err
	}

	if err = api.updatePackageBlacklistedChannels(tx, pkg, oldPkg); err != nil {
		return err
	}

	if err = api.updatePackageFiles(tx, pkg, oldPkg); err != nil {
		return err
	}

	if pkg.Type == PkgTypeFlatcar && pkg.FlatcarAction != nil {
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
func (api *API) DeletePackage(pkgID string) error {
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

	result, err := api.db.Exec(query)
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
		err = api.db.QueryRow(`
			SELECT 
				EXISTS(SELECT 1 FROM package WHERE id = $1),
				EXISTS(SELECT 1 FROM channel_package_floors WHERE package_id = $1)
		`, pkgID).Scan(&exists, &isFloor)
		if err != nil {
			return err
		}
		if !exists {
			return ErrNoRowsAffected
		}
		if isFloor {
			return ErrPackageIsFloor
		}
		return ErrNoRowsAffected
	}

	return nil
}

// GetPackage returns the package identified by the id provided.
func (api *API) GetPackage(pkgID string) (*Package, error) {
	return api.getPackage(null.StringFrom(pkgID))
}

// GetPackageByVersionAndArch returns the package identified by the
// application ID, version and arch provided.
func (api *API) GetPackageByVersionAndArch(appID, version string, arch Arch) (*Package, error) {
	var pkg Package
	if !isValidSemver(version) {
		return nil, fmt.Errorf("error GetPackageByVersionAndArch version %s is not valid", version)
	}
	query, _, err := api.packagesQuery().
		Where(goqu.C("application_id").Eq(appID), goqu.C("arch").Eq(arch), goqu.C("version").Eq(version)).
		ToSQL()
	if err != nil {
		return nil, err
	}
	err = api.db.QueryRowx(query).StructScan(&pkg)
	if err != nil {
		return nil, err
	}
	flatcarAction, err := api.getFlatcarAction(pkg.ID)
	switch err {
	case nil:
		pkg.FlatcarAction = flatcarAction
	case sql.ErrNoRows:
		pkg.FlatcarAction = nil
	default:
		return nil, err
	}
	return &pkg, nil
}

// GetPackagesCount retuns the total number of package in an app
func (api *API) GetPackagesCount(appID string, searchVersion *string) (int, error) {
	query := goqu.From(goqu.L("package LEFT JOIN package_channel_blacklist pcb ON package.id = pcb.package_id")).
		Select(goqu.L(`package.*,
	    array_agg(pcb.channel_id) FILTER (WHERE pcb.channel_id IS NOT NULL) as channels_blacklist
	    `)).Where(goqu.C("application_id").Eq(appID)).
		GroupBy("package.id")
	if searchVersion != nil {
		*searchVersion = "%" + strings.ToLower(*searchVersion) + "%"
		query = query.Where(goqu.I("version").ILike(*searchVersion))
	}
	query = goqu.From(query).Select(goqu.L("count (*)"))
	return api.GetCountQuery(query)
}

// GetPackages returns all packages associated to the application provided.
func (api *API) GetPackages(appID string, page, perPage uint64, searchVersion *string) ([]*Package, error) {
	page, perPage = validatePaginationParams(page, perPage)
	limit, offset := sqlPaginate(page, perPage)
	query := api.packagesQuery().
		Where(goqu.C("application_id").Eq(appID)).
		Limit(limit).
		Offset(offset)
	if searchVersion != nil {
		*searchVersion = "%" + strings.ToLower(*searchVersion) + "%"
		query = query.Where(goqu.I("version").ILike(*searchVersion))
	}
	queryString, _, err := query.ToSQL()
	if err != nil {
		return nil, err
	}

	return api.getPackagesFromQuery(queryString)
}

func (api *API) getPackagesFromQuery(query string) ([]*Package, error) {
	var pkgs []*Package
	rows, err := api.db.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Load all packages
	for rows.Next() {
		pkg := Package{}
		err = rows.StructScan(&pkg)
		if err != nil {
			return nil, err
		}
		pkgs = append(pkgs, &pkg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(pkgs) == 0 {
		return pkgs, nil
	}

	// Use loadPackageExtras to batch load extra files and actions
	return api.loadPackageExtras(pkgs)
}

// loadPackageExtras loads extra files and flatcar actions for packages efficiently
func (api *API) loadPackageExtras(packages []*Package) ([]*Package, error) {
	if len(packages) == 0 {
		return packages, nil
	}

	// Collect package IDs
	pkgIDs := make([]string, len(packages))
	for i, pkg := range packages {
		pkgIDs[i] = pkg.ID
	}

	// Load extra files
	query, _, err := goqu.From("package_file").
		Where(goqu.C("package_id").In(pkgIDs)).
		Order(goqu.C("package_id").Asc(), goqu.C("id").Asc()).
		ToSQL()
	if err != nil {
		return nil, err
	}

	var files []File
	if err := api.db.Select(&files, query); err != nil {
		return nil, err
	}

	filesByPkg := make(map[string][]File)
	for _, file := range files {
		filesByPkg[file.PackageID] = append(filesByPkg[file.PackageID], file)
	}

	// Load Flatcar actions
	query, _, err = goqu.From("flatcar_action").
		Where(goqu.C("package_id").In(pkgIDs)).
		ToSQL()
	if err != nil {
		return nil, err
	}

	var actions []FlatcarAction
	if err := api.db.Select(&actions, query); err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	actionsByPkg := make(map[string]*FlatcarAction)
	for i := range actions {
		actionsByPkg[actions[i].PackageID] = &actions[i]
	}

	// Assign files and actions to packages
	for _, pkg := range packages {
		if files, ok := filesByPkg[pkg.ID]; ok {
			pkg.ExtraFiles = files
		} else {
			pkg.ExtraFiles = nil
		}
		if action, ok := actionsByPkg[pkg.ID]; ok {
			pkg.FlatcarAction = action
		} else {
			pkg.FlatcarAction = nil
		}
	}

	return packages, nil
}

// packagesQuery returns a SelectDataset prepared to return all packages.
// This query is meant to be extended later in the methods using it to filter
// by a specific package id, all packages that belong to a given application,
// specify how to query the rows or their destination.
func (api *API) packagesQuery() *goqu.SelectDataset {
	// Note: semverToIntArray error handling is deferred to when ToSQL() is called
	// since goqu.SelectDataset doesn't support immediate error returns
	semverExpr, err := semverToIntArray("version")
	if err != nil {
		// Return an invalid query that will fail when ToSQL() is called
		return goqu.From("invalid_table_error_" + err.Error())
	}

	query := goqu.From(goqu.L("package LEFT JOIN package_channel_blacklist pcb ON package.id = pcb.package_id")).
		Select(goqu.L(`package.*,
	    array_agg(pcb.channel_id) FILTER (WHERE pcb.channel_id IS NOT NULL) as channels_blacklist
	    `)).
		GroupBy("package.id").Order(goqu.L(semverExpr).Desc())
	return query
}
func (api *API) getFlatcarActionQuery(packageID string) *goqu.SelectDataset {
	query := goqu.From("flatcar_action").Where(goqu.C("package_id").Eq(packageID))
	return query
}
func (api *API) getFlatcarAction(packageID string) (*FlatcarAction, error) {
	query, _, err := api.getFlatcarActionQuery(packageID).ToSQL()
	if err != nil {
		return nil, err
	}
	flatcarAction := FlatcarAction{}
	err = api.db.QueryRowx(query).StructScan(&flatcarAction)
	if err != nil {
		return nil, err
	}
	return &flatcarAction, nil
}

func (api *API) getExtraFiles(packageID string) ([]File, error) {
	query, _, err := goqu.From("package_file").Where(goqu.C("package_id").Eq(packageID)).Order(goqu.C("id").Asc()).ToSQL()
	if err != nil {
		return nil, err
	}

	var files []File
	rows, err := api.db.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		f := File{}
		if err := rows.StructScan(&f); err != nil {
			return nil, err
		}

		files = append(files, f)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return files, nil
}

func (api *API) getPackage(packageID null.String) (*Package, error) {
	query, _, err := api.packagesQuery().Where(goqu.C("id").Eq(packageID)).ToSQL()
	if err != nil {
		return nil, err
	}
	packageEntity := Package{}
	err = api.db.QueryRowx(query).StructScan(&packageEntity)
	if err != nil {
		return nil, err
	}
	flatcarAction, err := api.getFlatcarAction(packageEntity.ID)
	switch err {
	case nil:
		packageEntity.FlatcarAction = flatcarAction
	case sql.ErrNoRows:
		packageEntity.FlatcarAction = nil
	default:
		return nil, err
	}

	extraFiles, err := api.getExtraFiles(packageEntity.ID)
	switch err {
	case nil:
		packageEntity.ExtraFiles = extraFiles
	case sql.ErrNoRows:
		packageEntity.ExtraFiles = nil
	default:
		return nil, err
	}

	return &packageEntity, nil
}

// updatePackageBlacklistedChannels adds or removes as needed channels to the
// package's channels blacklist based on the new entries provided in the updated
// package entry.
//
// This method is part of the transaction that updates a package and when it's
// called, the package has already been updated except for the channels
// blacklist, that may happen here if needed.
func (api *API) updatePackageBlacklistedChannels(tx *sqlx.Tx, pkg *Package, oldPkg *Package) error {
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
		channel, err := api.GetChannel(channelID)
		if err != nil {
			return err
		}
		if channel.PackageID.String == pkg.ID {
			return ErrBlacklistingChannel
		}
		// Check if package is a floor for this channel
		isFloor, err := api.isPackageFloorForChannel(pkg.ID, channelID)
		if err != nil {
			return err
		}
		if isFloor {
			return ErrBlacklistingFloor
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
func (api *API) updatePackageFiles(tx *sqlx.Tx, pkg *Package, oldPkg *Package) error {
	var oldFiles map[int64]File

	if oldPkg != nil {
		oldFiles = make(map[int64]File, len(oldPkg.ExtraFiles))
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

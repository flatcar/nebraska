package api

import (
	"database/sql"
	"errors"
	"fmt"
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
)

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
}

// AddPackage registers the provided package.
func (api *API) AddPackage(pkg *Package) (*Package, error) {
	if !isValidSemver(pkg.Version) {
		return nil, ErrInvalidSemver
	}
	if !pkg.Arch.IsValid() {
		return nil, ErrInvalidArch
	}
	if len(pkg.ChannelsBlacklist) > 0 {
		blacklistedChannels, err := api.getSpecificChannels(pkg.ChannelsBlacklist...)
		if err != nil {
			return nil, err
		}
		for _, channel := range blacklistedChannels {
			if pkg.Arch != channel.Arch {
				return nil, ErrArchMismatch
			}
		}
	}

	tx, err := api.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			logger.Error().Err(err).Msg("AddPackage - could not roll back")
		}
	}()

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
		return nil, err
	}
	err = tx.QueryRowx(query).StructScan(pkg)
	if err != nil {
		return nil, err
	}
	if len(pkg.ChannelsBlacklist) > 0 {
		for _, channelID := range pkg.ChannelsBlacklist {
			query, _, err := goqu.Insert("package_channel_blacklist").
				Cols("package_id", "channel_id").
				Vals(goqu.Vals{pkg.ID, channelID}).
				ToSQL()
			if err != nil {
				return nil, err
			}
			_, err = tx.Exec(query)

			if err != nil {
				return nil, err
			}
		}
	}

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
		err = tx.QueryRowx(query).StructScan(flatcarAction)
		switch err {
		case nil:
			pkg.FlatcarAction = flatcarAction
		case sql.ErrNoRows:
			pkg.FlatcarAction = nil
		default:
			return nil, err
		}
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
			logger.Error().Err(err).Msg("UpdatePackage - could not roll back")
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

	if err := api.updatePackageBlacklistedChannels(tx, pkg); err != nil {
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
func (api *API) DeletePackage(pkgID string) error {
	query, _, err := goqu.Delete("package").
		Where(goqu.C("id").Eq(pkgID)).
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

// GetPackage returns the package identified by the id provided.
func (api *API) GetPackage(pkgID string) (*Package, error) {
	return api.getPackage(null.StringFrom(pkgID))
}

// GetPackageByVersionAndArch returns the package identified by the
// application ID, version and arch provided.
func (api *API) GetPackageByVersionAndArch(appID, version string, arch Arch) (*Package, error) {
	var pkg Package
	if !isValidSemver(version) {
		return nil, fmt.Errorf("Error GetPackageByVersionAndArch version %s is not valid", version)
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

// GetPackages returns all packages associated to the application provided.
func (api *API) GetPackages(appID string, page, perPage uint64) ([]*Package, error) {
	page, perPage = validatePaginationParams(page, perPage)
	limit, offset := sqlPaginate(page, perPage)
	query, _, err := api.packagesQuery().
		Where(goqu.C("application_id").Eq(appID)).
		Limit(limit).
		Offset(offset).
		ToSQL()
	if err != nil {
		return nil, err
	}
	return api.getPackagesFromQuery(query)
}

func (api *API) getPackagesFromQuery(query string) ([]*Package, error) {
	var pkgs []*Package
	rows, err := api.db.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var err error
		var flatcarAction *FlatcarAction
		pkg := Package{}

		err = rows.StructScan(&pkg)
		if err != nil {
			return nil, err
		}
		flatcarAction, err = api.getFlatcarAction(pkg.ID)
		switch err {
		case nil:
			pkg.FlatcarAction = flatcarAction
		case sql.ErrNoRows:
			pkg.FlatcarAction = nil
		default:
			return nil, err
		}
		pkgs = append(pkgs, &pkg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return pkgs, nil
}

// packagesQuery returns a SelectDataset prepared to return all packages.
// This query is meant to be extended later in the methods using it to filter
// by a specific package id, all packages that belong to a given application,
// specify how to query the rows or their destination.
func (api *API) packagesQuery() *goqu.SelectDataset {
	query := goqu.From(goqu.L("package LEFT JOIN package_channel_blacklist pcb ON package.id = pcb.package_id")).
		Select(goqu.L(`package.*,
	    array_agg(pcb.channel_id) FILTER (WHERE pcb.channel_id IS NOT NULL) as channels_blacklist
	    `)).
		GroupBy("package.id").Order(goqu.L("regexp_matches(version, '(\\d+)\\.(\\d+)\\.(\\d+)')::int[]").Desc())
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

	return &packageEntity, nil
}

// updatePackageBlacklistedChannels adds or removes as needed channels to the
// package's channels blacklist based on the new entries provided in the updated
// package entry.
//
// This method is part of the transaction that updates a package and when it's
// called, the package has already been updated except for the channels
// blacklist, that may happen here if needed.
func (api *API) updatePackageBlacklistedChannels(tx *sqlx.Tx, pkg *Package) error {
	pkgUpdated, err := api.GetPackage(pkg.ID)
	if err != nil {
		return err
	}

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

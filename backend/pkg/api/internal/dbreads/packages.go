package dbreads

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/types"
)

// GetPackage returns the package identified by the id provided.
func (q *Queries) GetPackage(pkgID string) (*types.Package, error) {
	return q.getPackage(null.StringFrom(pkgID))
}

// GetPackageByVersionAndArch returns the package identified by the
// application ID, version and arch provided.
func (q *Queries) GetPackageByVersionAndArch(appID, version string, arch types.Arch) (*types.Package, error) {
	var pkg types.Package
	if !IsValidSemver(version) {
		return nil, fmt.Errorf("error GetPackageByVersionAndArch version %s is not valid", version)
	}
	query, _, err := q.packagesQuery().
		Where(goqu.C("application_id").Eq(appID), goqu.C("arch").Eq(arch), goqu.C("version").Eq(version)).
		ToSQL()
	if err != nil {
		return nil, err
	}
	err = q.db.QueryRowx(query).StructScan(&pkg)
	if err != nil {
		return nil, err
	}
	flatcarAction, err := q.getFlatcarAction(pkg.ID)
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

// GetPackagesCount returns the total number of packages in an app
func (q *Queries) GetPackagesCount(appID string, searchVersion *string) (int, error) {
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
	return q.GetCountQuery(query)
}

// GetPackages returns all packages associated to the application provided.
func (q *Queries) GetPackages(appID string, page, perPage uint64, searchVersion *string) ([]*types.Package, error) {
	page, perPage = validatePaginationParams(page, perPage)
	limit, offset := sqlPaginate(page, perPage)
	query := q.packagesQuery().
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

	return q.getPackagesFromQuery(queryString)
}

func (q *Queries) getPackagesFromQuery(query string) ([]*types.Package, error) {
	var pkgs []*types.Package
	rows, err := q.db.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Load all packages
	for rows.Next() {
		pkg := types.Package{}
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
	return q.loadPackageExtras(pkgs)
}

// loadPackageExtras loads extra files and flatcar actions for packages efficiently
func (q *Queries) loadPackageExtras(packages []*types.Package) ([]*types.Package, error) {
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

	var files []types.File
	if err := q.db.Select(&files, query); err != nil {
		return nil, err
	}

	filesByPkg := make(map[string][]types.File)
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

	var actions []types.FlatcarAction
	if err := q.db.Select(&actions, query); err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	actionsByPkg := make(map[string]*types.FlatcarAction)
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
func (q *Queries) packagesQuery() *goqu.SelectDataset {
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
func (q *Queries) getFlatcarActionQuery(packageID string) *goqu.SelectDataset {
	query := goqu.From("flatcar_action").Where(goqu.C("package_id").Eq(packageID))
	return query
}
func (q *Queries) getFlatcarAction(packageID string) (*types.FlatcarAction, error) {
	query, _, err := q.getFlatcarActionQuery(packageID).ToSQL()
	if err != nil {
		return nil, err
	}
	flatcarAction := types.FlatcarAction{}
	err = q.db.QueryRowx(query).StructScan(&flatcarAction)
	if err != nil {
		return nil, err
	}
	return &flatcarAction, nil
}

func (q *Queries) getExtraFiles(packageID string) ([]types.File, error) {
	query, _, err := goqu.From("package_file").Where(goqu.C("package_id").Eq(packageID)).Order(goqu.C("id").Asc()).ToSQL()
	if err != nil {
		return nil, err
	}

	var files []types.File
	rows, err := q.db.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		f := types.File{}
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

func (q *Queries) getPackage(packageID null.String) (*types.Package, error) {
	query, _, err := q.packagesQuery().Where(goqu.C("id").Eq(packageID)).ToSQL()
	if err != nil {
		return nil, err
	}
	packageEntity := types.Package{}
	err = q.db.QueryRowx(query).StructScan(&packageEntity)
	if err != nil {
		return nil, err
	}
	flatcarAction, err := q.getFlatcarAction(packageEntity.ID)
	switch err {
	case nil:
		packageEntity.FlatcarAction = flatcarAction
	case sql.ErrNoRows:
		packageEntity.FlatcarAction = nil
	default:
		return nil, err
	}

	extraFiles, err := q.getExtraFiles(packageEntity.ID)
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

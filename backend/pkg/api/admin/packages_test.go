package admin

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/api"
	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

func TestAddPackageDuplicate(t *testing.T) {
	a, err := api.NewForTest(api.OptionInitDB)
	require.NoError(t, err)
	defer a.Close()
	svc := NewService(a.Conn(), a.Reads())

	tTeam, _ := svc.AddTeam(&types.Team{Name: "test_team"})
	tApp, _ := svc.AddApp(&types.Application{Name: "test_app", TeamID: tTeam.ID})

	_, err = svc.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID, Arch: types.ArchAArch64})
	assert.NoError(t, err)

	// Same application, version and arch violates package_appid_version_arch_unique.
	// A different URL still collides, the constraint does not cover it.
	_, err = svc.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/other", Version: "12.1.0", ApplicationID: tApp.ID, Arch: types.ArchAArch64})
	assert.ErrorIs(t, err, types.ErrDuplicatePackage)

	// The syncer reaches the same insert through AddPackageWithMetadata.
	_, err = svc.AddPackageWithMetadata(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID, Arch: types.ArchAArch64})
	assert.ErrorIs(t, err, types.ErrDuplicatePackage)

	// A different arch or version is not a duplicate.
	_, err = svc.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID, Arch: types.ArchX86})
	assert.NoError(t, err)

	_, err = svc.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.2.0", ApplicationID: tApp.ID, Arch: types.ArchAArch64})
	assert.NoError(t, err)
}

// A blacklist conflict is a different unique constraint and must not be
// reported as a duplicate package.
func TestAddPackageBlacklistConflictIsNotDuplicate(t *testing.T) {
	a, err := api.NewForTest(api.OptionInitDB)
	require.NoError(t, err)
	defer a.Close()
	svc := NewService(a.Conn(), a.Reads())

	tTeam, _ := svc.AddTeam(&types.Team{Name: "test_team"})
	tApp, _ := svc.AddApp(&types.Application{Name: "test_app", TeamID: tTeam.ID})
	tChannel, err := svc.AddChannel(&types.Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, Arch: types.ArchAArch64})
	require.NoError(t, err)

	_, err = svc.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0",
		ApplicationID: tApp.ID, Arch: types.ArchAArch64, ChannelsBlacklist: []string{tChannel.ID, tChannel.ID}})
	assert.Error(t, err)
	assert.NotErrorIs(t, err, types.ErrDuplicatePackage)
}

func TestIsUniqueViolation(t *testing.T) {
	assert.True(t, isUniqueViolation(&pgconn.PgError{Code: pgUniqueViolation}))
	assert.True(t, isUniqueViolation(fmt.Errorf("insert failed: %w", &pgconn.PgError{Code: pgUniqueViolation})))

	// Foreign key violation, not a duplicate.
	assert.False(t, isUniqueViolation(&pgconn.PgError{Code: "23503"}))
	assert.False(t, isUniqueViolation(errors.New("connection refused")))
	assert.False(t, isUniqueViolation(nil))
}

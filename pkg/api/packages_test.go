package api

import (
	"testing"

	"gopkg.in/guregu/null.v4"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAddPackage(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tChannel1, _ := a.AddChannel(&Channel{Name: "test_channel1", Color: "blue", ApplicationID: tApp.ID, Arch: ArchAArch64})
	tChannel2, _ := a.AddChannel(&Channel{Name: "test_channel2", Color: "green", ApplicationID: tApp.ID, Arch: ArchAArch64})

	pkg, err := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID, ChannelsBlacklist: []string{tChannel1.ID, tChannel2.ID}, Arch: ArchAArch64})
	assert.NoError(t, err)

	_, err = a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID, Arch: ArchX86})
	assert.NoError(t, err)

	pkgX, err := a.GetPackage(pkg.ID)
	assert.NoError(t, err)
	assert.Equal(t, PkgTypeOther, pkgX.Type)
	assert.Equal(t, "http://sample.url/pkg", pkgX.URL)
	assert.Equal(t, "12.1.0", pkgX.Version)
	assert.Equal(t, tApp.ID, pkgX.ApplicationID)
	assert.Contains(t, pkgX.ChannelsBlacklist, tChannel1.ID)
	assert.Contains(t, pkgX.ChannelsBlacklist, tChannel2.ID)
	assert.Equal(t, ArchAArch64, pkgX.Arch)

	_, err = a.AddPackage(&Package{URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	assert.Error(t, err, "Package type is required.")

	_, err = a.AddPackage(&Package{Type: PkgTypeOther, Version: "12.1.0", ApplicationID: tApp.ID})
	assert.Error(t, err, "Package url is required.")

	_, err = a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", ApplicationID: tApp.ID})
	assert.Error(t, err, "Package version is required.")

	_, err = a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "aaa12.1.0"})
	assert.Equal(t, ErrInvalidSemver, err, "Package version must be a valid semver.")

	_, err = a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0"})
	assert.Error(t, err, "App id is required and must be a valid uuid.")

	_, err = a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID, ChannelsBlacklist: []string{uuid.New().String()}})
	assert.Error(t, err, "Blacklisted channels must be existing channels ids.")

	_, err = a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID, ChannelsBlacklist: []string{"invalidChannelID"}})
	assert.Error(t, err, "Blacklisted channels must be valid existing channels ids.")

	_, err = a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID, ChannelsBlacklist: []string{tChannel1.ID}})
	assert.Error(t, err, "Blacklisted channels must have a matching arch.")

	_, err = a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID, Arch: Arch(77777)})
	assert.Error(t, err, "Arch must be a valid architecture")

	_, err = a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID, Arch: Arch(77777)})
	assert.Error(t, err, "Arch must be a valid architecture")
}

func TestAddPackageFlatcar(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	pkg := &Package{
		Type:          PkgTypeFlatcar,
		URL:           "https://commondatastorage.googleapis.com/update-storage.core-os.net/amd64-usr/766.3.0/",
		Filename:      null.StringFrom("update.gz"),
		Version:       "2016.6.6",
		Size:          null.StringFrom("123456"),
		Hash:          null.StringFrom("sha1:blablablabla"),
		ApplicationID: flatcarAppID,
		FlatcarAction: &FlatcarAction{
			Sha256: "sha256:blablablabla",
		},
	}
	_, err := a.AddPackage(pkg)
	assert.NoError(t, err)
	assert.Equal(t, "postinstall", pkg.FlatcarAction.Event)
	assert.Equal(t, false, pkg.FlatcarAction.NeedsAdmin)
	assert.Equal(t, false, pkg.FlatcarAction.IsDelta)
	assert.Equal(t, true, pkg.FlatcarAction.DisablePayloadBackoff)
	assert.Equal(t, "sha256:blablablabla", pkg.FlatcarAction.Sha256)
}

func TestUpdatePackage(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tChannel1, _ := a.AddChannel(&Channel{Name: "test_channel1", Color: "blue", ApplicationID: tApp.ID})
	tPkg, err := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID, ChannelsBlacklist: []string{tChannel1.ID}})
	assert.NoError(t, err)

	tChannel2, _ := a.AddChannel(&Channel{Name: "test_channel2", Color: "green", ApplicationID: tApp.ID})
	tChannel3, _ := a.AddChannel(&Channel{Name: "test_channel3", Color: "red", ApplicationID: tApp.ID})
	tChannel4, _ := a.AddChannel(&Channel{Name: "test_channel4", Color: "yellow", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})

	err = a.UpdatePackage(&Package{ID: tPkg.ID, Type: PkgTypeOther, URL: "http://sample.url/pkg_updated", Version: "12.2.0", ChannelsBlacklist: []string{tChannel2.ID, tChannel3.ID}})
	assert.NoError(t, err)

	pkg, err := a.GetPackage(tPkg.ID)
	assert.NoError(t, err)
	assert.Equal(t, "http://sample.url/pkg_updated", pkg.URL)
	assert.Equal(t, "12.2.0", pkg.Version)
	assert.NotContains(t, pkg.ChannelsBlacklist, tChannel1.ID)
	assert.Contains(t, pkg.ChannelsBlacklist, tChannel2.ID)
	assert.Contains(t, pkg.ChannelsBlacklist, tChannel3.ID)

	err = a.UpdatePackage(&Package{ID: tPkg.ID, Type: PkgTypeOther, URL: "http://sample.url/pkg_updated", Version: "12.2.0", ChannelsBlacklist: []string{tChannel4.ID}})
	assert.Equal(t, ErrBlacklistingChannel, err)

	err = a.UpdatePackage(&Package{ID: tPkg.ID, Type: PkgTypeOther, URL: "http://sample.url/pkg_updated", Version: "12.2.0", ChannelsBlacklist: nil, Arch: ArchAArch64})
	assert.NoError(t, err)
	pkg, _ = a.GetPackage(tPkg.ID)
	assert.Len(t, pkg.ChannelsBlacklist, 0)
	// can't change an arch of a package
	assert.Equal(t, ArchAll, pkg.Arch)
}

func TestUpdatePackageFlatcar(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	pkg := &Package{
		Type:          PkgTypeFlatcar,
		URL:           "https://commondatastorage.googleapis.com/update-storage.core-os.net/amd64-usr/766.3.0/",
		Filename:      null.StringFrom("update.gz"),
		Version:       "2016.6.6",
		Size:          null.StringFrom("123456"),
		Hash:          null.StringFrom("sha1:blablablabla"),
		ApplicationID: flatcarAppID,
	}
	pkg, err := a.AddPackage(pkg)
	assert.NoError(t, err)
	assert.Nil(t, pkg.FlatcarAction)
	pkg.Version = "2016.6.7"
	err = a.UpdatePackage(pkg)
	assert.NoError(t, err)
	assert.Nil(t, pkg.FlatcarAction)

	pkg.FlatcarAction = &FlatcarAction{
		Sha256: "sha256:blablablabla",
	}
	err = a.UpdatePackage(pkg)
	assert.NoError(t, err)
	assert.Equal(t, "postinstall", pkg.FlatcarAction.Event)
	assert.Equal(t, false, pkg.FlatcarAction.NeedsAdmin)
	assert.Equal(t, false, pkg.FlatcarAction.IsDelta)
	assert.Equal(t, true, pkg.FlatcarAction.DisablePayloadBackoff)
	assert.Equal(t, "sha256:blablablabla", pkg.FlatcarAction.Sha256)

	err = a.DeletePackage(pkg.ID)
	assert.NoError(t, err)

	pkg = &Package{
		Type:          PkgTypeFlatcar,
		URL:           "https://commondatastorage.googleapis.com/update-storage.core-os.net/amd64-usr/766.3.0/",
		Filename:      null.StringFrom("update.gz"),
		Version:       "2016.6.6",
		Size:          null.StringFrom("123456"),
		Hash:          null.StringFrom("sha1:blablablabla"),
		ApplicationID: flatcarAppID,
	}
	pkg.FlatcarAction = &FlatcarAction{
		Sha256: "sha256:blablablabla",
	}
	pkg, err = a.AddPackage(pkg)
	assert.NoError(t, err)
	assert.NotEqual(t, pkg.FlatcarAction.ID, "")

	flatcarActionID := pkg.FlatcarAction.ID
	pkg.FlatcarAction.Sha256 = "sha256:bleblebleble"
	err = a.UpdatePackage(pkg)
	assert.NoError(t, err)
	assert.Equal(t, "sha256:bleblebleble", pkg.FlatcarAction.Sha256)
	assert.Equal(t, flatcarActionID, pkg.FlatcarAction.ID)
}

func TestDeletePackage(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, err := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	assert.NoError(t, err)

	err = a.DeletePackage(tPkg.ID)
	assert.NoError(t, err)

	_, err = a.GetPackage(tPkg.ID)
	assert.Error(t, err, "Trying to get deleted package.")

	err = a.DeletePackage("invalidPackageID")
	assert.Error(t, err, "Package id must be a valid uuid.")
}

func TestGetPackage(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel1", Color: "blue", ApplicationID: tApp.ID})
	tPkg, err := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID, ChannelsBlacklist: []string{tChannel.ID}})
	assert.NoError(t, err)

	pkg, err := a.GetPackage(tPkg.ID)
	assert.NoError(t, err)
	assert.Equal(t, PkgTypeOther, pkg.Type)
	assert.Equal(t, "http://sample.url/pkg", pkg.URL)
	assert.Equal(t, "12.1.0", pkg.Version)
	assert.Equal(t, tApp.ID, pkg.ApplicationID)
	assert.Equal(t, StringArray([]string{tChannel.ID}), pkg.ChannelsBlacklist)
	assert.Equal(t, ArchAll, pkg.Arch)

	_, err = a.GetPackage("invalidPackageID")
	assert.Error(t, err, "Package id must be a valid uuid.")

	_, err = a.GetPackage(uuid.New().String())
	assert.Error(t, err, "Package id must exist.")
}

func TestGetPackageByVersionAndArch(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, err := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	assert.NoError(t, err)
	tPkgARM, err := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "13.2.1", ApplicationID: tApp.ID, Arch: ArchAArch64})
	assert.NoError(t, err)

	pkg, err := a.GetPackageByVersionAndArch(tApp.ID, tPkg.Version, ArchAll)
	assert.NoError(t, err)
	assert.Equal(t, PkgTypeOther, pkg.Type)
	assert.Equal(t, "http://sample.url/pkg", pkg.URL)
	assert.Equal(t, "12.1.0", pkg.Version)
	assert.Equal(t, tApp.ID, pkg.ApplicationID)
	assert.Equal(t, ArchAll, pkg.Arch)

	_, err = a.GetPackageByVersionAndArch("invalidAppID", "12.1.0", ArchAll)
	assert.Error(t, err, "Application id must be a valid uuid.")

	_, err = a.GetPackageByVersionAndArch(uuid.New().String(), "12.1.0", ArchAll)
	assert.Error(t, err, "Application id must exist.")

	_, err = a.GetPackageByVersionAndArch(tApp.ID, "hola", ArchAll)
	assert.Error(t, err, "Version must be a valid semver value.")

	_, err = a.GetPackageByVersionAndArch(tApp.ID, tPkgARM.Version, ArchAll)
	assert.Error(t, err, "Shouldn't pick the ARM version")

	pkg, err = a.GetPackageByVersionAndArch(tApp.ID, tPkgARM.Version, ArchAArch64)
	assert.NoError(t, err)
	assert.Equal(t, ArchAArch64, pkg.Arch)
}

func TestGetPackages(t *testing.T) {
	a := newForTest(t)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	_, _ = a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg1", Version: "1010.5.0+2016-05-27-1832", ApplicationID: tApp.ID, Arch: ArchAMD64})
	_, _ = a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg2", Version: "12.1.0", ApplicationID: tApp.ID, Arch: ArchX86})
	_, _ = a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg3", Version: "14.1.0", ApplicationID: tApp.ID, Arch: ArchAArch64})
	_, _ = a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg4", Version: "1010.6.0-blabla", ApplicationID: tApp.ID})

	pkgs, err := a.GetPackages(tApp.ID, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(pkgs))
	assert.Equal(t, "http://sample.url/pkg4", pkgs[0].URL)
	assert.Equal(t, "http://sample.url/pkg1", pkgs[1].URL)
	assert.Equal(t, "http://sample.url/pkg3", pkgs[2].URL)
	assert.Equal(t, "http://sample.url/pkg2", pkgs[3].URL)

	assert.Equal(t, ArchAll, pkgs[0].Arch)
	assert.Equal(t, ArchAMD64, pkgs[1].Arch)
	assert.Equal(t, ArchAArch64, pkgs[2].Arch)
	assert.Equal(t, ArchX86, pkgs[3].Arch)

	_, err = a.GetPackages("invalidAppID", 0, 0)
	assert.Error(t, err, "Add id must be a valid uuid.")

	_, err = a.GetPackages(uuid.New().String(), 0, 0)
	assert.NoError(t, err, "should be no error for non existing appID")
}

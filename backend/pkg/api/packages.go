package api

import "github.com/flatcar/nebraska/backend/pkg/api/internal/types"

const (
	// PkgTypeFlatcar indicates that the package is a Flatcar update package
	PkgTypeFlatcar = types.PkgTypeFlatcar

	// PkgTypeDocker indicates that the package is a Docker container
	PkgTypeDocker = types.PkgTypeDocker

	// PkgTypeRocket indicates that the package is a Rocket container
	PkgTypeRocket = types.PkgTypeRocket

	// PkgTypeOther is the generic package type.
	PkgTypeOther = types.PkgTypeOther
)

type (
	File                = types.File
	Package             = types.Package
	ChannelPackageFloor = types.ChannelPackageFloor
	StringArray         = types.StringArray
)

var (
	// ErrBlacklistingChannel error indicates that the channel the package is
	// trying to blacklist is already pointing to the package.
	ErrBlacklistingChannel = types.ErrBlacklistingChannel

	// ErrPackageIsFloor error indicates that the package cannot be deleted
	// because it is marked as a floor version for one or more channels.
	ErrPackageIsFloor = types.ErrPackageIsFloor

	// ErrBlacklistingFloor error indicates that the package cannot be blacklisted
	// because it is marked as a floor version for the channel.
	ErrBlacklistingFloor = types.ErrBlacklistingFloor
)

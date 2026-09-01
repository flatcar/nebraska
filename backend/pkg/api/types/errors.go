package types

import "errors"

// ErrNoRowsAffected indicates that no rows were affected in an update or
// delete database operation.
var ErrNoRowsAffected = errors.New("nebraska: no rows affected")

// ErrInvalidSemver indicates that the provided semver version is not valid.
var ErrInvalidSemver = errors.New("nebraska: invalid semver")

// ErrNoPackageFound indicates that the group doesn't have a channel
// assigned or that the channel doesn't have a package assigned.
var ErrNoPackageFound = errors.New("nebraska: no package found")

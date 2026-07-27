package types

import "errors"

// ErrNoRowsAffected indicates that no rows were affected in an update or
// delete database operation.
var ErrNoRowsAffected = errors.New("nebraska: no rows affected")

// ErrInvalidSemver indicates that the provided semver version is not valid.
var ErrInvalidSemver = errors.New("nebraska: invalid semver")

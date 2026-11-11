package dbreads

import (
	"github.com/blang/semver/v4"
	"github.com/doug-martin/goqu/v9"

	// Registers the postgres dialect. Without it goqu.Dialect falls back to
	// the default dialect and emits ? placeholders, which PostgreSQL rejects.
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/google/uuid"
)

const (
	defaultPage    uint64 = 1
	defaultPerPage uint64 = 10
)

// PGDialect builds statements with numbered placeholders, for queries whose
// arguments are bound rather than inlined.
var PGDialect = goqu.Dialect("postgres")

// validatePaginationParams validates the pagination parameters provided,
// setting them to the default values in case they are invalid.
func validatePaginationParams(page, perPage uint64) (uint64, uint64) {
	if page < 1 {
		page = defaultPage
	}

	if perPage < 1 {
		perPage = defaultPerPage
	}

	return page, perPage
}

// IsValidSemver checks if the provided string represents a valid semver
// version.
func IsValidSemver(version string) bool {
	if _, err := semver.Make(version); err != nil {
		return false
	}
	return true
}

// isCanonicalUUID reports whether s is a UUID in its 36-character form.
func isCanonicalUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	_, err := uuid.Parse(s)
	return err == nil
}

func sqlPaginate(page, perPage uint64) (uint, uint) {
	return uint(perPage), uint(page-1) * uint(perPage)
}

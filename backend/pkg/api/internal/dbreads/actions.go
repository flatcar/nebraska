package dbreads

import (
	"github.com/doug-martin/goqu/v9"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/types"
)

// GetFlatcarAction returns the Flatcar action entry associated to the package id
// provided.
func (q *Queries) GetFlatcarAction(packageID string) (*types.FlatcarAction, error) {
	action := types.FlatcarAction{}
	query, _, err := goqu.From("flatcar_action").
		Where(goqu.C("package_id").Eq(packageID)).ToSQL()
	if err != nil {
		return nil, err
	}
	err = q.db.QueryRowx(query).StructScan(&action)
	if err != nil {
		return nil, err
	}
	return &action, nil
}

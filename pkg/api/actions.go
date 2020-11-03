package api

import (
	"time"

	"github.com/doug-martin/goqu/v9"
)

// FlatcarAction represents an Omaha action with some Flatcar specific fields.
type FlatcarAction struct {
	ID                    string    `db:"id" json:"id"`
	Event                 string    `db:"event" json:"event"`
	ChromeOSVersion       string    `db:"chromeos_version" json:"chromeos_version"`
	Sha256                string    `db:"sha256" json:"sha256"`
	NeedsAdmin            bool      `db:"needs_admin" json:"needs_admin"`
	IsDelta               bool      `db:"is_delta" json:"is_delta"`
	DisablePayloadBackoff bool      `db:"disable_payload_backoff" json:"disable_payload_backoff"`
	MetadataSignatureRsa  string    `db:"metadata_signature_rsa" json:"metadata_signature_rsa"`
	MetadataSize          string    `db:"metadata_size" json:"metadata_size"`
	Deadline              string    `db:"deadline" json:"deadline"`
	CreatedTs             time.Time `db:"created_ts" json:"created_ts"`
	PackageID             string    `db:"package_id" json:"-"`
}

// AddFlatcarAction registers the provided Omaha Flatcar action.
func (api *API) AddFlatcarAction(action *FlatcarAction) (*FlatcarAction, error) {
	query, _, err := goqu.Insert("flatcar_action").
		Cols("event", "chromeos_version", "sha256", "needs_admin", "is_delta", "disable_payload_backoff", "metadata_signature_rsa", "metadata_size", "deadline", "package_id").
		Vals(goqu.Vals{
			action.Event,
			action.ChromeOSVersion,
			action.Sha256,
			action.NeedsAdmin,
			action.IsDelta,
			action.DisablePayloadBackoff,
			action.MetadataSignatureRsa,
			action.MetadataSize,
			action.Deadline,
			action.PackageID,
		}).
		Returning(goqu.T("flatcar_action").All()).
		ToSQL()
	if err != nil {
		return nil, err
	}
	err = api.readDb.QueryRowx(query).StructScan(action)
	if err != nil {
		return nil, err
	}
	return action, err
}

// GetFlatcarAction returns the Flatcar action entry associated to the package id
// provided.
func (api *API) GetFlatcarAction(packageID string) (*FlatcarAction, error) {
	action := FlatcarAction{}
	query, _, err := goqu.From("flatcar_action").
		Where(goqu.C("package_id").Eq(packageID)).ToSQL()
	if err != nil {
		return nil, err
	}
	err = api.readDb.QueryRowx(query).StructScan(&action)
	if err != nil {
		return nil, err
	}
	return &action, nil
}

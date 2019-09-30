package api

import "time"

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
	err := api.dbR.
		InsertInto("flatcar_action").
		Whitelist("event", "chromeos_version", "sha256", "needs_admin", "is_delta", "disable_payload_backoff", "metadata_signature_rsa", "metadata_size", "deadline", "package_id").
		Record(action).
		Returning("*").
		QueryStruct(action)

	return action, err
}

// GetFlatcarAction returns the Flatcar action entry associated to the package id
// provided.
func (api *API) GetFlatcarAction(packageID string) (*FlatcarAction, error) {
	var action FlatcarAction

	err := api.dbR.SelectDoc("*").
		From("flatcar_action").
		Where("package_id = $1", packageID).
		QueryStruct(&action)

	if err != nil {
		return nil, err
	}

	return &action, nil
}

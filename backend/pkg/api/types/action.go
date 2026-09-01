package types

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

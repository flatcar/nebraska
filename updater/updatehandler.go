package updater

import (
	"context"
)

// UpdateHandler is an interface that wraps the FetchUpdate and
// ApplyUpdate command. Both FetchUpdate and ApplyUpdate take
// context and UpdateInfo as args and must return non-nil
// error if fetching or applying the update fails.
type UpdateHandler interface {
	FetchUpdate(ctx context.Context, info UpdateInfo) error
	ApplyUpdate(ctx context.Context, info UpdateInfo) error
}

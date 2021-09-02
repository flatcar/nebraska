package updater

import (
	"context"
)

type UpdateHandler interface {
	FetchUpdate(ctx context.Context, info *UpdateInfo) error
	ApplyUpdate(ctx context.Context, info *UpdateInfo) error
}

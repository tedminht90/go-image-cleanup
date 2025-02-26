package cleanup

import (
	"context"
)

type CleanupUseCase interface {
	Cleanup(ctx context.Context) error
}

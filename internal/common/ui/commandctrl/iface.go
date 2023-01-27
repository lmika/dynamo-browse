package commandctrl

import (
	"context"
	"github.com/lmika/audax/internal/dynamo-browse/services"
)

type IterProvider interface {
	Iter(ctx context.Context, category string) services.HistoryProvider
}

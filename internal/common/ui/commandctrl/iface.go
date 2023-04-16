package commandctrl

import (
	"context"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/services"
)

type IterProvider interface {
	Iter(ctx context.Context, category string) services.HistoryProvider
}

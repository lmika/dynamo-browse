package inputhistory

import "context"

type HistoryItemStore interface {
	Items(ctx context.Context, category string) ([]string, error)
	PutItem(ctx context.Context, category string, item string) error
}

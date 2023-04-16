package inputhistorystore

import (
	"context"
	"github.com/asdine/storm"
	"github.com/lmika/dynamo-browse/internal/common/sliceutils"
	"github.com/lmika/dynamo-browse/internal/common/workspaces"
	"github.com/pkg/errors"
	"sort"
	"time"
)

const inputHistoryStore = "InputHistoryStore"

type Store struct {
	ws storm.Node
}

func NewInputHistoryStore(ws *workspaces.Workspace) *Store {
	return &Store{
		ws: ws.DB().From(inputHistoryStore),
	}
}

// Items returns all items from a category ordered by the time
func (as *Store) Items(ctx context.Context, category string) ([]string, error) {
	var items []inputHistoryItem
	if err := as.ws.Find("Category", category, &items); err != nil {
		if errors.Is(err, storm.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Time.Before(items[j].Time)
	})
	return sliceutils.Map(items, func(t inputHistoryItem) string {
		return t.Item
	}), nil
}

func (as *Store) PutItem(ctx context.Context, category string, item string) error {
	return as.ws.Save(&inputHistoryItem{
		Time:     time.Now(),
		Category: category,
		Item:     item,
	})
}

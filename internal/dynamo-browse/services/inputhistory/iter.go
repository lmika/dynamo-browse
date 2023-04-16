package inputhistory

import (
	"context"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/services"
	"log"
	"strings"
)

func (svc *Service) Iter(ctx context.Context, category string) services.HistoryProvider {
	items, err := svc.store.Items(ctx, category)
	if err != nil {
		log.Printf("warn: cannot get iter for '%v': %v", category, err)
		return nil
	}
	return &Iter{svc, items, category}
}

func (svc *Service) PutItem(ctx context.Context, category string, value string) error {
	return svc.store.PutItem(ctx, category, value)
}

type Iter struct {
	svc      *Service
	items    []string
	category string
}

func (i *Iter) Len() int {
	return len(i.items)
}

func (i *Iter) Item(idx int) string {
	return i.items[idx]
}

func (i *Iter) PutItem(item string) {
	if strings.TrimSpace(item) == "" {
		return
	}

	if len(i.items) > 0 && i.items[len(i.items)-1] == item {
		return
	}

	if err := i.svc.PutItem(context.Background(), i.category, item); err != nil {
		log.Printf("warn: cannot put input history: category = %v, value = %v, err = %v", i.category, item, err)
	}
}

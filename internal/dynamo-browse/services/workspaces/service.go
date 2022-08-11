package workspaces

import (
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/models/serialisable"
	"github.com/pkg/errors"
	"time"
)

type Service struct {
	store ResultSetSnapshotStore
}

func NewService(store ResultSetSnapshotStore) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) PushSnapshot(rs *models.ResultSet) error {
	newSnapshot := &serialisable.ResultSetSnapshot{
		Time:      time.Now(),
		TableInfo: rs.TableInfo,
	}
	if q := rs.Query; q != nil {
		newSnapshot.Query.Expression = q.String()
	}

	if head, err := s.store.Head(); head != nil {
		newSnapshot.BackLink = head.ID
	} else if err != nil {
		return errors.Wrap(err, "cannot get head result set")
	}

	if err := s.store.Save(newSnapshot); err != nil {
		return errors.Wrap(err, "cannot save snapshot")
	}
	if err := s.store.SetAsHead(newSnapshot.ID); err != nil {
		return errors.Wrap(err, "cannot set new snapshot as head")
	}

	return nil
}

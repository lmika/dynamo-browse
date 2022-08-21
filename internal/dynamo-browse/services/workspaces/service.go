package workspaces

import (
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/models/serialisable"
	"github.com/pkg/errors"
	"log"
	"time"
)

type ViewSnapshotService struct {
	store ViewSnapshotStore
}

func NewService(store ViewSnapshotStore) *ViewSnapshotService {
	return &ViewSnapshotService{
		store: store,
	}
}

func (s *ViewSnapshotService) PushSnapshot(rs *models.ResultSet, filter string) error {
	newSnapshot := &serialisable.ViewSnapshot{
		Time:      time.Now(),
		TableName: rs.TableInfo.Name,
	}
	if q := rs.Query; q != nil {
		newSnapshot.Query = q.String()
	}
	newSnapshot.Filter = filter

	head, err := s.store.Head()
	if err != nil {
		return errors.Wrap(err, "cannot get head result set")
	}

	if head != nil {
		if newSnapshot.IsSameView(head) {
			// Duplicate
			return nil
		}

		newSnapshot.BackLink = head.ID
	}

	if err := s.store.Save(newSnapshot); err != nil {
		return errors.Wrap(err, "cannot save snapshot")
	}
	if err := s.store.SetAsHead(newSnapshot.ID); err != nil {
		return errors.Wrap(err, "cannot set new snapshot as head")
	}

	return nil
}

func (s *ViewSnapshotService) PopSnapshot() (*serialisable.ViewSnapshot, error) {
	vs, err := s.store.Head()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get snapshot head")
	} else if vs == nil {
		return nil, nil
	}

	if vs.BackLink == 0 {
		return nil, nil
	}

	if err := s.store.SetAsHead(vs.BackLink); err != nil {
		return nil, errors.Wrap(err, "cannot set new head")
	}
	if err := s.store.Remove(vs.ID); err != nil {
		return nil, errors.Wrap(err, "cannot remove old ID")
	}

	vs, err = s.store.Head()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get snapshot head")
	} else if vs == nil {
		return nil, nil
	}

	log.Printf("returning backstack: table='%v', query='%v', filter='%v'", vs.TableName, vs.Query, vs.Filter)
	return vs, nil
}

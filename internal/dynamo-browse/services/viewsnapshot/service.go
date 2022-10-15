package viewsnapshot

import (
	"github.com/lmika/audax/internal/dynamo-browse/models/serialisable"
	"github.com/pkg/errors"
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

func (s *ViewSnapshotService) PushSnapshot(details serialisable.ViewSnapshotDetails) error {
	newSnapshot := &serialisable.ViewSnapshot{
		Time:    time.Now(),
		Details: details,
	}

	oldHead, err := s.store.CurrentlyViewedSnapshot()
	if err != nil {
		return errors.Wrap(err, "cannot get snapshot head")
	}

	if oldHead != nil && oldHead.Details == details {
		// Attempting to push a duplicate
		return nil
	}

	if oldHead != nil {
		newSnapshot.BackLink = oldHead.ID

		// Remove all nodes from this point on the head
		if err := s.store.Dehead(oldHead); err != nil {
			return errors.Wrap(err, "cannot remove head")
		}
	}

	if err := s.store.Save(newSnapshot); err != nil {
		return errors.Wrap(err, "cannot save snapshot")
	}

	if oldHead != nil {
		oldHead.ForeLink = newSnapshot.ID
		if err := s.store.Save(oldHead); err != nil {
			return errors.Wrap(err, "cannot update old head")
		}
	}

	if err := s.store.SetAsHead(newSnapshot.ID); err != nil {
		return errors.Wrap(err, "cannot set new snapshot as head")
	}
	if err := s.store.SetCurrentlyViewedSnapshot(newSnapshot.ID); err != nil {
		return errors.Wrap(err, "cannot set new snapshot as head")
	}

	return nil
}

func (s *ViewSnapshotService) Len() (int, error) {
	return s.store.Len()
}

func (s *ViewSnapshotService) ViewRestore() (*serialisable.ViewSnapshot, error) {
	vs, err := s.store.CurrentlyViewedSnapshot()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get snapshot head")
	}
	return vs, nil
}

func (s *ViewSnapshotService) ViewBack() (*serialisable.ViewSnapshot, error) {
	vs, err := s.store.CurrentlyViewedSnapshot()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get snapshot head")
	} else if vs == nil || vs.BackLink == 0 {
		return nil, nil
	}

	vsToReturn, err := s.store.Find(vs.BackLink)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get snapshot head")
	} else if vsToReturn == nil {
		return nil, nil
	}

	if err := s.store.SetCurrentlyViewedSnapshot(vsToReturn.ID); err != nil {
		return nil, errors.Wrap(err, "cannot set new head")
	}

	return vsToReturn, nil
}

func (s *ViewSnapshotService) ViewForward() (*serialisable.ViewSnapshot, error) {
	vs, err := s.store.CurrentlyViewedSnapshot()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get snapshot head")
	} else if vs == nil || vs.ForeLink == 0 {
		return nil, nil
	}

	vsToReturn, err := s.store.Find(vs.ForeLink)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get snapshot head")
	} else if vsToReturn == nil {
		return nil, nil
	}

	if err := s.store.SetCurrentlyViewedSnapshot(vsToReturn.ID); err != nil {
		return nil, errors.Wrap(err, "cannot set new head")
	}

	return vsToReturn, nil
}

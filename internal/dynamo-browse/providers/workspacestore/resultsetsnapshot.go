package workspacestore

import (
	"github.com/asdine/storm"
	"github.com/lmika/audax/internal/common/workspaces"
	"github.com/lmika/audax/internal/dynamo-browse/models/serialisable"
	"github.com/pkg/errors"
	"log"
)

const resultSetSnapshotsBucket = "ResultSetSnapshots"

type ResultSetSnapshotStore struct {
	ws storm.Node
}

func NewResultSetSnapshotStore(ws *workspaces.Workspace) *ResultSetSnapshotStore {
	return &ResultSetSnapshotStore{
		ws: ws.DB().From(resultSetSnapshotsBucket),
	}
}

func (s *ResultSetSnapshotStore) Save(rs *serialisable.ResultSetSnapshot) error {
	if err := s.ws.Save(rs); err != nil {
		return errors.Wrap(err, "cannot save result set")
	}
	log.Printf("saved result set")
	return nil
}

func (s *ResultSetSnapshotStore) SetAsHead(resultSetID int64) error {
	if err := s.ws.Set("head", "id", resultSetID); err != nil {
		return errors.Wrap(err, "cannot set as head")
	}
	log.Printf("saved result set head")
	return nil
}

func (s *ResultSetSnapshotStore) Head() (*serialisable.ResultSetSnapshot, error) {
	var headResultSetID int64
	if err := s.ws.Get("head", "id", &headResultSetID); err != nil && !errors.Is(err, storm.ErrNotFound) {
		return nil, errors.Wrap(err, "cannot get head")
	}

	var rss serialisable.ResultSetSnapshot
	if err := s.ws.One("ID", headResultSetID, &rss); err != nil {
		if errors.Is(err, storm.ErrNotFound) {
			return nil, nil
		} else {
			return nil, errors.Wrap(err, "cannot get head")
		}
	}

	return &rss, nil
}

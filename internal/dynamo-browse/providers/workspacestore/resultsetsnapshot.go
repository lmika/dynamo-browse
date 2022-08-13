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

func (s *ResultSetSnapshotStore) Save(rs *serialisable.ViewSnapshot) error {
	if err := s.ws.Save(rs); err != nil {
		return errors.Wrap(err, "cannot save result set")
	}
	log.Printf("saved result set: table='%v', query='%v', filter='%v'", rs.TableName, rs.Query, rs.Filter)
	return nil
}

func (s *ResultSetSnapshotStore) SetAsHead(resultSetID int64) error {
	if resultSetID == 0 {
		if err := s.ws.Delete("head", "id"); err != nil {
			return errors.Wrap(err, "cannot remove head")
		}
		return nil
	}

	if err := s.ws.Set("head", "id", resultSetID); err != nil {
		return errors.Wrap(err, "cannot set as head")
	}
	log.Printf("saved result set head")
	return nil
}

func (s *ResultSetSnapshotStore) Head() (*serialisable.ViewSnapshot, error) {
	var headResultSetID int64
	if err := s.ws.Get("head", "id", &headResultSetID); err != nil && !errors.Is(err, storm.ErrNotFound) {
		return nil, errors.Wrap(err, "cannot get head")
	}

	var rss serialisable.ViewSnapshot
	if err := s.ws.One("ID", headResultSetID, &rss); err != nil {
		if errors.Is(err, storm.ErrNotFound) {
			return nil, nil
		} else {
			return nil, errors.Wrap(err, "cannot get head")
		}
	}

	return &rss, nil
}

func (s *ResultSetSnapshotStore) Remove(resultSetId int64) error {
	var rss serialisable.ViewSnapshot
	if err := s.ws.One("ID", resultSetId, &rss); err != nil {
		if errors.Is(err, storm.ErrNotFound) {
			return nil
		} else {
			return errors.Wrapf(err, "cannot get snapshot with ID %v", resultSetId)
		}
	}

	if err := s.ws.DeleteStruct(&rss); err != nil {
		return errors.Wrap(err, "cannot delete snapshot")
	}
	return nil
}

package viewsnapshot

import "github.com/lmika/dynamo-browse/internal/dynamo-browse/models/serialisable"

type ViewSnapshotStore interface {
	Save(rs *serialisable.ViewSnapshot) error
	SetAsHead(resultSetId int64) error
	CurrentlyViewedSnapshot() (*serialisable.ViewSnapshot, error)
	SetCurrentlyViewedSnapshot(resultSetId int64) error
	Find(resultSetID int64) (*serialisable.ViewSnapshot, error)
	Len() (int, error)
	Head() (*serialisable.ViewSnapshot, error)
	Remove(resultSetId int64) error
	Dehead(fromNode *serialisable.ViewSnapshot) error
}

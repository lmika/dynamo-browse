package workspaces

import "github.com/lmika/audax/internal/dynamo-browse/models/serialisable"

type ViewSnapshotStore interface {
	Save(rs *serialisable.ViewSnapshot) error
	SetAsHead(resultSetId int64) error
	CurrentlyViewedSnapshot() (*serialisable.ViewSnapshot, error)
	SetCurrentlyViewedSnapshot(resultSetId int64) error
	Find(resultSetID int64) (*serialisable.ViewSnapshot, error)
	Head() (*serialisable.ViewSnapshot, error)
	Remove(resultSetId int64) error
	Dehead(fromNode *serialisable.ViewSnapshot) error
}

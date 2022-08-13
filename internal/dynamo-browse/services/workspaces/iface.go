package workspaces

import "github.com/lmika/audax/internal/dynamo-browse/models/serialisable"

type ViewSnapshotStore interface {
	Save(rs *serialisable.ViewSnapshot) error
	SetAsHead(resultSetId int64) error
	Head() (*serialisable.ViewSnapshot, error)
	Remove(resultSetId int64) error
}

package workspaces

import "github.com/lmika/audax/internal/dynamo-browse/models/serialisable"

type ResultSetSnapshotStore interface {
	Save(rs *serialisable.ResultSetSnapshot) error
	SetAsHead(resultSetId int64) error
	Head() (*serialisable.ResultSetSnapshot, error)
}

package viewsnapshot_test

import (
	"github.com/lmika/audax/internal/dynamo-browse/models/queryexpr"
	"github.com/lmika/audax/internal/dynamo-browse/models/serialisable"
	"github.com/lmika/audax/internal/dynamo-browse/providers/workspacestore"
	"github.com/lmika/audax/internal/dynamo-browse/services/viewsnapshot"
	"github.com/lmika/audax/test/testworkspace"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestViewSnapshotService_PushSnapshot(t *testing.T) {
	t.Run("should not push duplicate snapshots", func(t *testing.T) {
		ws := testworkspace.New(t)

		service := viewsnapshot.NewService(workspacestore.NewResultSetSnapshotStore(ws))
		q, _ := queryexpr.Parse("pk = \"abc\"")
		qbs, _ := q.SerializeToBytes()

		// Push some snapshots
		err := service.PushSnapshot(serialisable.ViewSnapshotDetails{
			TableName: "normal-table",
			Query:     qbs,
			QueryHash: q.HashCode(),
			Filter:    "",
		})
		assert.NoError(t, err)

		cnt, err := service.Len()
		assert.NoError(t, err)
		assert.Equal(t, 1, cnt)

		q2, _ := queryexpr.Parse("another = \"test\"")
		qbs2, _ := q.SerializeToBytes()

		err = service.PushSnapshot(serialisable.ViewSnapshotDetails{
			TableName: "abnormal-table",
			Query:     qbs2,
			QueryHash: q2.HashCode(),
			Filter:    "fla",
		})
		assert.NoError(t, err)

		cnt, err = service.Len()
		assert.NoError(t, err)
		assert.Equal(t, 2, cnt)

		// Push a duplicate
		err = service.PushSnapshot(serialisable.ViewSnapshotDetails{
			TableName: "abnormal-table",
			Query:     qbs2,
			QueryHash: q2.HashCode(),
			Filter:    "fla",
		})
		assert.NoError(t, err)

		cnt, err = service.Len()
		assert.NoError(t, err)
		assert.Equal(t, 2, cnt)
	})
}

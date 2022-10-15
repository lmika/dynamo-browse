package viewsnapshot_test

import (
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

		// Push some snapshots
		err := service.PushSnapshot(serialisable.ViewSnapshotDetails{
			TableName: "normal-table",
			Query:     "pk = 'abc'",
			Filter:    "",
		})
		assert.NoError(t, err)

		cnt, err := service.Len()
		assert.NoError(t, err)
		assert.Equal(t, 1, cnt)

		err = service.PushSnapshot(serialisable.ViewSnapshotDetails{
			TableName: "abnormal-table",
			Query:     "pk = 'abc'",
			Filter:    "fla",
		})
		assert.NoError(t, err)

		cnt, err = service.Len()
		assert.NoError(t, err)
		assert.Equal(t, 2, cnt)

		// Push a duplicate
		err = service.PushSnapshot(serialisable.ViewSnapshotDetails{
			TableName: "abnormal-table",
			Query:     "pk = 'abc'",
			Filter:    "fla",
		})
		assert.NoError(t, err)

		cnt, err = service.Len()
		assert.NoError(t, err)
		assert.Equal(t, 2, cnt)
	})
}

package testworkspace

import (
	"github.com/lmika/dynamo-browse/internal/common/workspaces"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func New(t *testing.T) *workspaces.Workspace {
	wsTempFile := tempFile(t)

	wsManager := workspaces.New(workspaces.MetaInfo{Command: "dynamo-browse"})
	ws, err := wsManager.Open(wsTempFile)
	if err != nil {
		t.Fatalf("cannot create workspace manager: %v", err)
	}
	t.Cleanup(func() { ws.Close() })

	return ws
}

func tempFile(t *testing.T) string {
	t.Helper()

	tempFile, err := os.CreateTemp("", "export.csv")
	assert.NoError(t, err)
	tempFile.Close()

	t.Cleanup(func() {
		os.Remove(tempFile.Name())
	})

	return tempFile.Name()
}

package scriptmanager_test

import (
	"context"
	"github.com/lmika/audax/internal/dynamo-browse/services/scriptmanager"
	"github.com/lmika/audax/internal/dynamo-browse/services/scriptmanager/mocks"
	"github.com/stretchr/testify/assert"
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestService_RunAdHocScript(t *testing.T) {
	t.Run("successfully loads and executes a script", func(t *testing.T) {
		testFS := testScriptFile(t, "test.tm", `
			ui.print("Hello, world")
		`)

		mockedUIService := mocks.NewUIService(t)
		mockedUIService.EXPECT().PrintMessage("Hello, world")

		srv := scriptmanager.New(testFS)
		srv.SetIFaces(scriptmanager.Ifaces{
			UI: mockedUIService,
		})

		ctx := context.Background()
		err := srv.RunAdHocScript(ctx, "test.tm")
		assert.NoError(t, err)

		mockedUIService.AssertExpectations(t)
	})
}

func testScriptFile(t *testing.T, filename, code string) fs.FS {
	t.Helper()

	testFs := fstest.MapFS{
		filename: &fstest.MapFile{
			Data: []byte(code),
		},
	}
	return testFs
}

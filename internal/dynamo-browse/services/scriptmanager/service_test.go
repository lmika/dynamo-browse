package scriptmanager_test

import (
	"context"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/services/scriptmanager"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/services/scriptmanager/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io/fs"
	"testing"
	"testing/fstest"
	"time"
)

func TestService_RunAdHocScript(t *testing.T) {
	t.Run("successfully loads and executes a script", func(t *testing.T) {
		testFS := testScriptFile(t, "test.tm", `
			ui.print("Hello, world")
		`)

		mockedUIService := mocks.NewUIService(t)
		mockedUIService.EXPECT().PrintMessage(mock.Anything, "Hello, world")

		srv := scriptmanager.New(scriptmanager.WithFS(testFS))
		srv.SetIFaces(scriptmanager.Ifaces{
			UI: mockedUIService,
		})

		ctx := context.Background()
		err := <-srv.RunAdHocScript(ctx, "test.tm")
		assert.NoError(t, err)

		mockedUIService.AssertExpectations(t)
	})
}

func TestService_LoadScript(t *testing.T) {
	t.Run("successfully loads a script and exposes it as a plugin", func(t *testing.T) {
		testFS := testScriptFile(t, "test.tm", `
			ext.command("somewhere", func(a) {
				ui.print("Hello, " + a)
			})
		`)

		ctx := context.Background()

		mockedUIService := mocks.NewUIService(t)
		mockedUIService.EXPECT().PrintMessage(mock.Anything, "Hello, someone")

		srv := scriptmanager.New(scriptmanager.WithFS(testFS))
		srv.SetIFaces(scriptmanager.Ifaces{
			UI: mockedUIService,
		})

		plugin, err := srv.LoadScript(ctx, "test.tm")
		assert.NoError(t, err)
		assert.NotNil(t, plugin)
		assert.Equal(t, "test", plugin.Name())

		cmd := srv.LookupCommand("somewhere")
		assert.NotNil(t, cmd)

		errChan := make(chan error)
		err = cmd.Invoke(ctx, []string{"someone"}, errChan)
		assert.NoError(t, err)
		assert.NoError(t, waitForErr(t, errChan))

		mockedUIService.AssertExpectations(t)
	})

	t.Run("reloading a script with the same name should remove the old one", func(t *testing.T) {
		testFS := fstest.MapFS{
			"test.tm": &fstest.MapFile{
				Data: []byte(`
					ext.command("somewhere", func(a) {
						ui.print("Hello, " + a)
					})
				`),
			},
		}

		ctx := context.Background()

		mockedUIService := mocks.NewUIService(t)
		mockedUIService.EXPECT().PrintMessage(mock.Anything, "Hello, someone").Once()
		mockedUIService.EXPECT().PrintMessage(mock.Anything, "Goodbye, someone").Once()

		srv := scriptmanager.New(scriptmanager.WithFS(testFS))
		srv.SetIFaces(scriptmanager.Ifaces{
			UI: mockedUIService,
		})

		// Execute the old script
		_, err := srv.LoadScript(ctx, "test.tm")
		assert.NoError(t, err)

		cmd := srv.LookupCommand("somewhere")
		assert.NotNil(t, cmd)

		errChan := make(chan error)
		err = cmd.Invoke(ctx, []string{"someone"}, errChan)
		assert.NoError(t, err)
		assert.NoError(t, waitForErr(t, errChan))

		// Change the script and reload
		testFS["test.tm"] = &fstest.MapFile{
			Data: []byte(`
				ext.command("somewhere", func(a) {
					ui.print("Goodbye, " + a)
				})
			`),
		}

		_, err = srv.LoadScript(ctx, "test.tm")
		assert.NoError(t, err)

		cmd = srv.LookupCommand("somewhere")
		assert.NotNil(t, cmd)

		errChan = make(chan error)
		err = cmd.Invoke(ctx, []string{"someone"}, errChan)
		assert.NoError(t, err)
		assert.NoError(t, waitForErr(t, errChan))

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

func waitForErr(t *testing.T, errChan chan error) error {
	t.Helper()

	select {
	case err := <-errChan:
		return err
	case <-time.After(5 * time.Second):
		t.Fatalf("timed-out waiting for an error")
	}
	return nil
}

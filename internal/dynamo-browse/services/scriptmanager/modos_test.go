package scriptmanager_test

import (
	"context"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/services/scriptmanager"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/services/scriptmanager/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestOSModule_Env(t *testing.T) {
	t.Run("should return value of environment variables", func(t *testing.T) {
		t.Setenv("FULL_VALUE", "this is a value")
		t.Setenv("EMPTY_VALUE", "")

		testFS := testScriptFile(t, "test.tm", `
			assert(os.env("FULL_VALUE") == "this is a value")
			assert(os.env("EMPTY_VALUE") == "")
			assert(os.env("MISSING_VALUE") == nil)

			assert(bool(os.env("FULL_VALUE")) == true)
			assert(bool(os.env("EMPTY_VALUE")) == false)
			assert(bool(os.env("MISSING_VALUE")) == false)
		`)

		srv := scriptmanager.New(scriptmanager.WithFS(testFS))
		srv.SetDefaultOptions(scriptmanager.Options{
			OSExecShell: "/bin/bash",
			Permissions: scriptmanager.Permissions{
				AllowEnv: true,
			},
		})

		ctx := context.Background()
		err := <-srv.RunAdHocScript(ctx, "test.tm")
		assert.NoError(t, err)
	})

	t.Run("should return nil when no access to environment variables", func(t *testing.T) {
		t.Setenv("FULL_VALUE", "this is a value")
		t.Setenv("EMPTY_VALUE", "")

		testFS := testScriptFile(t, "test.tm", `
			assert(os.env("FULL_VALUE") == nil)
			assert(os.env("EMPTY_VALUE") == nil)
			assert(os.env("MISSING_VALUE") == nil)

			assert(bool(os.env("FULL_VALUE")) == false)
			assert(bool(os.env("EMPTY_VALUE")) == false)
			assert(bool(os.env("MISSING_VALUE")) == false)
		`)

		srv := scriptmanager.New(scriptmanager.WithFS(testFS))
		srv.SetDefaultOptions(scriptmanager.Options{
			OSExecShell: "/bin/bash",
			Permissions: scriptmanager.Permissions{
				AllowEnv: false,
			},
		})

		ctx := context.Background()
		err := <-srv.RunAdHocScript(ctx, "test.tm")
		assert.NoError(t, err)
	})
}

func TestOSModule_Exec(t *testing.T) {
	t.Run("should run command and return stdout", func(t *testing.T) {
		mockedUIService := mocks.NewUIService(t)
		mockedUIService.EXPECT().PrintMessage(mock.Anything, "false")
		mockedUIService.EXPECT().PrintMessage(mock.Anything, "hello world\n")

		testFS := testScriptFile(t, "test.tm", `
			res := os.exec('echo "hello world"')
			ui.print(res.is_err())
			ui.print(res.unwrap())
		`)

		srv := scriptmanager.New(scriptmanager.WithFS(testFS))
		srv.SetDefaultOptions(scriptmanager.Options{
			OSExecShell: "/bin/bash",
			Permissions: scriptmanager.Permissions{
				AllowShellCommands: true,
			},
		})
		srv.SetIFaces(scriptmanager.Ifaces{
			UI: mockedUIService,
		})

		ctx := context.Background()
		err := <-srv.RunAdHocScript(ctx, "test.tm")
		assert.NoError(t, err)

		mockedUIService.AssertExpectations(t)
	})

	t.Run("should refuse to execute command if do not have permissions", func(t *testing.T) {
		mockedUIService := mocks.NewUIService(t)
		mockedUIService.EXPECT().PrintMessage(mock.Anything, "true")

		testFS := testScriptFile(t, "test.tm", `
			res := os.exec('echo "hello world"')
			ui.print(res.is_err())
		`)

		srv := scriptmanager.New(scriptmanager.WithFS(testFS))
		srv.SetDefaultOptions(scriptmanager.Options{
			OSExecShell: "/bin/bash",
			Permissions: scriptmanager.Permissions{
				AllowShellCommands: false,
			},
		})
		srv.SetIFaces(scriptmanager.Ifaces{
			UI: mockedUIService,
		})

		ctx := context.Background()
		err := <-srv.RunAdHocScript(ctx, "test.tm")
		assert.NoError(t, err)

		mockedUIService.AssertExpectations(t)
	})

	t.Run("should be able to change permissions which will affect plugins", func(t *testing.T) {
		mockedUIService := mocks.NewUIService(t)
		mockedUIService.EXPECT().PrintMessage(mock.Anything, "Loaded the plugin\n")
		mockedUIService.EXPECT().PrintMessage(mock.Anything, "true")

		testFS := testScriptFile(t, "test.tm", `
			ext.command("mycommand", func() {
				ui.print(os.exec('echo "this cannot run"').is_err())
			})

			ui.print(os.exec('echo "Loaded the plugin"').unwrap())
		`)

		srv := scriptmanager.New(scriptmanager.WithFS(testFS))
		srv.SetDefaultOptions(scriptmanager.Options{
			OSExecShell: "/bin/bash",
			Permissions: scriptmanager.Permissions{
				AllowShellCommands: true,
			},
		})
		srv.SetIFaces(scriptmanager.Ifaces{
			UI: mockedUIService,
		})

		ctx := context.Background()
		_, err := srv.LoadScript(ctx, "test.tm")
		assert.NoError(t, err)

		srv.SetDefaultOptions(scriptmanager.Options{
			OSExecShell: "/bin/bash",
			Permissions: scriptmanager.Permissions{
				AllowShellCommands: false,
			},
		})

		errChan := make(chan error)
		assert.NoError(t, srv.LookupCommand("mycommand").Invoke(ctx, []string{}, errChan))
		assert.NoError(t, waitForErr(t, errChan))

		mockedUIService.AssertExpectations(t)
	})
}

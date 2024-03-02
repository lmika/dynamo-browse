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
			assert(os.getenv("FULL_VALUE") == "this is a value")
			assert(os.getenv("EMPTY_VALUE") == "")
			assert(os.getenv("MISSING_VALUE") == "")

			assert(bool(os.getenv("FULL_VALUE")) == true)
			assert(bool(os.getenv("EMPTY_VALUE")) == false)
			assert(bool(os.getenv("MISSING_VALUE")) == false)
		`)

		srv := scriptmanager.New(scriptmanager.WithFS(testFS))

		ctx := context.Background()
		err := <-srv.RunAdHocScript(ctx, "test.tm")
		assert.NoError(t, err)
	})
}

func TestOSModule_Exec(t *testing.T) {
	t.Run("should run command and return stdout", func(t *testing.T) {
		mockedUIService := mocks.NewUIService(t)
		mockedUIService.EXPECT().PrintMessage(mock.Anything, "hello world\n")

		testFS := testScriptFile(t, "test.tm", `
			res := exec('echo', ["hello world"]).stdout
			ui.print(res)
		`)

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

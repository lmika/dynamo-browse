package scriptmanager_test

import (
	"context"
	"github.com/lmika/audax/internal/dynamo-browse/services/scriptmanager"
	"github.com/lmika/audax/internal/dynamo-browse/services/scriptmanager/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestModUI_Prompt(t *testing.T) {
	t.Run("should successfully return prompt value", func(t *testing.T) {
		testFS := testScriptFile(t, "test.tm", `
			ui.print("Hello, world")
			var name = ui.prompt("What is your name? ")
			ui.print("Hello, " + name)
		`)

		promptChan := make(chan string)
		go func() {
			promptChan <- "T. Test"
		}()

		mockedUIService := mocks.NewUIService(t)
		mockedUIService.EXPECT().PrintMessage(mock.Anything, "Hello, world")
		mockedUIService.EXPECT().Prompt(mock.Anything, "What is your name? ").Return(promptChan)
		mockedUIService.EXPECT().PrintMessage(mock.Anything, "Hello, T. Test")

		srv := scriptmanager.New(testFS)
		srv.SetIFaces(scriptmanager.Ifaces{
			UI: mockedUIService,
		})

		ctx := context.Background()
		err := <-srv.RunAdHocScript(ctx, "test.tm")
		assert.NoError(t, err)

		mockedUIService.AssertExpectations(t)
	})

	t.Run("should return error if prompt was cancelled", func(t *testing.T) {
		testFS := testScriptFile(t, "test.tm", `
			ui.print("Hello, world")
			var name = ui.prompt("What is your name? ")
			ui.print("After")
		`)

		promptChan := make(chan string)
		close(promptChan)

		ctx := context.Background()

		mockedUIService := mocks.NewUIService(t)
		mockedUIService.EXPECT().PrintMessage(mock.Anything, "Hello, world")
		mockedUIService.EXPECT().Prompt(mock.Anything, "What is your name? ").Return(promptChan)

		srv := scriptmanager.New(testFS)
		srv.SetIFaces(scriptmanager.Ifaces{
			UI: mockedUIService,
		})

		err := <-srv.RunAdHocScript(ctx, "test.tm")
		assert.Error(t, err)

		mockedUIService.AssertNotCalled(t, "Prompt", "after")
		mockedUIService.AssertExpectations(t)
	})

	t.Run("should return error if context was cancelled", func(t *testing.T) {
		testFS := testScriptFile(t, "test.tm", `
			ui.print("Hello, world")
			var name = ui.prompt("What is your name? ")
			ui.print("After")
		`)

		promptChan := make(chan string)
		ctx, cancelFn := context.WithCancel(context.Background())
		defer cancelFn()

		mockedUIService := mocks.NewUIService(t)
		mockedUIService.EXPECT().PrintMessage(mock.Anything, "Hello, world")
		mockedUIService.EXPECT().Prompt(mock.Anything, "What is your name? ").Run(func(ctx context.Context, msg string) {
			cancelFn()
		}).Return(promptChan)

		srv := scriptmanager.New(testFS)
		srv.SetIFaces(scriptmanager.Ifaces{
			UI: mockedUIService,
		})

		err := <-srv.RunAdHocScript(ctx, "test.tm")
		assert.Error(t, err)

		mockedUIService.AssertNotCalled(t, "Prompt", "after")
		mockedUIService.AssertExpectations(t)
	})
}

package scriptmanager_test

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/services/scriptmanager"
	"github.com/lmika/audax/internal/dynamo-browse/services/scriptmanager/mocks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestModSession_Query(t *testing.T) {
	t.Run("should successfully return query result", func(t *testing.T) {
		rs := &models.ResultSet{}
		rs.SetItems([]models.Item{
			{"pk": &types.AttributeValueMemberS{Value: "abc"}},
			{"pk": &types.AttributeValueMemberS{Value: "1232"}},
		})

		mockedSessionService := mocks.NewSessionService(t)
		mockedSessionService.EXPECT().Query(mock.Anything, "some expr").Return(rs, nil)

		mockedUIService := mocks.NewUIService(t)
		mockedUIService.EXPECT().PrintMessage(mock.Anything, "2")
		mockedUIService.EXPECT().PrintMessage(mock.Anything, "res[0]['pk'].S = abc")
		mockedUIService.EXPECT().PrintMessage(mock.Anything, "res[1]['pk'].S = 1232")
		mockedUIService.EXPECT().PrintMessage(mock.Anything, "res[1].value('size(pk)') = 4")

		testFS := testScriptFile(t, "test.tm", `
			res := session.query("some expr").unwrap()
			// ui.print(len(res))
			ui.print(res.length)
			ui.print("res[0]['pk'].S = ", res[0].value("pk"))
			ui.print("res[1]['pk'].S = ", res[1].value("pk"))
			ui.print("res[1].attr('size(pk)') = ", res[1].attr("size(pk)"))
		`)

		srv := scriptmanager.New(testFS)
		srv.SetIFaces(scriptmanager.Ifaces{
			UI:      mockedUIService,
			Session: mockedSessionService,
		})

		ctx := context.Background()
		err := <-srv.RunAdHocScript(ctx, "test.tm")
		assert.NoError(t, err)

		mockedUIService.AssertExpectations(t)
		mockedSessionService.AssertExpectations(t)
	})

	t.Run("should return error if query returns error", func(t *testing.T) {
		mockedSessionService := mocks.NewSessionService(t)
		mockedSessionService.EXPECT().Query(mock.Anything, "some expr").Return(nil, errors.New("bang"))

		mockedUIService := mocks.NewUIService(t)
		mockedUIService.EXPECT().PrintMessage(mock.Anything, "true")
		mockedUIService.EXPECT().PrintMessage(mock.Anything, "err(\"bang\")")

		testFS := testScriptFile(t, "test.tm", `
			res := session.query("some expr")
			ui.print(res.is_err())
			ui.print(res)
		`)

		srv := scriptmanager.New(testFS)
		srv.SetIFaces(scriptmanager.Ifaces{
			UI:      mockedUIService,
			Session: mockedSessionService,
		})

		ctx := context.Background()
		err := <-srv.RunAdHocScript(ctx, "test.tm")
		assert.NoError(t, err)

		mockedUIService.AssertExpectations(t)
		mockedSessionService.AssertExpectations(t)
	})
}

func TestModSession_SetResultSet(t *testing.T) {
	t.Run("should set the result set on the session", func(t *testing.T) {
		rs := &models.ResultSet{}
		rs.SetItems([]models.Item{
			{"pk": &types.AttributeValueMemberS{Value: "abc"}},
			{"pk": &types.AttributeValueMemberS{Value: "1232"}},
		})

		mockedSessionService := mocks.NewSessionService(t)
		mockedSessionService.EXPECT().Query(mock.Anything, "some expr").Return(rs, nil)
		mockedSessionService.EXPECT().SetResultSet(mock.Anything, rs)

		mockedUIService := mocks.NewUIService(t)

		testFS := testScriptFile(t, "test.tm", `
			res := session.query("some expr").unwrap()
			session.set_result_set(res)
		`)

		srv := scriptmanager.New(testFS)
		srv.SetIFaces(scriptmanager.Ifaces{
			UI:      mockedUIService,
			Session: mockedSessionService,
		})

		ctx := context.Background()
		err := <-srv.RunAdHocScript(ctx, "test.tm")
		assert.NoError(t, err)

		mockedUIService.AssertExpectations(t)
		mockedSessionService.AssertExpectations(t)
	})
}

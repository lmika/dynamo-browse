package scriptmanager_test

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/services/scriptmanager"
	"github.com/lmika/audax/internal/dynamo-browse/services/scriptmanager/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestResultSetProxy(t *testing.T) {
	t.Run("should property return properties of a resultset and item", func(t *testing.T) {
		rs := &models.ResultSet{}
		rs.SetItems([]models.Item{
			{"pk": &types.AttributeValueMemberS{Value: "abc"}},
			{"pk": &types.AttributeValueMemberS{Value: "1232"}},
		})

		mockedSessionService := mocks.NewSessionService(t)
		mockedSessionService.EXPECT().Query(mock.Anything, "some expr", scriptmanager.QueryOptions{}).Return(rs, nil)

		mockedUIService := mocks.NewUIService(t)

		testFS := testScriptFile(t, "test.tm", `
			res := session.query("some expr").unwrap()

			// Test properties of the result set
			assert(res == res, "result_set.equals")
			assert(res.length == 2, "result_set.length")
			
			// Test properties of items
			assert(res[0].index == 0, "res[0].index")
			assert(res[0].result_set == res, "res[0].result_set")
			assert(res[0].attr('pk') == 'abc', "res[0].attr('pk')")
			
			assert(res[1].attr('pk') == '1232', "res[1].attr('pk')")
		`)

		srv := scriptmanager.New(scriptmanager.WithFS(testFS))
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

func TestResultSetProxy_SetAttr(t *testing.T) {
	t.Run("should set the value of the item within a result set", func(t *testing.T) {
		rs := &models.ResultSet{}
		rs.SetItems([]models.Item{
			{"pk": &types.AttributeValueMemberS{Value: "abc"}},
			{"pk": &types.AttributeValueMemberS{Value: "1232"}},
		})

		mockedSessionService := mocks.NewSessionService(t)
		mockedSessionService.EXPECT().Query(mock.Anything, "some expr", scriptmanager.QueryOptions{}).Return(rs, nil)
		mockedSessionService.EXPECT().SetResultSet(mock.Anything, mock.MatchedBy(func(rs *models.ResultSet) bool {
			assert.Equal(t, "bla-di-bla", rs.Items()[0]["pk"].(*types.AttributeValueMemberS).Value)
			assert.True(t, rs.IsDirty(0))
			return true
		}))

		mockedUIService := mocks.NewUIService(t)

		testFS := testScriptFile(t, "test.tm", `
			res := session.query("some expr").unwrap()
			res[0].set_attr("pk", "bla-di-bla")
			session.set_result_set(res)
		`)

		srv := scriptmanager.New(scriptmanager.WithFS(testFS))
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

func TestResultSetProxy_DeleteAttr(t *testing.T) {
	t.Run("should delete the value of the item within a result set", func(t *testing.T) {
		rs := &models.ResultSet{}
		rs.SetItems([]models.Item{
			{"pk": &types.AttributeValueMemberS{Value: "abc"}, "deleteMe": &types.AttributeValueMemberBOOL{Value: true}},
			{"pk": &types.AttributeValueMemberS{Value: "1232"}},
		})

		mockedSessionService := mocks.NewSessionService(t)
		mockedSessionService.EXPECT().Query(mock.Anything, "some expr", scriptmanager.QueryOptions{}).Return(rs, nil)
		mockedSessionService.EXPECT().SetResultSet(mock.Anything, mock.MatchedBy(func(rs *models.ResultSet) bool {
			assert.Equal(t, "abc", rs.Items()[0]["pk"].(*types.AttributeValueMemberS).Value)
			assert.Nil(t, rs.Items()[0]["deleteMe"])
			assert.True(t, rs.IsDirty(0))
			return true
		}))

		mockedUIService := mocks.NewUIService(t)

		testFS := testScriptFile(t, "test.tm", `
			res := session.query("some expr").unwrap()
			res[0].delete_attr("deleteMe")
			session.set_result_set(res)
		`)

		srv := scriptmanager.New(scriptmanager.WithFS(testFS))
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

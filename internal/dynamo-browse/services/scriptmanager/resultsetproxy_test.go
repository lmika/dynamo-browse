package scriptmanager_test

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/services/scriptmanager"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/services/scriptmanager/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestResultSetProxy(t *testing.T) {
	t.Run("should property return properties of a resultset and item", func(t *testing.T) {
		rs := &models.ResultSet{
			TableInfo: &models.TableInfo{
				Name: "test-table",
			},
		}
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
			assert(res.table.name, "hello")

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

func TestResultSetProxy_Find(t *testing.T) {
	t.Run("should return the first item that matches the given expression", func(t *testing.T) {
		rs := &models.ResultSet{}
		rs.SetItems([]models.Item{
			{"pk": &types.AttributeValueMemberS{Value: "abc"}},
			{"pk": &types.AttributeValueMemberS{Value: "abc"}, "sk": &types.AttributeValueMemberS{Value: "abc"}, "primary": &types.AttributeValueMemberS{Value: "yes"}},
			{"pk": &types.AttributeValueMemberS{Value: "1232"}, "findMe": &types.AttributeValueMemberS{Value: "yes"}},
			{"pk": &types.AttributeValueMemberS{Value: "2345"}, "findMe": &types.AttributeValueMemberS{Value: "second"}},
		})

		mockedSessionService := mocks.NewSessionService(t)
		mockedSessionService.EXPECT().Query(mock.Anything, "some expr", scriptmanager.QueryOptions{}).Return(rs, nil)

		testFS := testScriptFile(t, "test.tm", `
			res := session.query("some expr").unwrap()

			assert(res.find('findMe is "any"').attr("pk") == "1232")
			assert(res.find('findMe = "second"').attr("pk") == "2345")
			// assert(res.find('pk = sk').attr("primary") == "yes")

			assert(res.find('findMe = "missing"') == nil)
		`)

		srv := scriptmanager.New(scriptmanager.WithFS(testFS))
		srv.SetIFaces(scriptmanager.Ifaces{
			Session: mockedSessionService,
		})

		ctx := context.Background()
		err := <-srv.RunAdHocScript(ctx, "test.tm")
		assert.NoError(t, err)

		mockedSessionService.AssertExpectations(t)
	})
}

func TestResultSetProxy_GetAttr(t *testing.T) {
	t.Run("should return the value of items within a result set", func(t *testing.T) {
		rs := &models.ResultSet{}
		rs.SetItems([]models.Item{
			{
				"pk":   &types.AttributeValueMemberS{Value: "abc"},
				"sk":   &types.AttributeValueMemberN{Value: "123"},
				"bool": &types.AttributeValueMemberBOOL{Value: true},
				"null": &types.AttributeValueMemberNULL{Value: true},
				"list": &types.AttributeValueMemberL{Value: []types.AttributeValue{
					&types.AttributeValueMemberS{Value: "apple"},
					&types.AttributeValueMemberS{Value: "banana"},
					&types.AttributeValueMemberS{Value: "cherry"},
				}},
				"map": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
					"this":    &types.AttributeValueMemberS{Value: "that"},
					"another": &types.AttributeValueMemberS{Value: "thing"},
				}},
				"strSet": &types.AttributeValueMemberSS{Value: []string{"apple", "banana", "cherry"}},
				"numSet": &types.AttributeValueMemberNS{Value: []string{"123", "45.67", "8.911", "-321"}},
			},
		})

		mockedSessionService := mocks.NewSessionService(t)
		mockedSessionService.EXPECT().Query(mock.Anything, "some expr", scriptmanager.QueryOptions{}).Return(rs, nil)

		testFS := testScriptFile(t, "test.tm", `
			res := session.query("some expr").unwrap()

			assert(res[0].attr("pk") == "abc", "str attr")
			assert(res[0].attr("sk") == 123, "num attr")
			assert(res[0].attr("bool") == true, "bool attr")
			assert(res[0].attr("null") == nil, "null attr")
			assert(res[0].attr("list") == ["apple","banana","cherry"], "list attr")
			assert(res[0].attr("map") == {"this":"that", "another":"thing"}, "map attr")
			assert(res[0].attr("strSet") == {"apple","banana","cherry"}, "string set")
			assert(res[0].attr("numSet") == {123, 45.67, 8.911, -321}, "number set")
		`)

		srv := scriptmanager.New(scriptmanager.WithFS(testFS))
		srv.SetIFaces(scriptmanager.Ifaces{
			Session: mockedSessionService,
		})

		ctx := context.Background()
		err := <-srv.RunAdHocScript(ctx, "test.tm")
		assert.NoError(t, err)

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
			assert.Equal(t, "123", rs.Items()[0]["num"].(*types.AttributeValueMemberN).Value)
			assert.Equal(t, "123.45", rs.Items()[0]["numFloat"].(*types.AttributeValueMemberN).Value)
			assert.Equal(t, true, rs.Items()[0]["bool"].(*types.AttributeValueMemberBOOL).Value)
			assert.Equal(t, true, rs.Items()[0]["nil"].(*types.AttributeValueMemberNULL).Value)

			list := rs.Items()[0]["lists"].(*types.AttributeValueMemberL).Value
			assert.Equal(t, "abc", list[0].(*types.AttributeValueMemberS).Value)
			assert.Equal(t, "123", list[1].(*types.AttributeValueMemberN).Value)
			assert.Equal(t, true, list[2].(*types.AttributeValueMemberBOOL).Value)

			nestedLists := rs.Items()[0]["nestedLists"].(*types.AttributeValueMemberL).Value
			assert.Equal(t, "1", nestedLists[0].(*types.AttributeValueMemberL).Value[0].(*types.AttributeValueMemberN).Value)
			assert.Equal(t, "2", nestedLists[0].(*types.AttributeValueMemberL).Value[1].(*types.AttributeValueMemberN).Value)
			assert.Equal(t, "3", nestedLists[1].(*types.AttributeValueMemberL).Value[0].(*types.AttributeValueMemberN).Value)
			assert.Equal(t, "4", nestedLists[1].(*types.AttributeValueMemberL).Value[1].(*types.AttributeValueMemberN).Value)

			mapValue := rs.Items()[0]["map"].(*types.AttributeValueMemberM).Value
			assert.Equal(t, "world", mapValue["hello"].(*types.AttributeValueMemberS).Value)
			assert.Equal(t, "213", mapValue["nums"].(*types.AttributeValueMemberN).Value)

			numSet := rs.Items()[0]["numSet"].(*types.AttributeValueMemberNS).Value
			assert.Len(t, numSet, 4)
			assert.Contains(t, numSet, "1")
			assert.Contains(t, numSet, "2")
			assert.Contains(t, numSet, "3")
			assert.Contains(t, numSet, "4.5")

			strSet := rs.Items()[0]["strSet"].(*types.AttributeValueMemberSS).Value
			assert.Len(t, strSet, 3)
			assert.Contains(t, strSet, "a")
			assert.Contains(t, strSet, "b")
			assert.Contains(t, strSet, "c")

			assert.True(t, rs.IsDirty(0))
			return true
		}))

		mockedUIService := mocks.NewUIService(t)

		testFS := testScriptFile(t, "test.tm", `
			res := session.query("some expr").unwrap()

			res[0].set_attr("pk", "bla-di-bla")
			res[0].set_attr("num", 123)
			res[0].set_attr("numFloat", 123.45)
			res[0].set_attr("bool", true)
			res[0].set_attr("nil", nil)
			res[0].set_attr("lists", ['abc', 123, true])
			res[0].set_attr("nestedLists", [[1,2], [3,4]])
			res[0].set_attr("map", {"hello": "world", "nums": 213})
			res[0].set_attr("numSet", {1,2,3,4.5})
			res[0].set_attr("strSet", {"a","b","c"})

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

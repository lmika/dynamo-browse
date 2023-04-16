package scriptmanager_test

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/services/scriptmanager"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/services/scriptmanager/mocks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestModSession_Table(t *testing.T) {
	t.Run("should return details of the current table", func(t *testing.T) {
		tableDef := models.TableInfo{
			Name: "test_table",
			Keys: models.KeyAttribute{
				PartitionKey: "pk",
				SortKey:      "sk",
			},
			GSIs: []models.TableGSI{
				{
					Name: "index-1",
					Keys: models.KeyAttribute{
						PartitionKey: "ipk",
						SortKey:      "isk",
					},
				},
			},
		}
		rs := models.ResultSet{TableInfo: &tableDef}

		mockedSessionService := mocks.NewSessionService(t)
		mockedSessionService.EXPECT().ResultSet(mock.Anything).Return(&rs)

		testFS := testScriptFile(t, "test.tm", `
			table := session.current_table()

			assert(table.name == "test_table")
			assert(table.keys["hash"] == "pk")
			assert(table.keys["range"] == "sk")
			assert(len(table.gsis) == 1)
			assert(table.gsis[0].name == "index-1")
			assert(table.gsis[0].keys["hash"] == "ipk")
			assert(table.gsis[0].keys["range"] == "isk")

			assert(table == session.result_set().table)
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

	t.Run("should return nil if no current result set", func(t *testing.T) {
		mockedSessionService := mocks.NewSessionService(t)
		mockedSessionService.EXPECT().ResultSet(mock.Anything).Return(nil)

		testFS := testScriptFile(t, "test.tm", `
			table := session.current_table()

			assert(table == nil)
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

func TestModSession_Query(t *testing.T) {
	t.Run("should successfully return query result", func(t *testing.T) {
		rs := &models.ResultSet{}
		rs.SetItems([]models.Item{
			{"pk": &types.AttributeValueMemberS{Value: "abc"}},
			{"pk": &types.AttributeValueMemberS{Value: "1232"}},
		})

		mockedSessionService := mocks.NewSessionService(t)
		mockedSessionService.EXPECT().Query(mock.Anything, "some expr", scriptmanager.QueryOptions{}).Return(rs, nil)

		mockedUIService := mocks.NewUIService(t)
		mockedUIService.EXPECT().PrintMessage(mock.Anything, "2")
		mockedUIService.EXPECT().PrintMessage(mock.Anything, "res[0]['pk'].S = abc")
		mockedUIService.EXPECT().PrintMessage(mock.Anything, "res[1]['pk'].S = 1232")
		mockedUIService.EXPECT().PrintMessage(mock.Anything, "res[1].attr('size(pk)') = 4")

		testFS := testScriptFile(t, "test.tm", `
			res := session.query("some expr").unwrap()
			ui.print(res.length)
			ui.print("res[0]['pk'].S = ", res[0].attr("pk"))
			ui.print("res[1]['pk'].S = ", res[1].attr("pk"))
			ui.print("res[1].attr('size(pk)') = ", res[1].attr("size(pk)"))
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

	t.Run("should return error if query returns error", func(t *testing.T) {
		mockedSessionService := mocks.NewSessionService(t)
		mockedSessionService.EXPECT().Query(mock.Anything, "some expr", scriptmanager.QueryOptions{}).Return(nil, errors.New("bang"))

		mockedUIService := mocks.NewUIService(t)
		mockedUIService.EXPECT().PrintMessage(mock.Anything, "true")
		mockedUIService.EXPECT().PrintMessage(mock.Anything, "err(\"bang\")")

		testFS := testScriptFile(t, "test.tm", `
			res := session.query("some expr")
			ui.print(res.is_err())
			ui.print(res)
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

	t.Run("should successfully specify table name", func(t *testing.T) {
		rs := &models.ResultSet{}

		mockedSessionService := mocks.NewSessionService(t)
		mockedSessionService.EXPECT().Query(mock.Anything, "some expr", scriptmanager.QueryOptions{
			TableName: "some-table",
		}).Return(rs, nil)

		mockedUIService := mocks.NewUIService(t)

		testFS := testScriptFile(t, "test.tm", `
			res := session.query("some expr", {
				table: "some-table",
			})
			assert(!res.is_err())
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

	t.Run("should successfully specify table proxy", func(t *testing.T) {
		rs := &models.ResultSet{}

		mockedSessionService := mocks.NewSessionService(t)
		mockedSessionService.EXPECT().ResultSet(mock.Anything).Return(&models.ResultSet{
			TableInfo: &models.TableInfo{
				Name: "some-resultset-table",
			},
		})
		mockedSessionService.EXPECT().Query(mock.Anything, "some expr", scriptmanager.QueryOptions{
			TableName: "some-resultset-table",
		}).Return(rs, nil)

		mockedUIService := mocks.NewUIService(t)

		testFS := testScriptFile(t, "test.tm", `
			res := session.query("some expr", {
				table: session.result_set().table,
			})
			assert(!res.is_err())
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

	t.Run("should set placeholder values", func(t *testing.T) {
		rs := &models.ResultSet{}

		mockedSessionService := mocks.NewSessionService(t)
		mockedSessionService.EXPECT().Query(mock.Anything, ":name = $value", scriptmanager.QueryOptions{
			NamePlaceholders: map[string]string{
				"name":  "hello",
				"value": "world",
			},
			ValuePlaceholders: map[string]types.AttributeValue{
				"name":  &types.AttributeValueMemberS{Value: "hello"},
				"value": &types.AttributeValueMemberS{Value: "world"},
			},
		}).Return(rs, nil)

		mockedUIService := mocks.NewUIService(t)

		testFS := testScriptFile(t, "test.tm", `
			res := session.query(":name = $value", {
				args: {
					name: "hello",
					value: "world",
				},
			})
			assert(!res.is_err())
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

	t.Run("should support various placeholder value type", func(t *testing.T) {
		rs := &models.ResultSet{}

		mockedSessionService := mocks.NewSessionService(t)
		mockedSessionService.EXPECT().Query(mock.Anything, ":name = $value", scriptmanager.QueryOptions{
			NamePlaceholders: map[string]string{
				"str": "hello",
			},
			ValuePlaceholders: map[string]types.AttributeValue{
				"str":   &types.AttributeValueMemberS{Value: "hello"},
				"int":   &types.AttributeValueMemberN{Value: "123"},
				"float": &types.AttributeValueMemberN{Value: "3.14"},
				"bool":  &types.AttributeValueMemberBOOL{Value: true},
				"nil":   &types.AttributeValueMemberNULL{Value: true},
			},
		}).Return(rs, nil)

		mockedUIService := mocks.NewUIService(t)

		testFS := testScriptFile(t, "test.tm", `
			res := session.query(":name = $value", {
				args: {
					"str": "hello",
					"int": 123,
					"float": 3.14,
					"bool": true,
					"nil": nil,
				},
			})
			assert(!res.is_err())
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

	t.Run("should return error when placeholder value type is unsupported", func(t *testing.T) {
		mockedSessionService := mocks.NewSessionService(t)
		mockedUIService := mocks.NewUIService(t)

		testFS := testScriptFile(t, "test.tm", `
			res := session.query(":name = $value", {
				args: {
					"bad": func() { },
				},
			})
			assert(res.is_err())
		`)

		srv := scriptmanager.New(scriptmanager.WithFS(testFS))
		srv.SetIFaces(scriptmanager.Ifaces{
			UI:      mockedUIService,
			Session: mockedSessionService,
		})

		ctx := context.Background()
		err := <-srv.RunAdHocScript(ctx, "test.tm")
		assert.Error(t, err)

		mockedUIService.AssertExpectations(t)
		mockedSessionService.AssertExpectations(t)
	})
}

func TestModSession_SelectedItem(t *testing.T) {
	t.Run("should return selected item from service implementation", func(t *testing.T) {
		rs := &models.ResultSet{}
		rs.SetItems([]models.Item{
			{"pk": &types.AttributeValueMemberS{Value: "abc"}},
			{"pk": &types.AttributeValueMemberS{Value: "1232"}},
		})

		mockedSessionService := mocks.NewSessionService(t)
		mockedSessionService.EXPECT().ResultSet(mock.Anything).Return(rs)
		mockedSessionService.EXPECT().SelectedItemIndex(mock.Anything).Return(1)

		testFS := testScriptFile(t, "test.tm", `
			selItem := session.selected_item()

			assert(selItem != nil, "selItem != nil")
			assert(selItem.index == 1, "selItem.index")
			assert(selItem.result_set == session.result_set(), "selItem.result_set")
			assert(selItem.attr('pk') == '1232', "selItem.attr('pk')")
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

	t.Run("should return nil if selected item returns -1", func(t *testing.T) {
		rs := &models.ResultSet{}
		rs.SetItems([]models.Item{
			{"pk": &types.AttributeValueMemberS{Value: "abc"}},
			{"pk": &types.AttributeValueMemberS{Value: "1232"}},
		})

		mockedSessionService := mocks.NewSessionService(t)
		mockedSessionService.EXPECT().ResultSet(mock.Anything).Return(rs)
		mockedSessionService.EXPECT().SelectedItemIndex(mock.Anything).Return(-1)

		testFS := testScriptFile(t, "test.tm", `
			selItem := session.selected_item()

			assert(selItem == nil, "selItem != nil")
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

func TestModSession_SetResultSet(t *testing.T) {
	t.Run("should set the result set on the session", func(t *testing.T) {
		rs := &models.ResultSet{}
		rs.SetItems([]models.Item{
			{"pk": &types.AttributeValueMemberS{Value: "abc"}},
			{"pk": &types.AttributeValueMemberS{Value: "1232"}},
		})

		mockedSessionService := mocks.NewSessionService(t)
		mockedSessionService.EXPECT().Query(mock.Anything, "some expr", scriptmanager.QueryOptions{}).Return(rs, nil)
		mockedSessionService.EXPECT().SetResultSet(mock.Anything, rs)

		mockedUIService := mocks.NewUIService(t)

		testFS := testScriptFile(t, "test.tm", `
			res := session.query("some expr").unwrap()
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

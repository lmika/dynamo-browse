package controllers

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/common/ui/events"
	"github.com/lmika/awstools/internal/dynamo-browse/models"
	"github.com/lmika/awstools/internal/dynamo-browse/services/tables"
	"github.com/pkg/errors"
)

type TableWriteController struct {
	state                *State
	tableService         *tables.Service
	tableReadControllers *TableReadController
}

func NewTableWriteController(state *State, tableService *tables.Service, tableReadControllers *TableReadController) *TableWriteController {
	return &TableWriteController{
		state:                state,
		tableService:         tableService,
		tableReadControllers: tableReadControllers,
	}
}

func (twc *TableWriteController) ToggleMark(idx int) tea.Cmd {
	return func() tea.Msg {
		twc.state.withResultSet(func(resultSet *models.ResultSet) {
			resultSet.SetMark(idx, !resultSet.Marked(idx))
		})

		return ResultSetUpdated{}
	}
}

func (twc *TableWriteController) NewItem() tea.Cmd {
	return func() tea.Msg {
		// Work out which keys we need to prompt for
		rs := twc.state.ResultSet()

		keyPrompts := &promptSequence{
			prompts: []string{rs.TableInfo.Keys.PartitionKey + ": "},
		}
		if rs.TableInfo.Keys.SortKey != "" {
			keyPrompts.prompts = append(keyPrompts.prompts, rs.TableInfo.Keys.SortKey+": ")
		}
		keyPrompts.onAllDone = func(values []string) tea.Msg {
			twc.state.withResultSet(func(set *models.ResultSet) {
				newItem := models.Item{}

				// TODO: deal with keys of different type
				newItem[rs.TableInfo.Keys.PartitionKey] = &types.AttributeValueMemberS{Value: values[0]}
				if len(values) == 2 {
					newItem[rs.TableInfo.Keys.SortKey] = &types.AttributeValueMemberS{Value: values[1]}
				}

				set.AddNewItem(newItem, models.ItemAttribute{
					New:   true,
					Dirty: true,
				})
			})
			return NewResultSet{twc.state.ResultSet(), "New item added"}
		}

		return keyPrompts.next()
	}
}

func (twc *TableWriteController) SetStringValue(idx int, key string) tea.Cmd {
	return func() tea.Msg {
		return events.PromptForInputMsg{
			Prompt: "string value: ",
			OnDone: func(value string) tea.Cmd {
				return func() tea.Msg {
					twc.state.withResultSet(func(set *models.ResultSet) {
						set.Items()[idx][key] = &types.AttributeValueMemberS{Value: value}
						set.SetDirty(idx, true)
					})
					return ResultSetUpdated{}
				}
			},
		}
	}
}

func (twc *TableWriteController) PutItem(idx int) tea.Cmd {
	return func() tea.Msg {
		resultSet := twc.state.ResultSet()
		if !resultSet.IsDirty(idx) {
			return events.Error(errors.New("item is not dirty"))
		}

		return events.PromptForInputMsg{
			Prompt: "put item? ",
			OnDone: func(value string) tea.Cmd {
				return func() tea.Msg {
					if value != "y" {
						return nil
					}

					if err := twc.tableService.PutItemAt(context.Background(), resultSet, idx); err != nil {
						return events.Error(err)
					}
					return ResultSetUpdated{}
				}
			},
		}
	}
}

func (twc *TableWriteController) TouchItem(idx int) tea.Cmd {
	return func() tea.Msg {
		resultSet := twc.state.ResultSet()
		if resultSet.IsDirty(idx) {
			return events.Error(errors.New("cannot touch dirty items"))
		}

		return events.PromptForInputMsg{
			Prompt: "touch item? ",
			OnDone: func(value string) tea.Cmd {
				return func() tea.Msg {
					if value != "y" {
						return nil
					}

					if err := twc.tableService.PutItemAt(context.Background(), resultSet, idx); err != nil {
						return events.Error(err)
					}
					return ResultSetUpdated{}
				}
			},
		}
	}
}

func (twc *TableWriteController) NoisyTouchItem(idx int) tea.Cmd {
	return func() tea.Msg {
		resultSet := twc.state.ResultSet()
		if resultSet.IsDirty(idx) {
			return events.Error(errors.New("cannot noisy touch dirty items"))
		}

		return events.PromptForInputMsg{
			Prompt: "noisy touch item? ",
			OnDone: func(value string) tea.Cmd {
				return func() tea.Msg {
					ctx := context.Background()

					if value != "y" {
						return nil
					}

					item := resultSet.Items()[0]
					if err := twc.tableService.Delete(ctx, resultSet.TableInfo, []models.Item{item}); err != nil {
						return events.Error(err)
					}

					if err := twc.tableService.Put(ctx, resultSet.TableInfo, item); err != nil {
						return events.Error(err)
					}

					return twc.tableReadControllers.doScan(ctx, resultSet)
				}
			},
		}
	}
}

func (twc *TableWriteController) DeleteMarked() tea.Cmd {
	return func() tea.Msg {
		resultSet := twc.state.ResultSet()
		markedItems := resultSet.MarkedItems()

		if len(markedItems) == 0 {
			return events.StatusMsg("no marked items")
		}

		return events.PromptForInputMsg{
			Prompt: fmt.Sprintf("delete %d items? ", len(markedItems)),
			OnDone: func(value string) tea.Cmd {
				if value != "y" {
					return events.SetStatus("operation aborted")
				}

				return func() tea.Msg {
					ctx := context.Background()
					if err := twc.tableService.Delete(ctx, resultSet.TableInfo, markedItems); err != nil {
						return events.Error(err)
					}

					return twc.tableReadControllers.doScan(ctx, resultSet)
				}
			},
		}
	}
}

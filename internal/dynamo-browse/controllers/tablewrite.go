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
	"strconv"
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
			return twc.state.buildNewResultSetMessage("New item added")
		}

		return keyPrompts.next()
	}
}

func (twc *TableWriteController) SetAttributeValue(idx int, itemType models.ItemType, key string) tea.Cmd {
	apPath := newAttrPath(key)

	var attrValue types.AttributeValue
	if err := twc.state.withResultSetReturningError(func(set *models.ResultSet) (err error) {
		attrValue, err = apPath.follow(set.Items()[idx])
		return err
	}); err != nil {
		return events.SetError(err)
	}

	switch itemType {
	case models.UnsetItemType:
		switch attrValue.(type) {
		case *types.AttributeValueMemberS:
			return twc.setStringValue(idx, apPath)
		case *types.AttributeValueMemberN:
			return twc.setNumberValue(idx, apPath)
		case *types.AttributeValueMemberBOOL:
			return twc.setBoolValue(idx, apPath)
		default:
			return events.SetError(errors.New("attribute type for key must be set"))
		}
	case models.StringItemType:
		return twc.setStringValue(idx, apPath)
	case models.NumberItemType:
		return twc.setNumberValue(idx, apPath)
	case models.BoolItemType:
		return twc.setBoolValue(idx, apPath)
	case models.NullItemType:
		return twc.setNullValue(idx, apPath)
	default:
		return events.SetError(errors.New("unsupported attribute type"))
	}
}

func (twc *TableWriteController) setStringValue(idx int, attr attrPath) tea.Cmd {
	return func() tea.Msg {
		return events.PromptForInputMsg{
			Prompt: "string value: ",
			OnDone: func(value string) tea.Cmd {
				return func() tea.Msg {
					if err := twc.state.withResultSetReturningError(func(set *models.ResultSet) error {
						err := attr.setAt(set.Items()[idx], &types.AttributeValueMemberS{Value: value})
						if err != nil {
							return err
						}

						set.SetDirty(idx, true)
						set.RefreshColumns()
						return nil
					}); err != nil {
						return events.Error(err)
					}
					return ResultSetUpdated{}
				}
			},
		}
	}
}

func (twc *TableWriteController) setNumberValue(idx int, attr attrPath) tea.Cmd {
	return func() tea.Msg {
		return events.PromptForInputMsg{
			Prompt: "number value: ",
			OnDone: func(value string) tea.Cmd {
				return func() tea.Msg {
					if err := twc.state.withResultSetReturningError(func(set *models.ResultSet) error {
						err := attr.setAt(set.Items()[idx], &types.AttributeValueMemberN{Value: value})
						if err != nil {
							return err
						}

						set.SetDirty(idx, true)
						set.RefreshColumns()
						return nil
					}); err != nil {
						return events.Error(err)
					}
					return ResultSetUpdated{}
				}
			},
		}
	}
}

func (twc *TableWriteController) setBoolValue(idx int, attr attrPath) tea.Cmd {
	return func() tea.Msg {
		return events.PromptForInputMsg{
			Prompt: "bool value: ",
			OnDone: func(value string) tea.Cmd {
				return func() tea.Msg {
					b, err := strconv.ParseBool(value)
					if err != nil {
						return events.Error(err)
					}

					if err := twc.state.withResultSetReturningError(func(set *models.ResultSet) error {
						err := attr.setAt(set.Items()[idx], &types.AttributeValueMemberBOOL{Value: b})
						if err != nil {
							return err
						}

						set.SetDirty(idx, true)
						set.RefreshColumns()
						return nil
					}); err != nil {
						return events.Error(err)
					}
					return ResultSetUpdated{}
				}
			},
		}
	}
}

func (twc *TableWriteController) setNullValue(idx int, attr attrPath) tea.Cmd {
	return func() tea.Msg {
		if err := twc.state.withResultSetReturningError(func(set *models.ResultSet) error {
			err := attr.setAt(set.Items()[idx], &types.AttributeValueMemberNULL{Value: true})
			if err != nil {
				return err
			}

			set.SetDirty(idx, true)
			set.RefreshColumns()
			return nil
		}); err != nil {
			return events.Error(err)
		}
		return ResultSetUpdated{}
	}
}

func (twc *TableWriteController) DeleteAttribute(idx int, key string) tea.Cmd {
	return func() tea.Msg {
		// Verify that the expression is valid
		apPath := newAttrPath(key)

		if err := twc.state.withResultSetReturningError(func(set *models.ResultSet) error {
			_, err := apPath.follow(set.Items()[idx])
			return err
		}); err != nil {
			return events.Error(err)
		}

		if err := twc.state.withResultSetReturningError(func(set *models.ResultSet) error {
			err := apPath.deleteAt(set.Items()[idx])
			if err != nil {
				return err
			}

			set.SetDirty(idx, true)
			set.RefreshColumns()
			return nil
		}); err != nil {
			return events.Error(err)
		}

		return ResultSetUpdated{}
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

func (twc *TableWriteController) PutItems() tea.Cmd {
	return func() tea.Msg {
		var (
			expectedPuts int
			markedItems  int
		)

		twc.state.withResultSet(func(rs *models.ResultSet) {
			markedItems = len(rs.MarkedItems())
			for i := range rs.Items() {
				if rs.IsDirty(i) && (markedItems == 0 || rs.Marked(i)) {
					expectedPuts++
				}
			}
		})

		if expectedPuts == 0 {
			if markedItems > 0 {
				return events.StatusMsg("no marked items are modified")
			} else {
				return events.StatusMsg("no items are modified")
			}
		}

		var promptMessage string
		if markedItems > 0 {
			promptMessage = applyToN("put ", expectedPuts, "marked item", "marked items", "? ")
		} else {
			promptMessage = applyToN("put ", expectedPuts, "item", "items", "? ")
		}

		return events.PromptForInputMsg{
			Prompt: promptMessage,
			OnDone: func(value string) tea.Cmd {
				if value != "y" {
					return events.SetStatus("operation aborted")
				}

				return func() tea.Msg {
					if err := twc.state.withResultSetReturningError(func(rs *models.ResultSet) error {
						updated, err := twc.tableService.PutSelectedItems(context.Background(), rs, func(idx int) bool {
							return rs.IsDirty(idx) && (markedItems == 0 || rs.Marked(idx))
						})
						if err != nil {
							return err
						} else if updated != expectedPuts {
							return errors.Errorf("expected %d updates but only %d were applied", expectedPuts, updated)
						}
						return nil
					}); err != nil {
						return events.Error(err)
					}

					return ResultSetUpdated{
						statusMessage: applyToN("", expectedPuts, "item", "item", " put to table"),
					}
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

					return twc.tableReadControllers.doScan(ctx, resultSet, resultSet.Query)
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
			Prompt: applyToN("delete ", len(markedItems), "item", "items", "? "),
			OnDone: func(value string) tea.Cmd {
				if value != "y" {
					return events.SetStatus("operation aborted")
				}

				return func() tea.Msg {
					ctx := context.Background()
					if err := twc.tableService.Delete(ctx, resultSet.TableInfo, markedItems); err != nil {
						return events.Error(err)
					}

					return twc.tableReadControllers.doScan(ctx, resultSet, resultSet.Query)
				}
			},
		}
	}
}

func applyToN(prefix string, n int, singular, plural, suffix string) string {
	if n == 1 {
		return fmt.Sprintf("%v%v %v%v", prefix, n, singular, suffix)
	}
	return fmt.Sprintf("%v%v %v%v", prefix, n, plural, suffix)
}

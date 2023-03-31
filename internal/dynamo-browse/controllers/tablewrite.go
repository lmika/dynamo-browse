package controllers

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/audax/internal/common/sliceutils"
	"github.com/lmika/audax/internal/common/ui/events"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/models/queryexpr"
	"github.com/lmika/audax/internal/dynamo-browse/services/tables"
	"github.com/pkg/errors"
	"log"
	"strconv"
)

type TableWriteController struct {
	state                *State
	tableService         *tables.Service
	jobController        *JobsController
	tableReadControllers *TableReadController
	settingProvider      SettingsProvider
}

func NewTableWriteController(
	state *State,
	tableService *tables.Service,
	jobController *JobsController,
	tableReadControllers *TableReadController,
	settingProvider SettingsProvider,
) *TableWriteController {
	return &TableWriteController{
		state:                state,
		tableService:         tableService,
		jobController:        jobController,
		tableReadControllers: tableReadControllers,
		settingProvider:      settingProvider,
	}
}

func (twc *TableWriteController) ToggleMark(idx int) tea.Msg {
	twc.state.withResultSet(func(resultSet *models.ResultSet) {
		resultSet.SetMark(idx, !resultSet.Marked(idx))
	})

	return ResultSetUpdated{}
}

func (twc *TableWriteController) NewItem() tea.Msg {
	if err := twc.assertReadWrite(); err != nil {
		return events.Error(err)
	}

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

func (twc *TableWriteController) SetAttributeValue(idx int, itemType models.ItemType, key string) tea.Msg {
	path, err := queryexpr.Parse(key)
	if err != nil {
		return events.Error(err)
	}

	var attrValue types.AttributeValue
	if err := twc.state.withResultSetReturningError(func(set *models.ResultSet) (err error) {
		if !path.IsModifiablePath(set.Items()[idx]) {
			return errors.Errorf("path cannot be used to set attribute value")
		}

		attrValue, err = path.EvalItem(set.Items()[idx])
		return err
	}); err != nil {
		return events.Error(err)
	}

	log.Printf("sa attribute value = %v", attrValue)

	switch itemType {
	case models.UnsetItemType:
		switch attrValue.(type) {
		case *types.AttributeValueMemberS:
			return twc.setStringValue(idx, path)
		case *types.AttributeValueMemberN:
			return twc.setNumberValue(idx, path)
		case *types.AttributeValueMemberBOOL:
			return twc.setBoolValue(idx, path)
		default:
			return events.Error(errors.New("attribute type for key must be set"))
		}
	case models.StringItemType:
		return twc.setStringValue(idx, path)
	case models.NumberItemType:
		return twc.setNumberValue(idx, path)
	case models.BoolItemType:
		return twc.setBoolValue(idx, path)
	case models.NullItemType:
		return twc.setNullValue(idx, path)
	case models.ExprValueItemType:
		return twc.setToExpressionValue(idx, path)
	default:
		return events.Error(errors.New("unsupported attribute type"))
	}
}

func (twc *TableWriteController) setStringValue(idx int, attr *queryexpr.QueryExpr) tea.Msg {
	return events.PromptForInputMsg{
		Prompt: "string value: ",
		OnDone: func(value string) tea.Msg {
			if err := twc.state.withResultSetReturningError(func(set *models.ResultSet) error {
				if err := applyToMarkedItems(set, idx, func(idx int, item models.Item) error {
					if err := attr.SetEvalItem(item, &types.AttributeValueMemberS{Value: value}); err != nil {
						return err
					}
					set.SetDirty(idx, true)
					return nil
				}); err != nil {
					return err
				}
				set.RefreshColumns()
				return nil
			}); err != nil {
				return events.Error(err)
			}
			return ResultSetUpdated{}
		},
	}
}

func (twc *TableWriteController) setToExpressionValue(idx int, attr *queryexpr.QueryExpr) tea.Msg {
	return events.PromptForInputMsg{
		Prompt: "expr value: ",
		OnDone: func(value string) tea.Msg {
			valueExpr, err := queryexpr.Parse(value)
			if err != nil {
				return events.Error(err)
			}

			if err := twc.state.withResultSetReturningError(func(set *models.ResultSet) error {
				if err := applyToMarkedItems(set, idx, func(idx int, item models.Item) error {
					newValue, err := valueExpr.EvalItem(item)
					if err != nil {
						return err
					}
					if err := attr.SetEvalItem(item, newValue); err != nil {
						return err
					}
					set.SetDirty(idx, true)
					return nil
				}); err != nil {
					return err
				}
				set.RefreshColumns()
				return nil
			}); err != nil {
				return events.Error(err)
			}
			return ResultSetUpdated{}
		},
	}
}

func (twc *TableWriteController) setNumberValue(idx int, attr *queryexpr.QueryExpr) tea.Msg {
	return events.PromptForInputMsg{
		Prompt: "number value: ",
		OnDone: func(value string) tea.Msg {
			if err := twc.state.withResultSetReturningError(func(set *models.ResultSet) error {
				if err := applyToMarkedItems(set, idx, func(idx int, item models.Item) error {
					if err := attr.SetEvalItem(item, &types.AttributeValueMemberN{Value: value}); err != nil {
						return err
					}
					set.SetDirty(idx, true)
					return nil
				}); err != nil {
					return err
				}
				set.RefreshColumns()
				return nil
			}); err != nil {
				return events.Error(err)
			}
			return ResultSetUpdated{}
		},
	}
}

func (twc *TableWriteController) setBoolValue(idx int, attr *queryexpr.QueryExpr) tea.Msg {
	return events.PromptForInputMsg{
		Prompt: "bool value: ",
		OnDone: func(value string) tea.Msg {
			b, err := strconv.ParseBool(value)
			if err != nil {
				return events.Error(err)
			}

			if err := twc.state.withResultSetReturningError(func(set *models.ResultSet) error {
				if err := applyToMarkedItems(set, idx, func(idx int, item models.Item) error {
					if err := attr.SetEvalItem(item, &types.AttributeValueMemberBOOL{Value: b}); err != nil {
						return err
					}
					set.SetDirty(idx, true)
					return nil
				}); err != nil {
					return err
				}
				set.RefreshColumns()
				return nil
			}); err != nil {
				return events.Error(err)
			}
			return ResultSetUpdated{}
		},
	}
}

func (twc *TableWriteController) setNullValue(idx int, attr *queryexpr.QueryExpr) tea.Msg {
	if err := twc.state.withResultSetReturningError(func(set *models.ResultSet) error {
		if err := applyToMarkedItems(set, idx, func(idx int, item models.Item) error {
			if err := attr.SetEvalItem(item, &types.AttributeValueMemberNULL{Value: true}); err != nil {
				return err
			}
			set.SetDirty(idx, true)
			return nil
		}); err != nil {
			return err
		}
		set.RefreshColumns()
		return nil
	}); err != nil {
		return events.Error(err)
	}
	return ResultSetUpdated{}
}

func (twc *TableWriteController) DeleteAttribute(idx int, key string) tea.Msg {
	path, err := queryexpr.Parse(key)
	if err != nil {
		return events.Error(err)
	}

	if err := twc.state.withResultSetReturningError(func(set *models.ResultSet) error {
		if !path.IsModifiablePath(set.Items()[idx]) {
			return errors.Errorf("path cannot be used to set attribute value")
		}
		return nil
	}); err != nil {
		return events.Error(err)
	}

	if err := twc.state.withResultSetReturningError(func(set *models.ResultSet) error {
		err := path.DeleteAttribute(set.Items()[idx])
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

func (twc *TableWriteController) PutItems() tea.Msg {
	if err := twc.assertReadWrite(); err != nil {
		return events.Error(err)
	}

	var (
		markedItemCount int
	)
	var itemsToPut []models.ItemIndex

	twc.state.withResultSet(func(rs *models.ResultSet) {
		if markedItems := rs.MarkedItems(); len(markedItems) > 0 {
			for _, mi := range markedItems {
				markedItemCount += 1
				if rs.IsDirty(mi.Index) {
					itemsToPut = append(itemsToPut, mi)
				}
			}
		} else {
			for i, itm := range rs.Items() {
				if rs.IsDirty(i) {
					itemsToPut = append(itemsToPut, models.ItemIndex{Item: itm, Index: i})
				}
			}
		}
	})

	if len(itemsToPut) == 0 {
		if markedItemCount > 0 {
			return events.StatusMsg("no marked items are modified")
		} else {
			return events.StatusMsg("no items are modified")
		}
	}

	var promptMessage string
	if markedItemCount > 0 {
		promptMessage = applyToN("put ", len(itemsToPut), "marked item", "marked items", "? ")
	} else {
		promptMessage = applyToN("put ", len(itemsToPut), "item", "items", "? ")
	}

	return events.PromptForInputMsg{
		Prompt: promptMessage,
		OnDone: func(value string) tea.Msg {
			if value != "y" {
				return events.StatusMsg("operation aborted")
			}

			return NewJob(twc.jobController, "Updating items…", func(ctx context.Context) (*models.ResultSet, error) {
				rs := twc.state.ResultSet()
				err := twc.tableService.PutSelectedItems(ctx, rs, itemsToPut)
				if err != nil {
					return nil, err
				}
				return rs, nil
			}).OnDone(func(rs *models.ResultSet) tea.Msg {
				return ResultSetUpdated{
					statusMessage: applyToN("", len(itemsToPut), "item", "item", " put to table"),
				}
			}).Submit()
		},
	}
}

func (twc *TableWriteController) TouchItem(idx int) tea.Msg {
	if err := twc.assertReadWrite(); err != nil {
		return events.Error(err)
	}

	resultSet := twc.state.ResultSet()
	if resultSet.IsDirty(idx) {
		return events.Error(errors.New("cannot touch dirty items"))
	}

	return events.PromptForInputMsg{
		Prompt: "touch item? ",
		OnDone: func(value string) tea.Msg {
			if value != "y" {
				return nil
			}

			if err := twc.tableService.PutItemAt(context.Background(), resultSet, idx); err != nil {
				return events.Error(err)
			}
			return ResultSetUpdated{}
		},
	}
}

func (twc *TableWriteController) NoisyTouchItem(idx int) tea.Msg {
	if err := twc.assertReadWrite(); err != nil {
		return events.Error(err)
	}

	resultSet := twc.state.ResultSet()
	if resultSet.IsDirty(idx) {
		return events.Error(errors.New("cannot noisy touch dirty items"))
	}

	return events.PromptForInputMsg{
		Prompt: "noisy touch item? ",
		OnDone: func(value string) tea.Msg {
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

			return twc.tableReadControllers.doScan(resultSet, resultSet.Query, false, resultSetUpdateTouch)
		},
	}
}

func (twc *TableWriteController) DeleteMarked() tea.Msg {
	if err := twc.assertReadWrite(); err != nil {
		return events.Error(err)
	}

	resultSet := twc.state.ResultSet()
	markedItems := resultSet.MarkedItems()

	if len(markedItems) == 0 {
		return events.StatusMsg("no marked items")
	}

	return events.PromptForInputMsg{
		Prompt: applyToN("delete ", len(markedItems), "item", "items", "? "),
		OnDone: func(value string) tea.Msg {
			if value != "y" {
				return events.StatusMsg("operation aborted")
			}

			return NewJob(twc.jobController, "Deleting items…", func(ctx context.Context) (struct{}, error) {
				err := twc.tableService.Delete(ctx, resultSet.TableInfo, sliceutils.Map(markedItems, func(index models.ItemIndex) models.Item {
					return index.Item
				}))
				return struct{}{}, err
			}).OnDone(func(_ struct{}) tea.Msg {
				return twc.tableReadControllers.doScan(resultSet, resultSet.Query, false, resultSetUpdateTouch)
			}).Submit()
		},
	}
}

func (twc *TableWriteController) assertReadWrite() error {
	b, err := twc.settingProvider.IsReadOnly()
	if err != nil {
		return err
	} else if b {
		return models.ErrReadOnly
	}
	return nil
}

func applyToN(prefix string, n int, singular, plural, suffix string) string {
	if n == 1 {
		return fmt.Sprintf("%v%v %v%v", prefix, n, singular, suffix)
	}
	return fmt.Sprintf("%v%v %v%v", prefix, n, plural, suffix)
}

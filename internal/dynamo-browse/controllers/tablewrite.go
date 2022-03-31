package controllers

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/common/ui/events"
	"github.com/lmika/awstools/internal/dynamo-browse/services/tables"
)

type TableWriteController struct {
	tableService         *tables.Service
	tableReadControllers *TableReadController
}

func NewTableWriteController(tableService *tables.Service, tableReadControllers *TableReadController) *TableWriteController {
	return &TableWriteController{
		tableService:         tableService,
		tableReadControllers: tableReadControllers,
	}
}

func (twc *TableWriteController) ToggleMark(idx int) tea.Cmd {
	return func() tea.Msg {
		resultSet := twc.tableReadControllers.ResultSet()
		resultSet.SetMark(idx, !resultSet.Marked(idx))

		return ResultSetUpdated{}
	}
}

func (twc *TableWriteController) DeleteMarked() tea.Cmd {
	return func() tea.Msg {
		resultSet := twc.tableReadControllers.ResultSet()
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

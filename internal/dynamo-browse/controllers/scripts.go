package controllers

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/audax/internal/common/ui/commandctrl"
	"github.com/lmika/audax/internal/common/ui/events"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/models/queryexpr"
	"github.com/lmika/audax/internal/dynamo-browse/services/scriptmanager"
	"github.com/pkg/errors"
)

type ScriptController struct {
	scriptManager       *scriptmanager.Service
	tableReadController *TableReadController
	sendMsg             func(msg tea.Msg)
}

func NewScriptController(scriptManager *scriptmanager.Service, tableReadController *TableReadController) *ScriptController {
	sc := &ScriptController{
		scriptManager:       scriptManager,
		tableReadController: tableReadController,
	}
	scriptManager.SetIFaces(scriptmanager.Ifaces{
		UI:      &uiImpl{sc: sc},
		Session: &sessionImpl{sc: sc},
	})
	return sc
}

func (sc *ScriptController) SetMessageSender(sendMsg func(msg tea.Msg)) {
	sc.sendMsg = sendMsg
}

func (sc *ScriptController) LoadScript(filename string) tea.Msg {
	ctx := context.Background()
	plugin, err := sc.scriptManager.LoadScript(ctx, filename)
	if err != nil {
		return events.Error(err)
	}

	return events.StatusMsg(fmt.Sprintf("Script '%v' loaded", plugin.Name()))
}

func (sc *ScriptController) RunScript(filename string) tea.Msg {
	ctx := context.Background()
	if err := sc.scriptManager.StartAdHocScript(ctx, filename, sc.waitAndPrintScriptError()); err != nil {
		return events.Error(err)
	}
	return nil
}

func (sc *ScriptController) waitAndPrintScriptError() chan error {
	errChan := make(chan error)
	go func() {
		if err := <-errChan; err != nil {
			sc.sendMsg(events.Error(err))
		}
	}()
	return errChan
}

func (sc *ScriptController) LookupCommand(name string) commandctrl.Command {
	cmd := sc.scriptManager.LookupCommand(name)
	if cmd == nil {
		return nil
	}

	return func(execCtx commandctrl.ExecContext, args []string) tea.Msg {
		errChan := sc.waitAndPrintScriptError()
		ctx := context.Background()

		if err := cmd.Invoke(ctx, args, errChan); err != nil {
			return events.Error(err)
		}
		return nil
	}
}

type uiImpl struct {
	sc *ScriptController
}

func (u uiImpl) PrintMessage(ctx context.Context, msg string) {
	u.sc.sendMsg(events.StatusMsg(msg))
}

func (u uiImpl) Prompt(ctx context.Context, msg string) chan string {
	resultChan := make(chan string)
	u.sc.sendMsg(events.PromptForInputMsg{
		Prompt: msg,
		OnDone: func(value string) tea.Msg {
			resultChan <- value
			return nil
		},
		OnCancel: func() tea.Msg {
			close(resultChan)
			return nil
		},
	})
	return resultChan
}

type sessionImpl struct {
	sc *ScriptController
}

func (s sessionImpl) ResultSet(ctx context.Context) *models.ResultSet {
	return s.sc.tableReadController.state.ResultSet()
}

func (s sessionImpl) SetResultSet(ctx context.Context, newResultSet *models.ResultSet) {
	state := s.sc.tableReadController.state
	state.setResultSetAndFilter(newResultSet, state.Filter())
	s.sc.sendMsg(state.buildNewResultSetMessage(""))
}

func (s sessionImpl) Query(ctx context.Context, query string) (*models.ResultSet, error) {
	currentResultSet := s.sc.tableReadController.state.ResultSet()
	if currentResultSet == nil {
		// TODO: this should only be used if there's no current table
		return nil, errors.New("no table selected")
	}

	expr, err := queryexpr.Parse(query)
	if err != nil {
		return nil, err
	}

	newResultSet, err := s.sc.tableReadController.tableService.ScanOrQuery(context.Background(), currentResultSet.TableInfo, expr)
	if err != nil {
		return nil, err
	}
	return newResultSet, nil
}

package controllers

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/audax/internal/common/ui/events"
	"github.com/lmika/audax/internal/dynamo-browse/services/scriptmanager"
)

type ScriptController struct {
	scriptManager *scriptmanager.Service
	sendMsg       func(msg tea.Msg)
}

func NewScriptController(scriptManager *scriptmanager.Service) *ScriptController {
	sc := &ScriptController{
		scriptManager: scriptManager,
	}
	scriptManager.SetIFaces(scriptmanager.Ifaces{
		UI: &uiImpl{sc: sc},
	})
	return sc
}

func (sc *ScriptController) SetMessageSender(sendMsg func(msg tea.Msg)) {
	sc.sendMsg = sendMsg
}

func (sc *ScriptController) RunScript(filename string, doneChan chan error) tea.Msg {
	ctx := context.Background()
	sc.scriptManager.StartAdHocScript(ctx, filename, doneChan)
	return nil
}

func (sc *ScriptController) WaitAndPrintScriptError() chan error {
	errChan := make(chan error)
	go func() {
		if err := <-errChan; err != nil {
			sc.sendMsg(events.Error(err))
		}
	}()
	return errChan
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

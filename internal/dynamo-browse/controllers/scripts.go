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

func (sc *ScriptController) RunScript(filename string) tea.Msg {
	// TEMP
	ctx := context.Background()
	if err := sc.scriptManager.RunAdHocScript(ctx, filename); err != nil {
		return events.Error(err)
	}
	// END TEMP
	return nil
}

type uiImpl struct {
	sc *ScriptController
}

func (u uiImpl) PrintMessage(msg string) {
	u.sc.sendMsg(events.StatusMsg(msg))
}

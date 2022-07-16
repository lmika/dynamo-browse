package controllers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/common/ui/events"
)

type promptSequence struct {
	prompts        []string
	receivedValues []string
	onAllDone      func(values []string) tea.Msg
}

func (ps *promptSequence) next() tea.Msg {
	if len(ps.receivedValues) < len(ps.prompts) {
		return events.PromptForInputMsg{
			Prompt: ps.prompts[len(ps.receivedValues)],
			OnDone: func(value string) tea.Cmd {
				ps.receivedValues = append(ps.receivedValues, value)
				return ps.next
			},
		}
	}
	return ps.onAllDone(ps.receivedValues)
}

//type SetAttributeArg struct {
//	attrType models.ItemType
//	attrName string
//}
//
//func ParseSetAttributeArgs(args []string) (attrArgs []SetAttributeArg, err error) {
//	var currArg SetAttributeArg
//	for _, arg := range args {
//		if arg[0] == '-' {
//			currArg.attrType = models.ItemType(arg[1:])
//		} else {
//			currArg.attrName = arg
//			attrArgs = append(attrArgs, currArg)
//			currArg = SetAttributeArg{}
//		}
//	}
//	return attrArgs, nil
//}

package controllers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/dynamo-browse/internal/common/ui/events"
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
			OnDone: func(value string) tea.Msg {
				ps.receivedValues = append(ps.receivedValues, value)
				return ps.next()
			},
		}
	}
	return ps.onAllDone(ps.receivedValues)
}

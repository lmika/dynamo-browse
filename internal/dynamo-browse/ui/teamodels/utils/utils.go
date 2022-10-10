package utils

import tea "github.com/charmbracelet/bubbletea"

type CmdCollector struct {
	cmds []tea.Cmd
}

func (c *CmdCollector) Add(cmd tea.Cmd) {
	if cmd != nil {
		c.cmds = append(c.cmds, cmd)
	}
}

func (c *CmdCollector) Collect(m any, cmd tea.Cmd) any {
	if cmd != nil {
		c.cmds = append(c.cmds, cmd)
	}
	return m
}

func (c CmdCollector) Cmd() tea.Cmd {
	switch len(c.cmds) {
	case 0:
		return nil
	case 1:
		return c.cmds[0]
	default:
		return tea.Batch(c.cmds...)
	}
}

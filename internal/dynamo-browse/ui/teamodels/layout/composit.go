package layout

import (
	"bufio"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type Compositor struct {
	background ResizingModel

	foreground   ResizingModel
	foreX, foreY int
	foreW, foreH int
}

func NewCompositor(background ResizingModel) *Compositor {
	return &Compositor{
		background: background,
	}
}

func (c *Compositor) Init() tea.Cmd {
	return c.background.Init()
}

func (c *Compositor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// TODO: allow the compositor the
	newM, cmd := c.background.Update(msg)
	c.background = newM.(ResizingModel)
	return c, cmd
}

func (c *Compositor) SetOverlay(m ResizingModel, x, y, w, h int) {
	c.foreground = m
	c.foreX, c.foreY = x, y
	c.foreW, c.foreH = w, h
}

func (c *Compositor) View() string {
	if c.foreground == nil {
		return c.background.View()
	}

	// Need to compose
	backgroundView := c.background.View()
	foregroundViewLines := strings.Split(c.foreground.View(), "\n")

	backgroundScanner := bufio.NewScanner(strings.NewReader(backgroundView))
	compositeOutput := new(strings.Builder)

	r := 0
	for backgroundScanner.Scan() {
		if r > 0 {
			compositeOutput.WriteRune('\n')
		}

		line := backgroundScanner.Text()
		if r >= c.foreY && r < c.foreY+c.foreH {
			compositeOutput.WriteString(line[:c.foreX])

			foregroundScanPos := r - c.foreY
			if foregroundScanPos < len(foregroundViewLines) {
				displayLine := foregroundViewLines[foregroundScanPos]
				compositeOutput.WriteString(lipgloss.PlaceHorizontal(c.foreW, lipgloss.Left, displayLine, lipgloss.WithWhitespaceChars(" ")))
			}

			compositeOutput.WriteString(line[c.foreX+c.foreW:])
		} else {
			compositeOutput.WriteString(line)
		}
		r++
	}

	return compositeOutput.String()
}

func (c *Compositor) Resize(w, h int) ResizingModel {
	c.background = c.background.Resize(w, h)
	if c.foreground != nil {
		c.foreground = c.foreground.Resize(c.foreW, c.foreH)
	}
	return c
}

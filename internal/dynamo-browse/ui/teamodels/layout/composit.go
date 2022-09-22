package layout

import (
	"bufio"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/ansi"
	"github.com/muesli/reflow/truncate"
	"strings"
)

type Compositor struct {
	background tea.Model

	foreground   tea.Model
	foreX, foreY int
	foreW, foreH int
}

func NewCompositor(background tea.Model) *Compositor {
	return &Compositor{
		background: background,
	}
}

func (c *Compositor) Init() tea.Cmd {
	return c.background.Init()
}

func (c *Compositor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if c.foreground != nil {
		c.foreground, cmd = c.foreground.Update(msg)
	} else {
		c.background, cmd = c.background.Update(msg)
	}
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
	_ = foregroundViewLines

	backgroundScanner := bufio.NewScanner(strings.NewReader(backgroundView))
	compositeOutput := new(strings.Builder)

	r := 0
	for backgroundScanner.Scan() {
		if r > 0 {
			compositeOutput.WriteRune('\n')
		}

		line := backgroundScanner.Text()
		if r >= c.foreY && r < c.foreY+c.foreH {
			compositeOutput.WriteString(truncate.String(line, uint(c.foreX)))

			foregroundScanPos := r - c.foreY
			if foregroundScanPos < len(foregroundViewLines) {
				displayLine := foregroundViewLines[foregroundScanPos]
				compositeOutput.WriteString(lipgloss.PlaceHorizontal(c.foreW, lipgloss.Left, displayLine, lipgloss.WithWhitespaceChars(" ")))
			}

			rightStr := c.renderBackgroundUpTo(line, c.foreX+c.foreW)

			// Need to find a way to cut the string here
			compositeOutput.WriteString(rightStr)
		} else {
			compositeOutput.WriteString(line)
		}
		r++
	}

	return compositeOutput.String()
}

func (c *Compositor) Resize(w, h int) ResizingModel {
	c.background = Resize(c.background, w, h)
	if c.foreground != nil {
		c.foreground = Resize(c.foreground, c.foreW, c.foreH)
	}
	return c
}

func (c *Compositor) renderBackgroundUpTo(line string, x int) string {
	ansiSequences := new(strings.Builder)
	posX := 0
	inAnsi := false

	for i, c := range line {
		if c == ansi.Marker {
			ansiSequences.WriteRune(c)
			inAnsi = true
		} else if inAnsi {
			ansiSequences.WriteRune(c)
			if ansi.IsTerminator(c) {
				inAnsi = false
			}
		} else {
			if posX >= x {
				return ansiSequences.String() + line[i:]
			}
			posX += runewidth.RuneWidth(c)
		}
	}
	return ""
}

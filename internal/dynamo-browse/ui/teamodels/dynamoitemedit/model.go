package dynamoitemedit

import (
	table "github.com/calyptia/go-bubble-table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/awstools/internal/dynamo-browse/ui/teamodels/utils"
)

type Model struct {
	submodel  tea.Model
	table     table.Model
	textInput textinput.Model

	visible  bool
	editMode *editMode
	w, h     int
}

type editMode struct {
	index     int
	textInput textinput.Model
}

func NewModel(submodel tea.Model) *Model {
	tbl := table.New([]string{"name", "type", "value"}, 0, 0)
	//rows := make([]table.Row, 0)

	model := &Model{
		submodel: submodel,
	}

	rows := []table.Row{
		itemModel{model: model, name: "pk", attrType: "S", attrValue: "3baa4a4b-f977-4966-bed4-ba9243dbc942"},
		itemModel{model: model, name: "sk", attrType: "S", attrValue: "3baa4a4b-f977-4966-bed4-ba9243dbc942"},
		itemModel{model: model, name: "name", attrType: "S", attrValue: "My name"},
	}
	tbl.SetRows(rows)

	model.table = tbl

	return model
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.editMode != nil {
			switch msg.String() {
			case "enter":
				m.editMode = nil
			case "ctrl+c", "esc":
				m.editMode = nil
			default:
				m.editMode.textInput, cmd = utils.Update(m.editMode.textInput, msg)
			}
			return m, nil
		} else if m.visible {
			switch msg.String() {
			case "i", "up":
				m.table.GoUp()
			case "k", "down":
				m.table.GoDown()
			case "enter":
				m.enterEditMode()
			case "ctrl+c", "esc":
				m.visible = false
			}
			return m, nil
		}
	}

	m.submodel, cmd = utils.Update(m.submodel, msg)
	return m, cmd
}

func (m *Model) enterEditMode() {
	m.editMode = &editMode{
		textInput: textinput.New(),
		index:     m.table.Cursor(),
	}
	m.editMode.textInput.Focus()
}

func (m *Model) View() string {
	if !m.visible {
		return m.submodel.View()
	}

	return m.table.View()
}

func (m *Model) Resize(w, h int) layout.ResizingModel {
	m.w, m.h = w, h
	m.table.SetSize(w, h)
	m.submodel = layout.Resize(m.submodel, w, h)
	return m
}

func (m *Model) Visible() {
	m.visible = true
}

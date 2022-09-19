package ui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/audax/internal/common/ui/commandctrl"
	"github.com/lmika/audax/internal/common/ui/events"
	"github.com/lmika/audax/internal/dynamo-browse/controllers"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/services/itemrenderer"
	"github.com/lmika/audax/internal/dynamo-browse/ui/keybindings"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/colselector"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/dialogprompt"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/dynamoitemedit"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/dynamoitemview"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/dynamotableview"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/statusandprompt"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/styles"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/tableselect"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/utils"
	"github.com/pkg/errors"
	"log"
	"os"
	"strings"
)

const (
	ViewModeTablePrimary   = 0
	ViewModeTableItemEqual = 1
	ViewModeItemPrimary    = 2
	ViewModeItemOnly       = 3
	ViewModeTableOnly      = 4

	ViewModeCount = 5

	initRCFilename = "$HOME/.config/audax/dynamo-browse/init.rc"
)

type Model struct {
	tableReadController  *controllers.TableReadController
	tableWriteController *controllers.TableWriteController
	commandController    *commandctrl.CommandController
	itemEdit             *dynamoitemedit.Model
	statusAndPrompt      *statusandprompt.StatusAndPrompt
	tableSelect          *tableselect.Model

	mainViewIndex int

	root      tea.Model
	tableView *dynamotableview.Model
	itemView  *dynamoitemview.Model
	mainView  tea.Model
	keyMap    *keybindings.ViewKeyBindings
}

func NewModel(
	rc *controllers.TableReadController,
	wc *controllers.TableWriteController,
	itemRendererService *itemrenderer.Service,
	cc *commandctrl.CommandController,
	keyBindingController *controllers.KeyBindingController,
	defaultKeyMap *keybindings.KeyBindings,
) Model {
	uiStyles := styles.DefaultStyles

	dtv := dynamotableview.New(defaultKeyMap.TableView, uiStyles)
	div := dynamoitemview.New(itemRendererService, uiStyles)
	mainView := layout.NewVBox(layout.LastChildFixedAt(14), dtv, div)

	colSelector := colselector.New(mainView)
	itemEdit := dynamoitemedit.NewModel(colSelector)
	statusAndPrompt := statusandprompt.New(itemEdit, "", uiStyles.StatusAndPrompt)
	dialogPrompt := dialogprompt.New(statusAndPrompt)
	tableSelect := tableselect.New(dialogPrompt, uiStyles)

	cc.AddCommands(&commandctrl.CommandList{
		Commands: map[string]commandctrl.Command{
			"quit": commandctrl.NoArgCommand(tea.Quit),
			"table": func(ctx commandctrl.ExecContext, args []string) tea.Msg {
				if len(args) == 0 {
					return rc.ListTables()
				} else {
					return rc.ScanTable(args[0])
				}
			},
			"export": func(ctx commandctrl.ExecContext, args []string) tea.Msg {
				if len(args) == 0 {
					return events.Error(errors.New("expected filename"))
				}
				return rc.ExportCSV(args[0])
			},
			"unmark": commandctrl.NoArgCommand(rc.Unmark),
			"delete": commandctrl.NoArgCommand(wc.DeleteMarked),

			// TEMP
			"new-item": commandctrl.NoArgCommand(wc.NewItem),
			"set-attr": func(ctx commandctrl.ExecContext, args []string) tea.Msg {
				if len(args) == 0 {
					return events.Error(errors.New("expected field"))
				}

				var itemType = models.UnsetItemType
				if len(args) == 2 {
					switch strings.ToUpper(args[0]) {
					case "-S":
						itemType = models.StringItemType
					case "-N":
						itemType = models.NumberItemType
					case "-BOOL":
						itemType = models.BoolItemType
					case "-NULL":
						itemType = models.NullItemType
					default:
						return events.Error(errors.New("unrecognised item type"))
					}
					args = args[1:]
				}

				return wc.SetAttributeValue(dtv.SelectedItemIndex(), itemType, args[0])
			},
			"del-attr": func(ctx commandctrl.ExecContext, args []string) tea.Msg {
				if len(args) == 0 {
					return events.Error(errors.New("expected field"))
				}
				return wc.DeleteAttribute(dtv.SelectedItemIndex(), args[0])
			},

			"put": func(ctx commandctrl.ExecContext, args []string) tea.Msg {
				return wc.PutItems()
			},
			"touch": func(ctx commandctrl.ExecContext, args []string) tea.Msg {
				return wc.TouchItem(dtv.SelectedItemIndex())
			},
			"noisy-touch": func(ctx commandctrl.ExecContext, args []string) tea.Msg {
				return wc.NoisyTouchItem(dtv.SelectedItemIndex())
			},

			"echo": func(ctx commandctrl.ExecContext, args []string) tea.Msg {
				s := new(strings.Builder)
				for _, arg := range args {
					s.WriteString(arg)
				}
				return events.StatusMsg(s.String())
			},
			"rebind": func(ctx commandctrl.ExecContext, args []string) tea.Msg {
				if len(args) != 2 {
					return events.Error(errors.New("expected: bindingName newKey"))
				}
				return keyBindingController.Rebind(args[0], args[1], ctx.FromFile)
			},

			// Aliases
			"sa": cc.Alias("set-attr"),
			"da": cc.Alias("del-attr"),
			"w":  cc.Alias("put"),
			"q":  cc.Alias("quit"),
		},
	})

	root := layout.FullScreen(tableSelect)

	return Model{
		tableReadController:  rc,
		tableWriteController: wc,
		commandController:    cc,
		itemEdit:             itemEdit,
		statusAndPrompt:      statusAndPrompt,
		tableSelect:          tableSelect,
		root:                 root,
		tableView:            dtv,
		itemView:             div,
		mainView:             mainView,
		keyMap:               defaultKeyMap.View,
	}
}

func (m Model) Init() tea.Cmd {
	// TODO: this should probably be moved somewhere else
	rcFilename := os.ExpandEnv(initRCFilename)
	if err := m.commandController.ExecuteFile(rcFilename); err != nil {
		log.Println(err)
	}

	return m.tableReadController.Init
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case controllers.SetTableItemView:
		cmd := m.setMainViewIndex(msg.ViewIndex)
		return m, cmd
	case controllers.ResultSetUpdated:
		return m, tea.Batch(
			m.tableView.Refresh(),
			events.SetStatus(msg.StatusMessage()),
		)
	case tea.KeyMsg:
		if !m.statusAndPrompt.InPrompt() && !m.tableSelect.Visible() {
			switch {
			case key.Matches(msg, m.keyMap.Mark):
				if idx := m.tableView.SelectedItemIndex(); idx >= 0 {
					return m, func() tea.Msg { return m.tableWriteController.ToggleMark(idx) }
				}
			case key.Matches(msg, m.keyMap.CopyItemToClipboard):
				if idx := m.tableView.SelectedItemIndex(); idx >= 0 {
					return m, func() tea.Msg { return m.tableReadController.CopyItemToClipboard(idx) }
				}
			case key.Matches(msg, m.keyMap.Rescan):
				return m, m.tableReadController.Rescan
			case key.Matches(msg, m.keyMap.PromptForQuery):
				return m, m.tableReadController.PromptForQuery
			case key.Matches(msg, m.keyMap.PromptForFilter):
				return m, m.tableReadController.Filter
			case key.Matches(msg, m.keyMap.ViewBack):
				return m, m.tableReadController.ViewBack
			case key.Matches(msg, m.keyMap.ViewForward):
				return m, m.tableReadController.ViewForward
			case key.Matches(msg, m.keyMap.CycleLayoutForward):
				return m, func() tea.Msg {
					return controllers.SetTableItemView{ViewIndex: utils.Cycle(m.mainViewIndex, 1, ViewModeCount)}
				}
			case key.Matches(msg, m.keyMap.CycleLayoutBackwards):
				return m, func() tea.Msg {
					return controllers.SetTableItemView{ViewIndex: utils.Cycle(m.mainViewIndex, -1, ViewModeCount)}
				}
			//case "e":
			//	m.itemEdit.Visible()
			//	return m, nil
			case key.Matches(msg, m.keyMap.PromptForCommand):
				return m, m.commandController.Prompt
			case key.Matches(msg, m.keyMap.PromptForTable):
				return m, func() tea.Msg {
					return m.tableReadController.ListTables()
				}
			case key.Matches(msg, m.keyMap.Quit):
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.root, cmd = m.root.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return m.root.View()
}

func (m *Model) setMainViewIndex(viewIndex int) tea.Cmd {
	log.Printf("setting view index = %v", viewIndex)

	var newMainView tea.Model
	switch viewIndex {
	case ViewModeTablePrimary:
		newMainView = layout.NewVBox(layout.LastChildFixedAt(14), m.tableView, m.itemView)
	case ViewModeTableItemEqual:
		newMainView = layout.NewVBox(layout.EqualSize(), m.tableView, m.itemView)
	case ViewModeItemPrimary:
		newMainView = layout.NewVBox(layout.FirstChildFixedAt(7), m.tableView, m.itemView)
	case ViewModeItemOnly:
		newMainView = layout.NewZStack(m.itemView, m.tableView)
	case ViewModeTableOnly:
		newMainView = layout.NewZStack(m.tableView, m.tableView)
	default:
		newMainView = m.mainView
	}

	m.mainViewIndex = viewIndex
	m.mainView = newMainView
	m.itemEdit.SetSubmodel(m.mainView)
	return m.tableView.Refresh()
}

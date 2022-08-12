package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/audax/internal/common/ui/commandctrl"
	"github.com/lmika/audax/internal/common/ui/events"
	"github.com/lmika/audax/internal/dynamo-browse/controllers"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/services/pluginruntime"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/dialogprompt"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/dynamoitemedit"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/dynamoitemview"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/dynamotableview"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/layout"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/statusandprompt"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/styles"
	"github.com/lmika/audax/internal/dynamo-browse/ui/teamodels/tableselect"
	"github.com/pkg/errors"
	"log"
	"os"
	"strings"
)

type Model struct {
	tableReadController  *controllers.TableReadController
	tableWriteController *controllers.TableWriteController
	commandController    *commandctrl.CommandController

	dynamoTableView *dynamotableview.Model
	itemEdit        *dynamoitemedit.Model
	statusAndPrompt *statusandprompt.StatusAndPrompt
	tableSelect     *tableselect.Model

	root      tea.Model
	tableView *dynamotableview.Model
	itemView  *dynamoitemview.Model
}

func NewModel(
	rc *controllers.TableReadController,
	wc *controllers.TableWriteController,
	cc *commandctrl.CommandController,
	prs *pluginruntime.Service,
) Model {
	uiStyles := styles.DefaultStyles

	dtv := dynamotableview.New(uiStyles)
	div := dynamoitemview.New(uiStyles)
	mainView := layout.NewVBox(layout.LastChildFixedAt(17), dtv, div)

	itemEdit := dynamoitemedit.NewModel(mainView)
	statusAndPrompt := statusandprompt.New(itemEdit, "", uiStyles.StatusAndPrompt)
	dialogPrompt := dialogprompt.New(statusAndPrompt)
	tableSelect := tableselect.New(dialogPrompt, uiStyles)

	cc.AddCommands(&commandctrl.CommandContext{
		Commands: map[string]commandctrl.Command{
			"quit": commandctrl.NoArgCommand(tea.Quit),
			"table": func(args []string) tea.Cmd {
				if len(args) == 0 {
					return rc.ListTables()
				} else {
					return rc.ScanTable(args[0])
				}
			},
			"export": func(args []string) tea.Cmd {
				if len(args) == 0 {
					return events.SetError(errors.New("expected filename"))
				}
				return rc.ExportCSV(args[0])
			},
			"unmark": commandctrl.NoArgCommand(rc.Unmark()),
			"delete": commandctrl.NoArgCommand(wc.DeleteMarked()),

			"new-item": commandctrl.NoArgCommand(wc.NewItem()),
			"set-attr": func(args []string) tea.Cmd {
				if len(args) == 0 {
					return events.SetError(errors.New("expected field"))
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
						return events.SetError(errors.New("unrecognised item type"))
					}
					args = args[1:]
				}

				return wc.SetAttributeValue(dtv.SelectedItemIndex(), itemType, args[0])
			},
			"del-attr": func(args []string) tea.Cmd {
				if len(args) == 0 {
					return events.SetError(errors.New("expected field"))
				}
				return wc.DeleteAttribute(dtv.SelectedItemIndex(), args[0])
			},

			"put": func(args []string) tea.Cmd {
				return wc.PutItems()
			},
			"touch": func(args []string) tea.Cmd {
				return wc.TouchItem(dtv.SelectedItemIndex())
			},
			"noisy-touch": func(args []string) tea.Cmd {
				return wc.NoisyTouchItem(dtv.SelectedItemIndex())
			},

			// TEMP
			"loadscript": func(args []string) tea.Cmd {
				if len(args) == 0 {
					return events.SetError(errors.New("expected filename"))
				}

				_, err := prs.Load(args[0])
				if err != nil {
					log.Println(err)
				}

				return nil
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
		dynamoTableView:      dtv,
		itemEdit:             itemEdit,
		statusAndPrompt:      statusAndPrompt,
		tableSelect:          tableSelect,
		root:                 root,
		tableView:            dtv,
		itemView:             div,
	}
}

func (m Model) Init() tea.Cmd {
	// TEMP
	rcFilename := os.ExpandEnv("$HOME/.config/audax/dynamo-browse/init.rc")
	if rcFile, err := os.ReadFile(rcFilename); err == nil {
		if err := m.commandController.ExecuteFile(rcFile, rcFilename); err != nil {
			log.Printf("error executing %v: %v", err)
		}
	}
	// END TEMP

	return m.tableReadController.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case controllers.ResultSetUpdated:
		return m, m.tableView.Refresh()
	case tea.KeyMsg:
		if !m.statusAndPrompt.InPrompt() && !m.tableSelect.Visible() {
			switch msg.String() {
			case "m":
				if idx := m.tableView.SelectedItemIndex(); idx >= 0 {
					return m, m.tableWriteController.ToggleMark(idx)
				}
			case "R":
				return m, m.tableReadController.Rescan()
			case "?":
				return m, m.tableReadController.PromptForQuery()
			case "/":
				return m, m.tableReadController.Filter()
			//case "e":
			//	m.itemEdit.Visible()
			//	return m, nil
			case ":":
				return m, m.commandController.Prompt()
			case "ctrl+c", "esc":
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

func (m Model) SelectedItemIndex() int {
	return m.dynamoTableView.SelectedItemIndex()
}

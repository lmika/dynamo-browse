package controllers

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/dynamo-browse/internal/common/ui/events"
	bus "github.com/lmika/events"
	"github.com/pkg/errors"
	"log"
	"strconv"
)

const (
	BusEventSettingsUpdated = "settings.updated"
)

type SettingsController struct {
	settings SettingsProvider
	bus      *bus.Bus
}

func NewSettingsController(sp SettingsProvider, bus *bus.Bus) *SettingsController {
	return &SettingsController{
		settings: sp,
		bus:      bus,
	}
}

func (sc *SettingsController) SetSetting(name string, value string) tea.Msg {
	switch name {
	case "ro":
		if err := sc.settings.SetReadOnly(true); err != nil {
			return events.Error(err)
		}
		return events.WrappedStatusMsg{
			Message: "In read-only mode",
			Next:    SettingsUpdated{},
		}
	case "rw":
		if err := sc.settings.SetReadOnly(false); err != nil {
			return events.Error(err)
		}
		return events.WrappedStatusMsg{
			Message: "In read-write mode",
			Next:    SettingsUpdated{},
		}
	case "read-only":
		if value == "" {
			isRO, _ := sc.settings.IsReadOnly()
			return events.StatusMsg(fmt.Sprintf("read-only = %v", isRO))
		}

		newRO, err := strconv.ParseBool(value)
		if err != nil {
			return events.Error(errors.Wrapf(err, "bad value: %v", value))
		}

		if err := sc.settings.SetReadOnly(newRO); err != nil {
			return events.Error(err)
		}
	case "default-limit":
		if value == "" {
			return events.StatusMsg(fmt.Sprintf("default-limit = %v", sc.settings.DefaultLimit()))
		}

		newLimit, err := strconv.Atoi(value)
		if err != nil {
			return events.Error(errors.Wrapf(err, "bad value: %v", value))
		}

		if err := sc.settings.SetDefaultLimit(newLimit); err != nil {
			return events.Error(err)
		}
		return events.WrappedStatusMsg{
			Message: events.StatusMsg(fmt.Sprintf("Default query limit now %v", newLimit)),
			Next:    SettingsUpdated{},
		}
	case "script.lookup-path":
		if value == "" {
			return events.StatusMsg(fmt.Sprintf("script.lookup-path = '%v'", sc.settings.ScriptLookupPaths()))
		}

		if err := sc.settings.SetScriptLookupPaths(value); err != nil {
			return events.Error(err)
		}
		sc.bus.Fire(BusEventSettingsUpdated, name, value)
		return SettingsUpdated{}
	}

	return events.Error(errors.Errorf("unrecognised setting: %v", name))
}

func (sc *SettingsController) IsReadOnly() bool {
	ro, err := sc.settings.IsReadOnly()
	if err != nil {
		log.Printf("warn: unable to determine if R/O is available: %v", err)
		return false
	}
	return ro
}

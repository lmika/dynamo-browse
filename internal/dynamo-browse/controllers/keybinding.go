package controllers

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/dynamo-browse/internal/common/ui/events"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/services/keybindings"
	"github.com/pkg/errors"
)

type KeyBindingController struct {
	service             *keybindings.Service
	customBindingSource CustomKeyBindingSource
}

func NewKeyBindingController(service *keybindings.Service, customBindingSource CustomKeyBindingSource) *KeyBindingController {
	return &KeyBindingController{
		service:             service,
		customBindingSource: customBindingSource,
	}
}

func (kb *KeyBindingController) Rebind(bindingName string, newKey string, force bool) tea.Msg {
	existingBinding := kb.findExistingBinding(newKey)
	if existingBinding == "" {
		if err := kb.rebind(bindingName, newKey); err != nil {
			return events.Error(err)
		}
		return events.StatusMsg(fmt.Sprintf("Binding '%v' now bound to '%v'", bindingName, newKey))
	}

	//err := kb.rebind(bindingName, newKey, force)
	//if err == nil {
	//	return events.StatusMsg(fmt.Sprintf("Binding '%v' now bound to '%v'", bindingName, newKey))
	//} else if force {
	//	return events.Error(errors.Wrapf(err, "cannot bind '%v' to '%v'", bindingName, newKey))
	//}
	//
	//var keyAlreadyBoundErr keybindings.KeyAlreadyBoundError
	//if errors.As(err, &keyAlreadyBoundErr) {
	promptMsg := fmt.Sprintf("Key '%v' already bound to '%v'.  Continue? ", newKey, existingBinding)
	return events.ConfirmYes(promptMsg, func() tea.Msg {
		kb.unbindKey(newKey)

		err := kb.rebind(bindingName, newKey)
		if err != nil {
			return events.Error(err)
		}
		return events.StatusMsg(fmt.Sprintf("Binding '%v' now bound to '%v'", bindingName, newKey))
	})
	//}

	//return events.Error(err)
}

func (kb *KeyBindingController) rebind(bindingName string, newKey string) error {
	err := kb.service.Rebind(bindingName, newKey)
	if err == nil {
		return nil
	}

	var invalidBinding keybindings.InvalidBindingError
	if !errors.As(err, &invalidBinding) {
		return err
	}

	return kb.customBindingSource.Rebind(bindingName, newKey)
}

func (kb *KeyBindingController) unbindKey(key string) {
	kb.service.UnbindKey(key)
	kb.customBindingSource.UnbindKey(key)
}

func (kb *KeyBindingController) findExistingBinding(key string) string {
	if binding := kb.service.LookupBinding(key); binding != "" {
		return binding
	}

	return kb.customBindingSource.LookupBinding(key)
}

func (kb *KeyBindingController) LookupCustomBinding(key string) tea.Cmd {
	if kb.customBindingSource == nil {
		return nil
	}
	return kb.customBindingSource.CustomKeyCommand(key)
}

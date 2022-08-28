package controllers

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/audax/internal/common/ui/events"
	"github.com/lmika/audax/internal/dynamo-browse/services/keybindings"
	"github.com/pkg/errors"
)

type KeyBindingController struct {
	service *keybindings.Service
}

func NewKeyBindingController(service *keybindings.Service) *KeyBindingController {
	return &KeyBindingController{service: service}
}

func (kb *KeyBindingController) Rebind(bindingName string, newKey string, force bool) tea.Msg {
	err := kb.service.Rebind(bindingName, newKey, force)
	if err == nil {
		return events.SetStatus(fmt.Sprintf("Binding '%v' now bound to '%v'", bindingName, newKey))
	} else if force {
		return events.Error(errors.Wrapf(err, "cannot bind '%v' to '%v'", bindingName, newKey))
	}

	var keyAlreadyBoundErr keybindings.KeyAlreadyBoundError
	if errors.As(err, &keyAlreadyBoundErr) {
		promptMsg := fmt.Sprintf("Key '%v' already bound to '%v'.  Continue? ", keyAlreadyBoundErr.Key, keyAlreadyBoundErr.ExistingBindingName)
		return events.Confirm(promptMsg, func() tea.Msg {
			err := kb.service.Rebind(bindingName, newKey, true)
			if err != nil {
				return events.Error(err)
			}
			return events.SetStatus(fmt.Sprintf("Binding '%v' now bound to '%v'", bindingName, newKey))
		})
	}

	return events.Error(err)
}

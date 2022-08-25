package keybindings

import (
	"fmt"
)

type KeyAlreadyBoundError struct {
	Key                 string
	ExistingBindingName string
}

func (e KeyAlreadyBoundError) Error() string {
	return fmt.Sprintf("key '%v' already bound to '%v'", e.Key, e.ExistingBindingName)
}

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

type InvalidBindingError string

func (e InvalidBindingError) Error() string {
	return fmt.Sprintf("invalid binding: %v", string(e))
}

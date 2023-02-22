package keybindings

import (
	"github.com/charmbracelet/bubbles/key"
	"reflect"
	"strings"
)

type Service struct {
	keyBindingValue reflect.Value
}

func NewService(keyBinding any) *Service {
	v := reflect.ValueOf(keyBinding)
	if v.Kind() != reflect.Pointer {
		panic("keyBinding must be a pointer to a struct")
	}

	return &Service{
		keyBindingValue: v.Elem(),
	}
}

func (s *Service) LookupBinding(theKey string) string {
	var foundBinding = ""
	s.walkBindingFields(func(bindingName string, binding *key.Binding) bool {
		for _, boundKey := range binding.Keys() {
			if boundKey == theKey {
				foundBinding = bindingName
				return false
			}
		}
		return true
	})
	return foundBinding
}

func (s *Service) UnbindKey(theKey string) {
	s.walkBindingFields(func(bindingName string, binding *key.Binding) bool {
		for _, boundKey := range binding.Keys() {
			if boundKey == theKey {
				l := len(binding.Keys())
				if l == 1 {
					*binding = key.NewBinding()
				} else if l > 1 {
					newKeys := make([]string, 0)
					for _, k := range binding.Keys() {
						if k != theKey {
							newKeys = append(newKeys, k)
						}
					}
					*binding = key.NewBinding(key.WithKeys(newKeys...))
				}
			}
		}
		return true
	})
}

func (s *Service) Rebind(name string, newKey string) error {
	// Check if there already exists a binding (or clear it)
	//var foundBinding = ""
	//s.walkBindingFields(func(bindingName string, binding *key.Binding) bool {
	//	for _, boundKey := range binding.Keys() {
	//		if boundKey == newKey {
	//			if force {
	//				// TODO: only filter out "boundKey" rather clear
	//				log.Printf("clearing binding of %v", bindingName)
	//				*binding = key.NewBinding()
	//				return true
	//			} else {
	//				foundBinding = bindingName
	//				return false
	//			}
	//		}
	//	}
	//	return true
	//})
	//
	//if foundBinding != "" {
	//	return KeyAlreadyBoundError{Key: newKey, ExistingBindingName: foundBinding}
	//}

	// Rebind
	binding := s.findFieldForBinding(name)
	if binding == nil {
		return InvalidBindingError(name)
	}

	if len(binding.Keys()) == 0 {
		*binding = key.NewBinding(key.WithKeys(newKey))
	} else {
		newKeys := append([]string{newKey}, binding.Keys()...)
		*binding = key.NewBinding(key.WithKeys(newKeys...))
	}

	return nil
}

func (s *Service) findFieldForBinding(name string) *key.Binding {
	return s.findFieldForBindingInGroup(s.keyBindingValue, name)
}

func (s *Service) findFieldForBindingInGroup(group reflect.Value, name string) *key.Binding {
	bindingName, bindingSuffix, _ := strings.Cut(name, ".")

	groupType := group.Type()
	for i := 0; i < group.NumField(); i++ {
		fieldType := groupType.Field(i)

		keymapTag := fieldType.Tag.Get("keymap")
		if keymapTag != bindingName {
			continue
		}

		if fieldType.Type.Kind() == reflect.Pointer && fieldType.Type.Elem().Kind() == reflect.Struct {
			return s.findFieldForBindingInGroup(group.Field(i).Elem(), bindingSuffix)
		}

		binding, isBinding := group.Field(i).Addr().Interface().(*key.Binding)
		if !isBinding {
			return nil
		}
		return binding
	}
	return nil
}

func (s *Service) walkBindingFields(fn func(name string, binding *key.Binding) bool) {
	s.walkBindingFieldsInGroup(s.keyBindingValue, "", fn)
}

func (s *Service) walkBindingFieldsInGroup(group reflect.Value, prefix string, fn func(name string, binding *key.Binding) bool) bool {
	groupType := group.Type()
	for i := 0; i < group.NumField(); i++ {
		fieldType := groupType.Field(i)

		keymapTag := fieldType.Tag.Get("keymap")

		var fullName string
		if prefix != "" {
			fullName = prefix + "." + keymapTag
		} else {
			fullName = keymapTag
		}

		if fieldType.Type.Kind() == reflect.Pointer && fieldType.Type.Elem().Kind() == reflect.Struct {
			if !s.walkBindingFieldsInGroup(group.Field(i).Elem(), fullName, fn) {
				return false
			}
		}

		binding, isBinding := group.Field(i).Addr().Interface().(*key.Binding)
		if !isBinding {
			continue
		}
		if !fn(fullName, binding) {
			return false
		}
	}
	return true
}

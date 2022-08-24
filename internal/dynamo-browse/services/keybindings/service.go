package keybindings

import "reflect"

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

func (s *Service) Rebind(name string, key string) error {

}

func (s *Service) findFieldForBinding(name string) reflect.Value {

}

func (s *Service) findFieldForBindingInGroup(group reflect.Value, name string) reflect.Value {
	for i := 0; i < group.NumField(); i++ {
		group.Field(i).Type().
	}
}

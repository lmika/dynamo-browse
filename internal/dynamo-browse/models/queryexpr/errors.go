package queryexpr

import (
	"fmt"
	"strings"
)

// NameNotFoundError is returned if the given name cannot be found
type NameNotFoundError string

func (n NameNotFoundError) Error() string {
	return fmt.Sprintf("%v: name not found", string(n))
}

// ValueNotAMapError is return if the given name is not a map
type ValueNotAMapError []string

func (n ValueNotAMapError) Error() string {
	return fmt.Sprintf("%v: name is not a map", strings.Join(n, "."))
}

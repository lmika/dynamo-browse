package itemrender

import "fmt"

func cardinality(c int, single, multi string) string {
	if c == 1 {
		return fmt.Sprintf("(%d %v)", c, single)
	}
	return fmt.Sprintf("(%d %v)", c, multi)
}

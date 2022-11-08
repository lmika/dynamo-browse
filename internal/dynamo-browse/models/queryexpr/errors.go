package queryexpr

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models/attrutils"
	"github.com/lmika/audax/internal/dynamo-browse/models/itemrender"
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

// ValuesNotComparable indicates that two values are not comparable
type ValuesNotComparable struct {
	Left, Right types.AttributeValue
}

func (n ValuesNotComparable) Error() string {
	leftStr, _ := attrutils.AttributeToString(n.Left)
	rightStr, _ := attrutils.AttributeToString(n.Right)
	return fmt.Sprintf("values '%v' and '%v' are not comparable", leftStr, rightStr)
}

// ValueNotConvertableToString indicates that a value is not convertable to a string
type ValueNotConvertableToString struct {
	Val types.AttributeValue
}

func (n ValueNotConvertableToString) Error() string {
	render := itemrender.ToRenderer(n.Val)
	return fmt.Sprintf("values '%v', type %v, is not convertable to string", render.StringValue(), render.TypeName())
}

type ValueMustBeLiteral struct{}

func (n ValueMustBeLiteral) Error() string {
	return "value must be a literal"
}

package queryexpr

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models/attrutils"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models/itemrender"
	"strings"
)

// NameNotFoundError is returned if the given name cannot be found
type NameNotFoundError string

func (n NameNotFoundError) Error() string {
	return fmt.Sprintf("%v: name not found", string(n))
}

type OperandNotANameError string

func (n OperandNotANameError) Error() string {
	return fmt.Sprintf("operand '%v' is not a name", string(n))
}

type OperandNotAnOperandError struct{}

func (n OperandNotAnOperandError) Error() string {
	return "element must be an operand"
}

// ValueNotAMapError is return if the given name is not a map
type ValueNotAMapError []string

func (n ValueNotAMapError) Error() string {
	return fmt.Sprintf("%v: name is not a map", strings.Join(n, "."))
}

// ValueNotAListError is return if the given name is not a map
type ValueNotAListError []string

func (n ValueNotAListError) Error() string {
	return fmt.Sprintf("%v: name is not a list", strings.Join(n, "."))
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

// ValuesNotInnable indicates that a values cannot be used on the right side of an in
type ValuesNotInnableError struct {
	Val types.AttributeValue
}

func (n ValuesNotInnableError) Error() string {
	leftStr, _ := attrutils.AttributeToString(n.Val)
	return fmt.Sprintf("values '%v' cannot be used as the right side of an 'in'", leftStr)
}

// ValueNotConvertableToString indicates that a value is not convertable to a string
type ValueNotConvertableToString struct {
	Val types.AttributeValue
}

func (n ValueNotConvertableToString) Error() string {
	render := itemrender.ToRenderer(n.Val)
	return fmt.Sprintf("values '%v', type %v, is not convertable to string", render.StringValue(), render.TypeName())
}

type NodeCannotBeConvertedToQueryError struct{}

func (n NodeCannotBeConvertedToQueryError) Error() string {
	return "node cannot be converted to query"
}

type ValueMustBeLiteralError struct{}

func (n ValueMustBeLiteralError) Error() string {
	return "value must be a literal"
}

type ValueMustBeStringError struct{}

func (n ValueMustBeStringError) Error() string {
	return "value must be a string"
}

type InvalidTypeForIsError struct {
	TypeName string
}

func (n InvalidTypeForIsError) Error() string {
	return "invalid type for 'is': " + n.TypeName
}

type InvalidTypeForBetweenError struct {
	TypeName string
}

func (n InvalidTypeForBetweenError) Error() string {
	return "invalid type for 'between': " + n.TypeName
}

type InvalidArgumentNumberError struct {
	Name     string
	Expected int
	Actual   int
}

func (e InvalidArgumentNumberError) Error() string {
	return fmt.Sprintf("function '%v' expected %v args but received %v", e.Name, e.Expected, e.Actual)
}

type InvalidArgumentTypeError struct {
	Name     string
	ArgIndex int
	Expected string
}

func (e InvalidArgumentTypeError) Error() string {
	return fmt.Sprintf("function '%v' expected arg %v to be of type %v", e.Name, e.ArgIndex, e.Expected)
}

type UnrecognisedFunctionError struct {
	Name string
}

func (e UnrecognisedFunctionError) Error() string {
	return "unrecognised function '" + e.Name + "'"
}

type PathNotSettableError struct {
}

func (e PathNotSettableError) Error() string {
	return "path cannot be set a value"
}

type MissingPlaceholderError struct {
	Placeholder string
}

func (e MissingPlaceholderError) Error() string {
	return "undefined placeholder '" + e.Placeholder + "'"
}

type ValueNotUsableAsASubref struct {
}

func (e ValueNotUsableAsASubref) Error() string {
	return "value cannot be used as a subref"
}

type MultiplePlansWithIndexError struct {
	PossibleIndices []string
}

func (e MultiplePlansWithIndexError) Error() string {
	return fmt.Sprintf("multiple plans with index found. Specify index or scan with 'using' clause: possible indices are %v", e.PossibleIndices)
}

type NoPlausiblePlanWithIndexError struct {
	PreferredIndex  string
	PossibleIndices []string
}

func (e NoPlausiblePlanWithIndexError) Error() string {
	return fmt.Sprintf("no plan with index '%v' found: possible indices are %v", e.PreferredIndex, e.PossibleIndices)
}

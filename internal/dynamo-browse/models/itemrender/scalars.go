package itemrender

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type StringRenderer types.AttributeValueMemberS

func (sr *StringRenderer) TypeName() string {
	return "S"
}

func (sr *StringRenderer) StringValue() string {
	return sr.Value
}

func (sr *StringRenderer) SubItems() []SubItem {
	return nil
}

type NumberRenderer types.AttributeValueMemberN

func (sr *NumberRenderer) TypeName() string {
	return "N"
}

func (sr *NumberRenderer) StringValue() string {
	return sr.Value
}

func (sr *NumberRenderer) SubItems() []SubItem {
	return nil
}

type BoolRenderer types.AttributeValueMemberBOOL

func (sr *BoolRenderer) TypeName() string {
	return "BOOL"
}

func (sr *BoolRenderer) StringValue() string {
	if sr.Value {
		return "True"
	}
	return "False"
}

func (sr *BoolRenderer) SubItems() []SubItem {
	return nil
}

type BinaryRenderer types.AttributeValueMemberB

func (sr *BinaryRenderer) TypeName() string {
	return "B"
}

func (sr *BinaryRenderer) StringValue() string {
	if len(sr.Value) == 1 {
		return fmt.Sprintf("(1 byte)")
	}
	return fmt.Sprintf("(%d bytes)", len(sr.Value))
}

func (sr *BinaryRenderer) SubItems() []SubItem {
	return nil
}

type NullRenderer types.AttributeValueMemberNULL

func (sr *NullRenderer) TypeName() string {
	return "NULL"
}

func (sr *NullRenderer) StringValue() string {
	return "null"
}

func (sr *NullRenderer) SubItems() []SubItem {
	return nil
}

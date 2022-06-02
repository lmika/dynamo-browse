package itemrender

import "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

type Renderer interface {
	TypeName() string
	StringValue() string
	SubItems() []SubItem
}

func ToRenderer(v types.AttributeValue) Renderer {
	switch colVal := v.(type) {
	case nil:
		return nil
	case *types.AttributeValueMemberS:
		x := StringRenderer(*colVal)
		return &x
	case *types.AttributeValueMemberN:
		x := NumberRenderer(*colVal)
		return &x
	case *types.AttributeValueMemberBOOL:
		x := BoolRenderer(*colVal)
		return &x
	case *types.AttributeValueMemberNULL:
		x := NullRenderer(*colVal)
		return &x
	case *types.AttributeValueMemberB:
		x := BinaryRenderer(*colVal)
		return &x
	case *types.AttributeValueMemberL:
		x := ListRenderer(*colVal)
		return &x
	case *types.AttributeValueMemberM:
		x := MapRenderer(*colVal)
		return &x
	case *types.AttributeValueMemberBS:
		return newBinarySetRenderer(colVal)
	case *types.AttributeValueMemberNS:
		return newNumberSetRenderer(colVal)
	case *types.AttributeValueMemberSS:
		return newStringSetRenderer(colVal)
	}
	return OtherRenderer{}
}

type SubItem struct {
	Key   string
	Value Renderer
}

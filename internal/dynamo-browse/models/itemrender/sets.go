package itemrender

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type GenericRenderer struct {
	typeName     string
	subitemValue []Renderer
}

func (sr *GenericRenderer) TypeName() string {
	return sr.typeName
}

func (sr *GenericRenderer) StringValue() string {
	return cardinality(len(sr.subitemValue), "item", "items")
}

func (sr *GenericRenderer) SubItems() []SubItem {
	subitems := make([]SubItem, len(sr.subitemValue))
	for i, r := range sr.subitemValue {
		subitems[i] = SubItem{Key: fmt.Sprint(i), Value: r}
	}
	return subitems
}

func newBinarySetRenderer(v *types.AttributeValueMemberBS) *GenericRenderer {
	vs := make([]Renderer, len(v.Value))
	for i, b := range v.Value {
		vs[i] = &BinaryRenderer{Value: b}
	}
	return &GenericRenderer{typeName: "BS", subitemValue: vs}
}

func newNumberSetRenderer(v *types.AttributeValueMemberNS) *GenericRenderer {
	vs := make([]Renderer, len(v.Value))
	for i, n := range v.Value {
		vs[i] = &NumberRenderer{Value: n}
	}
	return &GenericRenderer{typeName: "NS", subitemValue: vs}
}

func newStringSetRenderer(v *types.AttributeValueMemberSS) *GenericRenderer {
	vs := make([]Renderer, len(v.Value))
	for i, s := range v.Value {
		vs[i] = &StringRenderer{Value: s}
	}
	return &GenericRenderer{typeName: "SS", subitemValue: vs}
}

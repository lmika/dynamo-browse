package itemrender

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"sort"
)

type ListRenderer types.AttributeValueMemberL

func (sr *ListRenderer) TypeName() string {
	return "L"
}

func (sr *ListRenderer) StringValue() string {
	return ""
}

func (sr *ListRenderer) MetaInfo() string {
	if len(sr.Value) == 1 {
		return fmt.Sprintf("(1 item)")
	}
	return fmt.Sprintf("(%d items)", len(sr.Value))
}

func (sr *ListRenderer) SubItems() []SubItem {
	subitems := make([]SubItem, len(sr.Value))
	for i, r := range sr.Value {
		subitems[i] = SubItem{Key: fmt.Sprint(i), Value: ToRenderer(r)}
	}
	return subitems
}

type MapRenderer types.AttributeValueMemberM

func (sr *MapRenderer) TypeName() string {
	return "M"
}

func (sr *MapRenderer) StringValue() string {
	return ""
}

func (sr *MapRenderer) MetaInfo() string {
	if len(sr.Value) == 1 {
		return fmt.Sprintf("(1 item)")
	}
	return fmt.Sprintf("(%d items)", len(sr.Value))
}

func (sr *MapRenderer) SubItems() []SubItem {
	subitems := make([]SubItem, 0)
	for k, r := range sr.Value {
		subitems = append(subitems, SubItem{Key: k, Value: ToRenderer(r)})
	}
	sort.Slice(subitems, func(i, j int) bool {
		return subitems[i].Key < subitems[j].Key
	})
	return subitems
}

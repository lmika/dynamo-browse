package scriptmanager

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cloudcmds/tamarin/object"
	"github.com/lmika/dynamo-browse/internal/common/maputils"
	"github.com/lmika/dynamo-browse/internal/common/sliceutils"
	"github.com/pkg/errors"
	"regexp"
	"strconv"
)

func tamarinValueToAttributeValue(val object.Object) (types.AttributeValue, error) {
	switch v := val.(type) {
	case *object.String:
		return &types.AttributeValueMemberS{Value: v.Value()}, nil
	case *object.Int:
		return &types.AttributeValueMemberN{Value: strconv.FormatInt(v.Value(), 10)}, nil
	case *object.Float:
		return &types.AttributeValueMemberN{Value: strconv.FormatFloat(v.Value(), 'f', -1, 64)}, nil
	case *object.Bool:
		return &types.AttributeValueMemberBOOL{Value: v.Value()}, nil
	case *object.NilType:
		return &types.AttributeValueMemberNULL{Value: true}, nil
	case *object.List:
		attrValue, err := sliceutils.MapWithError(v.Value(), tamarinValueToAttributeValue)
		if err != nil {
			return nil, err
		}
		return &types.AttributeValueMemberL{Value: attrValue}, nil
	case *object.Map:
		attrValue, err := maputils.MapValuesWithError(v.Value(), tamarinValueToAttributeValue)
		if err != nil {
			return nil, err
		}
		return &types.AttributeValueMemberM{Value: attrValue}, nil
	case *object.Set:
		values := maputils.Values(v.Value())
		canBeNumSet := sliceutils.All(values, func(t object.Object) bool {
			_, isInt := t.(*object.Int)
			_, isFloat := t.(*object.Float)
			return isInt || isFloat
		})

		if canBeNumSet {
			return &types.AttributeValueMemberNS{
				Value: sliceutils.Map(values, func(t object.Object) string {
					switch v := t.(type) {
					case *object.Int:
						return strconv.FormatInt(v.Value(), 10)
					case *object.Float:
						return strconv.FormatFloat(v.Value(), 'f', -1, 64)
					}
					panic(fmt.Sprintf("unhandled object type: %v", t.Type()))
				}),
			}, nil
		}
		return &types.AttributeValueMemberSS{
			Value: sliceutils.Map(values, func(t object.Object) string {
				v, _ := object.AsString(t)
				return v
			}),
		}, nil
	}
	return nil, errors.Errorf("type error: unsupported value type (got %v)", val.Type())
}

func attributeValueToTamarin(val types.AttributeValue) (object.Object, error) {
	if val == nil {
		return object.Nil, nil
	}

	switch v := val.(type) {
	case *types.AttributeValueMemberS:
		return object.NewString(v.Value), nil
	case *types.AttributeValueMemberN:
		f, err := convertNumAttributeToTamarinValue(v.Value)
		if err != nil {
			return nil, errors.Errorf("value error: invalid N value: %v", v.Value)
		}
		return f, nil
	case *types.AttributeValueMemberBOOL:
		if v.Value {
			return object.True, nil
		}
		return object.False, nil
	case *types.AttributeValueMemberNULL:
		return object.Nil, nil
	case *types.AttributeValueMemberL:
		list, err := sliceutils.MapWithError(v.Value, attributeValueToTamarin)
		if err != nil {
			return nil, err
		}
		return object.NewList(list), nil
	case *types.AttributeValueMemberM:
		objMap, err := maputils.MapValuesWithError(v.Value, attributeValueToTamarin)
		if err != nil {
			return nil, err
		}
		return object.NewMap(objMap), nil
	case *types.AttributeValueMemberSS:
		return object.NewSet(sliceutils.Map(v.Value, func(s string) object.Object {
			return object.NewString(s)
		})), nil
	case *types.AttributeValueMemberNS:
		nums, err := sliceutils.MapWithError(v.Value, func(s string) (object.Object, error) {
			return convertNumAttributeToTamarinValue(s)
		})
		if err != nil {
			return nil, err
		}
		return object.NewSet(nums), nil
	}
	return nil, errors.Errorf("value error: cannot convert type %T to tamarin object", val)
}

var intNumberPattern = regexp.MustCompile(`^[-]?[0-9]+$`)

// XXX - this is pretty crappy in that it does not support large values
func convertNumAttributeToTamarinValue(n string) (object.Object, error) {
	if intNumberPattern.MatchString(n) {
		parsedInt, err := strconv.ParseInt(n, 10, 64)
		if err != nil {
			return nil, err
		}
		return object.NewInt(parsedInt), nil
	}

	f, err := strconv.ParseFloat(n, 64)
	if err != nil {
		return nil, err
	}
	return object.NewFloat(f), nil
}

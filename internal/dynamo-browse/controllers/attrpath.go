package controllers

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/awstools/internal/dynamo-browse/models"
	"github.com/pkg/errors"
	"strings"
)

type attrPath []string

func newAttrPath(expr string) attrPath {
	return strings.Split(expr, ".")
}

func (ap attrPath) follow(item models.Item) (types.AttributeValue, error) {
	var step types.AttributeValue
	for i, seg := range ap {
		if i == 0 {
			step = item[seg]
			continue
		}

		switch s := step.(type) {
		case *types.AttributeValueMemberM:
			step = s.Value[seg]
		default:
			return nil, errors.Errorf("seg %v expected to be a map", i)
		}
	}
	return step, nil
}

func (ap attrPath) setAt(item models.Item, newValue types.AttributeValue) error {
	if len(ap) == 1 {
		item[ap[0]] = newValue
		return nil
	}

	var step types.AttributeValue
	for i, seg := range ap[:len(ap)-1] {
		if i == 0 {
			step = item[seg]
			continue
		}

		switch s := step.(type) {
		case *types.AttributeValueMemberM:
			step = s.Value[seg]
		default:
			return errors.Errorf("seg %v expected to be a map", i)
		}
	}

	lastSeg := ap[len(ap)-1]
	switch s := step.(type) {
	case *types.AttributeValueMemberM:
		s.Value[lastSeg] = newValue
	default:
		return errors.Errorf("last seg expected to be a map, but was %T", lastSeg)
	}

	return nil
}

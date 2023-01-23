package attrcodec

import (
	"bytes"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/pkg/errors"
)

func SerializeMapToBytes(ms map[string]types.AttributeValue) ([]byte, error) {
	bs := new(bytes.Buffer)
	if err := NewEncoder(bs).Encode(&types.AttributeValueMemberM{
		Value: ms,
	}); err != nil {
		return nil, err
	}
	return bs.Bytes(), nil
}

func DeseralizedMapFromBytes(bs []byte) (map[string]types.AttributeValue, error) {
	attr, err := NewDecoder(bytes.NewReader(bs)).Decode()
	if err != nil {
		return nil, err
	}

	mapAttr, isMapAttr := attr.(*types.AttributeValueMemberM)
	if !isMapAttr {
		return nil, errors.New("expected attribute value to be a map")
	}
	return mapAttr.Value, nil
}

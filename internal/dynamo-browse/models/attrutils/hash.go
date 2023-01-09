package attrutils

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"hash"
	"hash/fnv"
)

func HashCode(x types.AttributeValue) uint64 {
	h := fnv.New64a()
	doHash(x, h)
	return h.Sum64()
}

func HashTo(h hash.Hash, x types.AttributeValue) {
	doHash(x, h)
}

func doHash(x types.AttributeValue, h hash.Hash) {
	switch xVal := x.(type) {
	case *types.AttributeValueMemberS:
		h.Write([]byte(xVal.Value))
	case *types.AttributeValueMemberN:
		h.Write([]byte(xVal.Value))
	case *types.AttributeValueMemberBOOL:
		if xVal.Value {
			h.Write([]byte{0})
		} else {
			h.Write([]byte{1})
		}
	case *types.AttributeValueMemberB:
		h.Write(xVal.Value)
	case *types.AttributeValueMemberNULL:
		if xVal.Value {
			h.Write([]byte{0})
		} else {
			h.Write([]byte{1})
		}
	case *types.AttributeValueMemberL:
		for _, v := range xVal.Value {
			doHash(v, h)
		}
	case *types.AttributeValueMemberM:
		// To keep this consistent, this will need to be in key sorted order
		sortedKeys := make([]string, len(xVal.Value))
		copy(sortedKeys, maps.Keys(xVal.Value))
		slices.Sort(sortedKeys)

		for _, k := range sortedKeys {
			h.Write([]byte(k))
			doHash(xVal.Value[k], h)
		}
	case *types.AttributeValueMemberBS:
		for _, v := range xVal.Value {
			h.Write(v)
		}
	case *types.AttributeValueMemberNS:
		for _, v := range xVal.Value {
			h.Write([]byte(v))
		}
	case *types.AttributeValueMemberSS:
		for _, v := range xVal.Value {
			h.Write([]byte(v))
		}
	}
}

package queryexpr

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"strings"
)

func (dt *astDot) unqualifiedName() (string, bool) {
	if len(dt.Quals) == 0 {
		return dt.Name, true
	}
	return "", false
}

func (dt *astDot) evalItem(item models.Item) (types.AttributeValue, error) {
	res, hasV := item[dt.Name]
	if !hasV {
		return nil, NameNotFoundError(dt.String())
	}

	for i, qualName := range dt.Quals {
		mapRes, isMapRes := res.(*types.AttributeValueMemberM)
		if !isMapRes {
			return nil, ValueNotAMapError(append([]string{dt.Name}, dt.Quals[:i+1]...))
		}

		res, hasV = mapRes.Value[qualName]
		if !hasV {
			return nil, NameNotFoundError(dt.String())
		}
	}

	return res, nil
}

func (a *astDot) String() string {
	var sb strings.Builder

	sb.WriteString(a.Name)
	for _, q := range a.Quals {
		sb.WriteRune('.')
		sb.WriteString(q)
	}

	return sb.String()
}

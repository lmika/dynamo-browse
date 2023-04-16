package modexpr

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models"
)

type patchMod interface {
	Apply(item models.Item)
}

type setAttributeMod struct {
	key string
	to  types.AttributeValue
}

func (sa setAttributeMod) Apply(item models.Item) {
	item[sa.key] = sa.to
}

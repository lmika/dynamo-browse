package modexpr

import "github.com/lmika/awstools/internal/dynamo-browse/models"

type ModExpr struct {
	ast *astExpr
}

func (me *ModExpr) Patch(item models.Item) (models.Item, error) {
	newItem := item.Clone()

	for _, attribute := range me.ast.Attributes {
		var err error
		name := attribute.Name
		newItem[name], err = attribute.dynamoValue()
		if err != nil {
			return nil, err
		}
	}

	return newItem, nil
}

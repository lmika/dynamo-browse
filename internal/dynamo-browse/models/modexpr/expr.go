package modexpr

import "github.com/lmika/dynamo-browse/internal/dynamo-browse/models"

type ModExpr struct {
	ast *astExpr
}

func (me *ModExpr) Patch(item models.Item) (models.Item, error) {
	mods, err := me.ast.calcPatchMods(item)
	if err != nil {
		return nil, err
	}

	newItem := item.Clone()
	for _, mod := range mods {
		mod.Apply(newItem)
	}

	return newItem, nil
}

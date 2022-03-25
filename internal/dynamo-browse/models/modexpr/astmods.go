package modexpr

import "github.com/lmika/awstools/internal/dynamo-browse/models"

func (a *astExpr) calcPatchMods(item models.Item) ([]patchMod, error) {
	patchMods := make([]patchMod, 0)

	for _, attr := range a.Attributes {
		attrPatchMods, err := attr.calcPatchMods(item)
		if err != nil {
			return nil, err
		}
		patchMods = append(patchMods, attrPatchMods...)
	}

	return patchMods, nil
}

func (a *astAttribute) calcPatchMods(item models.Item) ([]patchMod, error) {
	value, err := a.Value.dynamoValue()
	if err != nil {
		return nil, err
	}

	patchMods := make([]patchMod, 0)
	for _, key := range a.Names.Names {
		patchMods = append(patchMods, setAttributeMod{key: key, to: value})
	}

	return patchMods, nil
}

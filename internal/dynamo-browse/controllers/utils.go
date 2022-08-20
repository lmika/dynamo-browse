package controllers

import "github.com/lmika/audax/internal/dynamo-browse/models"

func applyToMarkedItems(rs *models.ResultSet, selectedIndex int, applyFn func(idx int, item models.Item) error) error {
	if markedItems := rs.MarkedItems(); len(markedItems) > 0 {
		for _, mi := range markedItems {
			if err := applyFn(mi.Index, mi.Item); err != nil {
				return err
			}
		}
		return nil
	}

	return applyFn(selectedIndex, rs.Items()[selectedIndex])
}

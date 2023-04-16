package dynamoitemview

import "github.com/lmika/dynamo-browse/internal/dynamo-browse/models"

type NewItemSelected struct {
	ResultSet *models.ResultSet
	Item      models.Item
}

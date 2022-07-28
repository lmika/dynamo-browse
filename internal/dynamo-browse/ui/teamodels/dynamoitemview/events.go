package dynamoitemview

import "github.com/lmika/audax/internal/dynamo-browse/models"

type NewItemSelected struct {
	ResultSet *models.ResultSet
	Item      models.Item
}

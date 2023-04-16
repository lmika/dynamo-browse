package scriptmanager

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models"
)

//go:generate mockery --with-expecter --name UIService
//go:generate mockery --with-expecter --name SessionService

type Ifaces struct {
	UI      UIService
	Session SessionService
}

type UIService interface {
	PrintMessage(ctx context.Context, msg string)

	// Prompt should return a channel which will provide the input from the user.  If the user
	// provides no input, prompt should close the channel without providing anything.
	Prompt(ctx context.Context, msg string) chan string
}

type SessionService interface {
	Query(ctx context.Context, expr string, queryOptions QueryOptions) (*models.ResultSet, error)

	ResultSet(ctx context.Context) *models.ResultSet
	SelectedItemIndex(ctx context.Context) int
	SetResultSet(ctx context.Context, newResultSet *models.ResultSet)
}

type QueryOptions struct {
	TableName         string
	IndexName         string
	NamePlaceholders  map[string]string
	ValuePlaceholders map[string]types.AttributeValue
}

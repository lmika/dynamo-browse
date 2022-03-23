package dynamo

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/awstools/internal/dynamo-browse/models"
	"github.com/pkg/errors"
)

type Provider struct {
	client *dynamodb.Client
}

func NewProvider(client *dynamodb.Client) *Provider {
	return &Provider{client: client}
}

func (p *Provider) ScanItems(ctx context.Context, tableName string) ([]models.Item, error) {
	res, err := p.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "cannot execute scan on table %v", tableName)
	}

	items := make([]models.Item, len(res.Items))
	for i, itm := range res.Items {
		items[i] = itm
	}

	return items, nil
}

func (p *Provider) DeleteItem(ctx context.Context, tableName string, key map[string]types.AttributeValue) error {
	_, err := p.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: key,
	})
	return errors.Wrap(err, "could not delete item")
}
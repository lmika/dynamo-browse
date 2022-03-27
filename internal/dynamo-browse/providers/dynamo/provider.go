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

func (p *Provider) ListTables(ctx context.Context) ([]string, error) {
	out, err := p.client.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		return nil, errors.Wrapf(err, "cannot list tables")
	}

	return out.TableNames, nil
}

func (p *Provider) DescribeTable(ctx context.Context, tableName string) (*models.TableInfo, error) {
	out, err := p.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "cannot describe table %v", tableName)
	}

	var tableInfo models.TableInfo
	tableInfo.Name = aws.ToString(out.Table.TableName)

	for _, keySchema := range out.Table.KeySchema {
		if keySchema.KeyType == types.KeyTypeHash {
			tableInfo.Keys.PartitionKey = aws.ToString(keySchema.AttributeName)
		} else if keySchema.KeyType == types.KeyTypeRange {
			tableInfo.Keys.SortKey = aws.ToString(keySchema.AttributeName)
		}
	}

	for _, definedAttribute := range out.Table.AttributeDefinitions {
		tableInfo.DefinedAttributes = append(tableInfo.DefinedAttributes, aws.ToString(definedAttribute.AttributeName))
	}

	return &tableInfo, nil
}

func (p *Provider) PutItem(ctx context.Context, name string, item models.Item) error {
	_, err := p.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(name),
		Item:      item,
	})
	if err != nil {
		return errors.Wrapf(err, "cannot execute put on table %v", name)
	}
	return nil
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
		Key:       key,
	})
	return errors.Wrap(err, "could not delete item")
}

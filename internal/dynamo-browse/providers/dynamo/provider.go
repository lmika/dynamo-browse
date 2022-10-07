package dynamo

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/audax/internal/common/sliceutils"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/services/jobs"
	"github.com/pkg/errors"
	"log"
	"time"
)

type Provider struct {
	client *dynamodb.Client
}

func NewProvider(client *dynamodb.Client) *Provider {
	return &Provider{client: client}
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

func (p *Provider) PutItems(ctx context.Context, name string, items []models.Item) error {
	return p.batchPutItems(ctx, name, items)
}

func (p *Provider) batchPutItems(ctx context.Context, name string, items []models.Item) error {
	nextUpdate := time.Now().Add(1 * time.Second)

	reqs := len(items)/25 + 1
	for rn := 0; rn < reqs; rn++ {
		s, f := rn*25, (rn+1)*25
		if f > len(items) {
			f = len(items)
		}

		itemsInThisRequest := items[s:f]
		writeRequests := sliceutils.Map(itemsInThisRequest, func(item models.Item) types.WriteRequest {
			return types.WriteRequest{PutRequest: &types.PutRequest{Item: item}}
		})

		log.Printf("Page")
		_, err := p.client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				name: writeRequests,
			},
		})
		if err != nil {
			errors.Wrapf(err, "unable to put page %v of back puts", rn)
		}

		if time.Now().After(nextUpdate) {
			jobs.PostUpdate(ctx, fmt.Sprintf("updated %d items", f))
			nextUpdate = time.Now().Add(1 * time.Second)
		}
	}
	return nil
}

func (p *Provider) ScanItems(ctx context.Context, tableName string, filterExpr *expression.Expression, maxItems int) ([]models.Item, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
		Limit:     aws.Int32(int32(maxItems)),
	}
	if filterExpr != nil {
		input.FilterExpression = filterExpr.Filter()
		input.ExpressionAttributeNames = filterExpr.Names()
		input.ExpressionAttributeValues = filterExpr.Values()
	}

	paginator := dynamodb.NewScanPaginator(p.client, input)

	items := make([]models.Item, 0)

	nextUpdate := time.Now().Add(1 * time.Second)

outer:
	for paginator.HasMorePages() {
		res, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot execute scan on table %v", tableName)
		}

		for _, itm := range res.Items {
			items = append(items, itm)
			if len(items) >= maxItems {
				break outer
			}

			if time.Now().After(nextUpdate) {
				jobs.PostUpdate(ctx, fmt.Sprintf("found %d items", len(items)))
				nextUpdate = time.Now().Add(1 * time.Second)
			}
		}
	}

	return items, nil
}

func (p *Provider) QueryItems(ctx context.Context, tableName string, filterExpr *expression.Expression, maxItems int) ([]models.Item, error) {
	input := &dynamodb.QueryInput{
		TableName: aws.String(tableName),
		Limit:     aws.Int32(int32(maxItems)),
	}
	if filterExpr != nil {
		input.KeyConditionExpression = filterExpr.KeyCondition()
		input.FilterExpression = filterExpr.Filter()
		input.ExpressionAttributeNames = filterExpr.Names()
		input.ExpressionAttributeValues = filterExpr.Values()
	}

	paginator := dynamodb.NewQueryPaginator(p.client, input)

	items := make([]models.Item, 0)

	nextUpdate := time.Now().Add(1 * time.Second)

outer:
	for paginator.HasMorePages() {
		res, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot execute query on table %v", tableName)
		}

		for _, itm := range res.Items {
			items = append(items, itm)
			if len(items) >= maxItems {
				break outer
			}

			if time.Now().After(nextUpdate) {
				jobs.PostUpdate(ctx, fmt.Sprintf("found %d items", len(items)))
				nextUpdate = time.Now().Add(1 * time.Second)
			}
		}
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

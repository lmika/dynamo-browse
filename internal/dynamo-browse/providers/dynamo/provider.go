package dynamo

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lmika/dynamo-browse/internal/common/sliceutils"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/services/jobs"
	"github.com/pkg/errors"
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
	tableInfo.Keys = p.keySchemaToKeyAttributes(out.Table.KeySchema)

	tableInfo.GSIs = make([]models.TableGSI, len(out.Table.GlobalSecondaryIndexes))
	for i, gsiIndex := range out.Table.GlobalSecondaryIndexes {
		tableInfo.GSIs[i] = models.TableGSI{
			Name: aws.ToString(gsiIndex.IndexName),
			Keys: p.keySchemaToKeyAttributes(gsiIndex.KeySchema),
		}
	}

	for _, definedAttribute := range out.Table.AttributeDefinitions {
		tableInfo.DefinedAttributes = append(tableInfo.DefinedAttributes, aws.ToString(definedAttribute.AttributeName))
	}

	return &tableInfo, nil
}

func (p *Provider) keySchemaToKeyAttributes(keySchemaElements []types.KeySchemaElement) (keyAttribute models.KeyAttribute) {
	for _, keySchema := range keySchemaElements {
		if keySchema.KeyType == types.KeyTypeHash {
			keyAttribute.PartitionKey = aws.ToString(keySchema.AttributeName)
		} else if keySchema.KeyType == types.KeyTypeRange {
			keyAttribute.SortKey = aws.ToString(keySchema.AttributeName)
		}
	}
	return keyAttribute
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

func (p *Provider) ScanItems(
	ctx context.Context,
	tableName string,
	filterExpr *expression.Expression,
	exclusiveStartKey map[string]types.AttributeValue,
	maxItems int,
) ([]models.Item, map[string]types.AttributeValue, error) {
	const maxItemsPerPage = 100

	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
		//Limit:             aws.Int32(int32(maxItems)),
		//Limit:             aws.Int32(100),
		//ExclusiveStartKey: exclusiveStartKey,
	}
	if filterExpr != nil {
		input.FilterExpression = filterExpr.Filter()
		input.ExpressionAttributeNames = filterExpr.Names()
		input.ExpressionAttributeValues = filterExpr.Values()
	}

	var (
		items       = make([]models.Item, 0)
		nextUpdate  = time.Now().Add(1 * time.Second)
		lastEvalKey = exclusiveStartKey
	)

	for len(items) < maxItems {
		remainingItemsToFetch := maxItems - len(items)
		if remainingItemsToFetch > maxItemsPerPage {
			input.Limit = aws.Int32(maxItemsPerPage)
		} else {
			input.Limit = aws.Int32(int32(remainingItemsToFetch))
		}
		input.ExclusiveStartKey = lastEvalKey

		out, err := p.client.Scan(ctx, input)
		if err != nil {
			if ctx.Err() != nil {
				return items, nil, models.NewPartialResultsError(ctx.Err())
			}
			return nil, nil, errors.Wrapf(err, "cannot execute scan on table %v", tableName)
		}

		for _, itm := range out.Items {
			items = append(items, itm)
		}

		if time.Now().After(nextUpdate) {
			jobs.PostUpdate(ctx, fmt.Sprintf("found %d items", len(items)))
			nextUpdate = time.Now().Add(1 * time.Second)
		}

		lastEvalKey = out.LastEvaluatedKey
		if lastEvalKey == nil {
			// We've reached the last page
			break
		}
	}

	return items, lastEvalKey, nil
}

func (p *Provider) QueryItems(
	ctx context.Context,
	tableName string,
	indexName string,
	filterExpr *expression.Expression,
	exclusiveStartKey map[string]types.AttributeValue,
	maxItems int,
) ([]models.Item, map[string]types.AttributeValue, error) {
	const maxItemsPerPage = 100

	input := &dynamodb.QueryInput{
		TableName: aws.String(tableName),
	}
	if indexName != "" {
		input.IndexName = aws.String(indexName)
	}
	if filterExpr != nil {
		input.KeyConditionExpression = filterExpr.KeyCondition()
		input.FilterExpression = filterExpr.Filter()
		input.ExpressionAttributeNames = filterExpr.Names()
		input.ExpressionAttributeValues = filterExpr.Values()
	}

	var (
		items       = make([]models.Item, 0)
		nextUpdate  = time.Now().Add(1 * time.Second)
		lastEvalKey = exclusiveStartKey
	)

	for len(items) < maxItems {
		remainingItemsToFetch := maxItems - len(items)
		if remainingItemsToFetch > maxItemsPerPage {
			input.Limit = aws.Int32(maxItemsPerPage)
		} else {
			input.Limit = aws.Int32(int32(remainingItemsToFetch))
		}
		input.ExclusiveStartKey = lastEvalKey

		out, err := p.client.Query(ctx, input)
		if err != nil {
			if ctx.Err() != nil {
				return items, nil, models.NewPartialResultsError(ctx.Err())
			}
			return nil, nil, errors.Wrapf(err, "cannot execute scan on table %v", tableName)
		}

		for _, itm := range out.Items {
			items = append(items, itm)
		}

		if time.Now().After(nextUpdate) {
			jobs.PostUpdate(ctx, fmt.Sprintf("found %d items", len(items)))
			nextUpdate = time.Now().Add(1 * time.Second)
		}

		lastEvalKey = out.LastEvaluatedKey
		if lastEvalKey == nil {
			// We've reached the last page
			break
		}
	}

	return items, lastEvalKey, nil
}

func (p *Provider) DeleteItem(ctx context.Context, tableName string, key map[string]types.AttributeValue) error {
	_, err := p.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key:       key,
	})
	return errors.Wrap(err, "could not delete item")
}

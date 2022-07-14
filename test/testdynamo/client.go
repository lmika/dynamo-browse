package testdynamo

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

type TestData struct {
	TableName string
	Data      []map[string]interface{}
}

func SetupTestTable(t *testing.T, testData []TestData) (*dynamodb.Client, func()) {
	t.Helper()
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("ap-southeast-2"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("abc", "123", "")))
	assert.NoError(t, err)

	dynamoClient := dynamodb.NewFromConfig(cfg,
		dynamodb.WithEndpointResolver(dynamodb.EndpointResolverFromURL("http://localhost:18000")))

	for _, table := range testData {
		_, err = dynamoClient.CreateTable(ctx, &dynamodb.CreateTableInput{
			TableName: aws.String(table.TableName),
			KeySchema: []types.KeySchemaElement{
				{AttributeName: aws.String("pk"), KeyType: types.KeyTypeHash},
				{AttributeName: aws.String("sk"), KeyType: types.KeyTypeRange},
			},
			AttributeDefinitions: []types.AttributeDefinition{
				{AttributeName: aws.String("pk"), AttributeType: types.ScalarAttributeTypeS},
				{AttributeName: aws.String("sk"), AttributeType: types.ScalarAttributeTypeS},
			},
			ProvisionedThroughput: &types.ProvisionedThroughput{
				ReadCapacityUnits:  aws.Int64(100),
				WriteCapacityUnits: aws.Int64(100),
			},
		})
		assert.NoError(t, err)

		for _, item := range table.Data {
			m, err := attributevalue.MarshalMap(item)
			assert.NoError(t, err)

			_, err = dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
				TableName: aws.String(table.TableName),
				Item:      m,
			})
			assert.NoError(t, err)
		}
	}

	t.Cleanup(func() {
		for _, table := range testData {
			dynamoClient.DeleteTable(ctx, &dynamodb.DeleteTableInput{
				TableName: aws.String(table.TableName),
			})
		}
	})

	return dynamoClient, func() {}
}

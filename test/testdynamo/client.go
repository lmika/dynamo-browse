package testdynamo

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestData []map[string]interface{}

func SetupTestTable(t *testing.T, tableName string, testData TestData) *dynamodb.Client {
	t.Helper()
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("ap-southeast-2"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("abc", "123", "")))
	assert.NoError(t, err)

	dynamoClient := dynamodb.NewFromConfig(cfg,
		dynamodb.WithEndpointResolver(dynamodb.EndpointResolverFromURL("http://localhost:8000")))

	dynamoClient.DeleteTable(ctx, &dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	})

	_, err = dynamoClient.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
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

	for _, item := range testData {
		m, err := attributevalue.MarshalMap(item)
		assert.NoError(t, err)

		_, err = dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item:      m,
		})
		assert.NoError(t, err)
	}

	return dynamoClient
}

package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/lmika/awstools/internal/dynamo-browse/models"
	"github.com/lmika/awstools/internal/dynamo-browse/providers/dynamo"
	"github.com/lmika/awstools/internal/dynamo-browse/services/tables"
	"github.com/lmika/gopkgs/cli"
	"log"
)

func main() {
	ctx := context.Background()
	tableName := "business-addresses"
	totalItems := 500

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		cli.Fatalf("cannot load AWS config: %v", err)
	}

	dynamoClient := dynamodb.NewFromConfig(cfg,
		dynamodb.WithEndpointResolver(dynamodb.EndpointResolverFromURL("http://localhost:4566")))

	if _, err = dynamoClient.DeleteTable(ctx, &dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	}); err != nil {
		log.Printf("warn: cannot delete table: %v: %v", tableName, err)
	}

	if _, err = dynamoClient.CreateTable(ctx, &dynamodb.CreateTableInput{
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
	}); err != nil {
		log.Fatalf("warn: cannot create table: %v", tableName)
	}

	tableInfo := &models.TableInfo{
		Name: tableName,
		Keys: models.KeyAttribute{PartitionKey: "pk", SortKey: "sk"},
	}

	dynamoProvider := dynamo.NewProvider(dynamoClient)
	tableService := tables.NewService(dynamoProvider)

	_, _ = tableService, tableInfo

	for i := 0; i < totalItems; i++ {
		key := uuid.New().String()
		if err := tableService.Put(ctx, tableInfo, models.Item{
			"pk":           &types.AttributeValueMemberS{Value: key},
			"sk":           &types.AttributeValueMemberS{Value: key},
			"name":         &types.AttributeValueMemberS{Value: gofakeit.Name()},
			"address":      &types.AttributeValueMemberS{Value: gofakeit.Address().Address},
			"city":         &types.AttributeValueMemberS{Value: gofakeit.Address().City},
			"phone":        &types.AttributeValueMemberN{Value: gofakeit.Phone()},
			"web":          &types.AttributeValueMemberS{Value: gofakeit.URL()},
			"officeOpened": &types.AttributeValueMemberBOOL{Value: gofakeit.Bool()},
			"ratings": &types.AttributeValueMemberL{Value: []types.AttributeValue{
				&types.AttributeValueMemberN{Value: fmt.Sprint(gofakeit.IntRange(0, 5))},
				&types.AttributeValueMemberN{Value: fmt.Sprint(gofakeit.IntRange(0, 5))},
				&types.AttributeValueMemberN{Value: fmt.Sprint(gofakeit.IntRange(0, 5))},
			}},
		}); err != nil {
			log.Fatalln(err)
		}
	}

	log.Printf("table '%v' created with %v items", tableName, totalItems)
}

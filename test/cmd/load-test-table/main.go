package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"github.com/lmika/audax/internal/dynamo-browse/providers/dynamo"
	"github.com/lmika/audax/internal/dynamo-browse/services/tables"
	"github.com/lmika/gopkgs/cli"
	"github.com/pkg/errors"
	"log"
)

func main() {
	var flagSeed = flag.Int64("seed", 0, "random seed to use")
	var flagCount = flag.Int("count", 500, "number of items to produce")
	flag.Parse()

	ctx := context.Background()
	tableName := "business-addresses"
	totalItems := *flagCount

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		cli.Fatalf("cannot load AWS config: %v", err)
	}

	dynamoClient := dynamodb.NewFromConfig(cfg,
		dynamodb.WithEndpointResolver(dynamodb.EndpointResolverFromURL("http://localhost:4566")))

	// Other tables
	if err := createTable(ctx, dynamoClient, "user-accounts"); err != nil {
		log.Fatal(err)
	}

	if err := createTable(ctx, dynamoClient, "inventory"); err != nil {
		log.Fatal(err)
	}

	if err := createTable(ctx, dynamoClient, tableName); err != nil {
		log.Fatal(err)
	}

	tableInfo := &models.TableInfo{
		Name: tableName,
		Keys: models.KeyAttribute{PartitionKey: "pk", SortKey: "sk"},
	}

	dynamoProvider := dynamo.NewProvider(dynamoClient)
	tableService := tables.NewService(dynamoProvider, notROService{})

	_, _ = tableService, tableInfo

	log.Printf("using seed: %v", *flagSeed)
	gofakeit.Seed(*flagSeed)

	for i := 0; i < totalItems; i++ {
		key := gofakeit.UUID()
		if err := tableService.Put(ctx, tableInfo, models.Item{
			"pk":           &types.AttributeValueMemberS{Value: key},
			"sk":           &types.AttributeValueMemberS{Value: key},
			"name":         &types.AttributeValueMemberS{Value: gofakeit.Name()},
			"address":      &types.AttributeValueMemberS{Value: gofakeit.Address().Address},
			"city":         &types.AttributeValueMemberS{Value: gofakeit.Address().City},
			"phone":        &types.AttributeValueMemberN{Value: gofakeit.Phone()},
			"web":          &types.AttributeValueMemberS{Value: gofakeit.URL()},
			"officeOpened": &types.AttributeValueMemberBOOL{Value: gofakeit.Bool()},
			"colors": &types.AttributeValueMemberM{
				Value: map[string]types.AttributeValue{
					"door":  &types.AttributeValueMemberS{Value: gofakeit.Color()},
					"front": &types.AttributeValueMemberS{Value: gofakeit.Color()},
				},
			},
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

func createTable(ctx context.Context, dynamoClient *dynamodb.Client, tableName string) error {
	if _, err := dynamoClient.DeleteTable(ctx, &dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	}); err != nil {
		log.Printf("warn: cannot delete table: %v: %v", tableName, err)
	}

	if _, err := dynamoClient.CreateTable(ctx, &dynamodb.CreateTableInput{
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
		return errors.Wrapf(err, "cannot create table: %v", tableName)
	}
	return nil
}

type notROService struct{}

func (n notROService) DefaultLimit() int {
	return 1000
}

func (n notROService) IsReadOnly() (bool, error) {
	return false, nil
}

package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/lmika/gopkgs/cli"
)

func main() {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		cli.Fatalf("cannot load AWS config: %v", err)
	}

	ssmClient := ssm.NewFromConfig(cfg,
		ssm.WithEndpointResolver(ssm.EndpointResolverFromURL("http://localhost:4566")))

	if _, err := ssmClient.PutParameter(ctx, &ssm.PutParameterInput{
		Name:  aws.String("/alpha/bravo"),
		Type:  types.ParameterTypeString,
		Value: aws.String("This is a parameter value"),
	}); err != nil {
		cli.Fatal(err)
	}
}

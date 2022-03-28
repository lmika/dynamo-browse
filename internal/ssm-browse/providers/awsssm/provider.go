package awsssm

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/lmika/awstools/internal/ssm-browse/models"
	"github.com/pkg/errors"
	"log"
)

type Provider struct {
	client *ssm.Client
}

func NewProvider(client *ssm.Client) *Provider {
	return &Provider{
		client: client,
	}
}

func (p *Provider) List(ctx context.Context, prefix string, nextToken string) (*models.SSMParameters, error) {
	log.Printf("new prefix: %v", prefix)

	var nextTokenStr *string = nil
	if nextToken != "" {
		nextTokenStr = aws.String(nextToken)
	}
	pars, err := p.client.GetParametersByPath(ctx, &ssm.GetParametersByPathInput{
		Path:       aws.String(prefix),
		NextToken:  nextTokenStr,
		MaxResults: 10,
		Recursive:  true,
	})
	if err != nil {
		return nil, errors.Wrap(err, "cannot get parameters from path")
	}

	res := &models.SSMParameters{
		Items: make([]models.SSMParameter, len(pars.Parameters)),
		NextToken: aws.ToString(pars.NextToken),
	}
	for i, p := range pars.Parameters {
		res.Items[i] = models.SSMParameter{
			Name:  aws.ToString(p.Name),
			Value: aws.ToString(p.Value),
		}
	}

	return res, nil
}

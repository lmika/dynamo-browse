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

func (p *Provider) List(ctx context.Context, prefix string, maxCount int) (*models.SSMParameters, error) {
	log.Printf("new prefix: %v", prefix)

	pager := ssm.NewGetParametersByPathPaginator(p.client, &ssm.GetParametersByPathInput{
		Path:       aws.String(prefix),
		Recursive:  true,
		WithDecryption: true,
	})

	items := make([]models.SSMParameter, 0)
	outer: for pager.HasMorePages() {
		out, err := pager.NextPage(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get parameters from path")
		}

		for _, p := range out.Parameters {
			items = append(items, models.SSMParameter{
				Name:  aws.ToString(p.Name),
				Value: aws.ToString(p.Value),
			})
			if len(items) >= maxCount {
				break outer
			}
		}
	}

	return &models.SSMParameters{Items: items}, nil
}

package awsssm

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/lmika/awstools/internal/ssm-browse/models"
	"github.com/pkg/errors"
	"log"
)

const defaultKMSKeyIDForSecureStrings = "alias/aws/ssm"

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
		Path:           aws.String(prefix),
		Recursive:      true,
		WithDecryption: true,
	})

	items := make([]models.SSMParameter, 0)
outer:
	for pager.HasMorePages() {
		out, err := pager.NextPage(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get parameters from path")
		}

		for _, p := range out.Parameters {
			items = append(items, models.SSMParameter{
				Name:  aws.ToString(p.Name),
				Type:  p.Type,
				Value: aws.ToString(p.Value),
			})
			if len(items) >= maxCount {
				break outer
			}
		}
	}

	return &models.SSMParameters{Items: items}, nil
}

func (p *Provider) Put(ctx context.Context, param models.SSMParameter, override bool) error {
	in := &ssm.PutParameterInput{
		Name:      aws.String(param.Name),
		Type:      param.Type,
		Value:     aws.String(param.Value),
		Overwrite: override,
	}
	if param.Type == types.ParameterTypeSecureString {
		in.KeyId = aws.String(defaultKMSKeyIDForSecureStrings)
	}

	_, err := p.client.PutParameter(ctx, in)
	if err != nil {
		return errors.Wrap(err, "unable to put new SSM parameter")
	}

	return nil
}

func (p *Provider) Delete(ctx context.Context, param models.SSMParameter) error {
	_, err := p.client.DeleteParameter(ctx, &ssm.DeleteParameterInput{
		Name: aws.String(param.Name),
	})
	if err != nil {
		return errors.Wrap(err, "unable to delete SSM parameter")
	}
	return nil
}

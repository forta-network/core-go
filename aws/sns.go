package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type SNSClient interface {
	Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
	PublishBatch(ctx context.Context, params *sns.PublishBatchInput, optFns ...func(*sns.Options)) (*sns.PublishBatchOutput, error)
}

func NewSnsClient(ctx context.Context) (*sns.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, ClientOptions()...)
	if err != nil {
		return nil, err
	}

	return sns.NewFromConfig(cfg), nil
}

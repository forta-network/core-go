package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

type SESClient interface {
	SendEmail(ctx context.Context, params *ses.SendEmailInput, optFns ...func(*ses.Options)) (*ses.SendEmailOutput, error)
}

func NewSesClient(ctx context.Context) (*ses.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, ClientOptions()...)
	if err != nil {
		return nil, err
	}

	return ses.NewFromConfig(cfg), nil
}

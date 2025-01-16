//go:build integration_test
// +build integration_test

package aws

import (
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/smithy-go/logging"
	log "github.com/sirupsen/logrus"
)

type awsLogger struct{}

func (l awsLogger) Logf(classification logging.Classification, format string, v ...interface{}) {
	if classification == logging.Warn {
		log.Warnf(format, v...)
	} else {
		log.Infof(format, v...)
	}
}

func ClientOptions() []func(*config.LoadOptions) error {
	var opts []func(*config.LoadOptions) error
	opts = append(opts, config.WithRetryer(func() aws.Retryer {
		return retry.AddWithMaxAttempts(retry.NewStandard(), 5)
	}))
	if os.Getenv("AWS_DEBUG") != "" {
		opts = append(opts, config.WithLogger(awsLogger{}))
		opts = append(opts, config.WithClientLogMode(aws.LogRetries|aws.LogRequestWithBody|aws.LogResponseWithBody))

	}

	// set all credentials to test for sanity
	err := os.Setenv("AWS_ACCESS_KEY_ID", "test")
	if err != nil {
		panic("can not set integration test variables")
	}

	err = os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	if err != nil {
		panic("can not set integration test variables")
	}

	// use test credentials provider
	opts = append(opts, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")))

	// use localstack endpoint for AWS during integration tests
	var options config.LoadOptionsFunc = func(o *config.LoadOptions) error {
		o.Region = "eu-west-2"

		o.EndpointResolver = aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           "http://localstack:4566",
				SigningRegion: region,
			}, nil
		})
		return nil
	}
	opts = append(opts, options)

	return opts
}

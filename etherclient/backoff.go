package etherclient

import (
	"context"
	"net/url"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
)

const (
	backoffInitialInterval = time.Second
	backoffMaxInterval     = time.Minute
	backoffMaxElapsedTime  = time.Minute * 5
	backoffContextTimeout  = time.Minute
)

type retryOptions struct {
	MaxElapsedTime time.Duration
	MinBackoff     time.Duration
	MaxBackoff     time.Duration
}

func (ec *etherClient) withBackoff(
	ctx context.Context,
	method string,
	operation func(ctx context.Context, ethClient *ethclient.Client) error,
	options ...retryOptions,
) error {
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = backoffInitialInterval
	bo.MaxInterval = backoffMaxInterval
	bo.MaxElapsedTime = backoffMaxElapsedTime
	if options != nil {
		opts := options[0]
		if opts.MinBackoff > 0 {
			bo.InitialInterval = opts.MinBackoff
		}
		if opts.MaxBackoff > 0 {
			bo.MaxInterval = opts.MaxBackoff
		}
		if opts.MaxElapsedTime > 0 {
			bo.MaxElapsedTime = opts.MaxElapsedTime
		}
	}
	err := backoff.Retry(func() error {
		if ctx.Err() != nil {
			return backoff.Permanent(ctx.Err())
		}

		wrapper := ec.provider.Provide()
		ethClient := wrapper.Client
		tCtx, cancel := context.WithTimeout(ctx, backoffContextTimeout)
		err := operation(tCtx, ethClient)
		cancel()

		// If metrics handler is set, call with the RPC URL and the client method that was used.
		if ec.metricsHandler != nil {
			rpcUrl := wrapper.url
			u, err := url.Parse(rpcUrl)
			if err == nil {
				ec.metricsHandler(u.Host, method)
			}
		}

		if err != nil {
			// Move onto the next provider.
			ec.provider.Next()
		}
		return handleRetryErr(ctx, method, err)
	}, bo)
	if err != nil {
		logrus.WithError(err).WithField("method", method).Error("retry failed with error")
	}
	return err
}

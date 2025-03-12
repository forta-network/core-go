package etherclient

import (
	"context"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/sirupsen/logrus"
)

const (
	backoffInitialInterval = time.Second
	backoffMaxInterval     = time.Minute
	backoffMaxElapsedTime  = time.Minute * 5
	backoffContextTimeout  = time.Minute
)

func withBackoff[C Element, P Provider[C]](
	ctx context.Context,
	provider P,
	method string,
	operation func(ctx context.Context, client C) error,
) error {
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = backoffInitialInterval
	bo.MaxInterval = backoffMaxInterval
	bo.MaxElapsedTime = backoffMaxElapsedTime
	err := backoff.Retry(func() error {
		if ctx.Err() != nil {
			return backoff.Permanent(ctx.Err())
		}
		tCtx, cancel := context.WithTimeout(ctx, backoffContextTimeout)
		err := operation(tCtx, provider.Provide())
		cancel()
		if err != nil {
			// Move onto the next provider.
			provider.Next()
		}
		return handleRetryErr(ctx, method, err)
	}, bo)
	if err != nil {
		logrus.WithError(err).WithField("method", method).Error("retry failed with error")
	}
	return err
}

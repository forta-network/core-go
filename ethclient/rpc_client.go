package ethclient

import (
	"context"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/forta-network/core-go/ethclient/provider"
	"github.com/sirupsen/logrus"
)

const (
	backoffInitialInterval = time.Second
	backoffMaxInterval     = time.Minute
	backoffMaxElapsedTime  = time.Minute * 5
	backoffContextTimeout  = time.Minute
)

// any non-retriable failure errors can be listed here
var permanentErrors = []string{
	"method not found",
	"hash is not currently canonical",
	//"unknown block",
	"unable to complete request at this time",
	"503 service unavailable",
	"trace_block is not available",
	"invalid host",
	"receipt was empty",
}

func isPermanentError(err error) bool {
	if err == nil {
		return false
	}
	for _, pe := range permanentErrors {
		if strings.Contains(strings.ToLower(err.Error()), pe) {
			return true
		}
	}
	return false
}

type rpcClient struct {
	provider provider.Provider[*rpc.Client]
}

func (rc *rpcClient) Close() {
	rc.provider.Close()
}

func (rc *rpcClient) CallContext(
	ctx context.Context, result interface{}, method string, args ...interface{},
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
		rpcClient := rc.provider.Provide()
		err := rpcClient.CallContext(tCtx, result, method, args...)
		cancel()
		if err != nil {
			// Move onto the next provider.
			rc.provider.Next()
		}
		return handleRetryErr(ctx, method, err)
	}, bo)
	if err != nil {
		logrus.WithError(err).WithField("method", method).Error("retry failed with error")
	}
	return err
}

func (rc *rpcClient) BatchCallContext(ctx context.Context, b []rpc.BatchElem) error {
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = backoffInitialInterval
	bo.MaxInterval = backoffMaxInterval
	bo.MaxElapsedTime = backoffMaxElapsedTime
	err := backoff.Retry(func() error {
		if ctx.Err() != nil {
			return backoff.Permanent(ctx.Err())
		}
		tCtx, cancel := context.WithTimeout(ctx, backoffContextTimeout)
		rpcClient := rc.provider.Provide()
		err := rpcClient.BatchCallContext(tCtx, b)
		cancel()
		if err != nil {
			// Move onto the next provider.
			rc.provider.Next()
		}
		return handleRetryErr(ctx, "", err)
	}, bo)
	if err != nil {
		logrus.WithError(err).Error("retry failed with error")
	}
	return err
}

func (rc *rpcClient) EthSubscribe(ctx context.Context, channel interface{}, args ...interface{}) (sub *rpc.ClientSubscription, err error) {
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = backoffInitialInterval
	bo.MaxInterval = backoffMaxInterval
	bo.MaxElapsedTime = backoffMaxElapsedTime
	err = backoff.Retry(func() error {
		if ctx.Err() != nil {
			return backoff.Permanent(ctx.Err())
		}
		tCtx, cancel := context.WithTimeout(ctx, backoffContextTimeout)
		rpcClient := rc.provider.Provide()
		sub, err = rpcClient.Subscribe(tCtx, "eth", channel, args...)
		cancel()
		if err != nil {
			// Move onto the next provider.
			rc.provider.Next()
		}
		return handleRetryErr(ctx, "", err)
	}, bo)
	if err != nil {
		logrus.WithError(err).Error("retry failed with error")
	}
	return sub, err
}

func handleRetryErr(ctx context.Context, method string, err error) error {
	if err == nil {
		return nil
	}
	logger := logrus.NewEntry(logrus.StandardLogger())
	if len(method) > 0 {
		logger = logger.WithField("method", method)
	}
	if isPermanentError(err) {
		logger.WithError(err).Error("backoff permanent error")
		return backoff.Permanent(err)
	}
	if ctx.Err() != nil {
		logger.WithError(ctx.Err()).Error("context err")
		return backoff.Permanent(ctx.Err())
	}
	logger.WithError(err).Warn("failed...retrying")
	return err
}

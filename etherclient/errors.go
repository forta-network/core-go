package etherclient

import (
	"context"
	"strings"

	"github.com/cenkalti/backoff"
	"github.com/sirupsen/logrus"
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

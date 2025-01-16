package aws

import (
	"github.com/aws/smithy-go/logging"
	"testing"
)

func TestAwsLogger_Logf(t *testing.T) {
	l := awsLogger{}
	l.Logf(logging.Debug, "%s : %s", "key", "value")
}

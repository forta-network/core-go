package aws

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

type SignerV4 struct {
	RoundTripper http.RoundTripper
	Credentials  aws.Credentials
	Region       string
	Service      string
}

func (s *SignerV4) RoundTrip(req *http.Request) (*http.Response, error) {
	signer := v4.NewSigner()
	payloadHash, newReader, err := hashPayload(req.Body)
	if err != nil {
		return nil, err
	}
	req.Body = newReader
	err = signer.SignHTTP(context.Background(), s.Credentials, req, payloadHash, s.Service, s.Region, time.Now())
	if err != nil {
		return nil, fmt.Errorf("error signing request: %w", err)
	}
	return s.RoundTripper.RoundTrip(req)
}

func hashPayload(r io.ReadCloser) (payloadHash string, newReader io.ReadCloser, err error) {
	var payload []byte
	if r == nil {
		payload = []byte("")
	} else {
		payload, err = ioutil.ReadAll(r)
		if err != nil {
			return
		}
		newReader = ioutil.NopCloser(bytes.NewReader(payload))
	}
	hash := sha256.Sum256(payload)
	payloadHash = hex.EncodeToString(hash[:])
	return
}

func GetCredentialsFromProfile(ctx context.Context, profile string) (*aws.Credentials, error) {
	cfg, err := config.LoadSharedConfigProfile(ctx, profile)
	if err != nil {
		return nil, err
	}
	return &cfg.Credentials, nil
}

func GetCredentials(ctx context.Context) (*aws.Credentials, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRetryer(func() aws.Retryer {
		return retry.AddWithMaxAttempts(retry.NewStandard(), 5)
	}))
	if err != nil {
		return nil, err
	}

	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return nil, err
	}

	return &aws.Credentials{
		AccessKeyID:     creds.AccessKeyID,
		SecretAccessKey: creds.SecretAccessKey,
		SessionToken:    creds.SessionToken,
		CanExpire:       creds.CanExpire,
		Expires:         creds.Expires,
		Source:          creds.Source,
	}, nil

}

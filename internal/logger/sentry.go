package logger

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
)

func NewSentryClient(dsn, env, version string) (*sentry.Client, error) {
	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:         dsn,
		Release:     version,
		Environment: env,
		HTTPTransport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create new sentry client, err=%v", err)
	}

	return client, nil
}

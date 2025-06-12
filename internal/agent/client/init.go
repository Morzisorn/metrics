package agent

import (
	"time"

	"github.com/morzisorn/metrics/config"
	"resty.dev/v3"
)

// HTTPClient contains pointer to resty client, server base url and delays for retrier
type HTTPClient struct {
	BaseURL     string
	Client      *resty.Client
	retryDelays []time.Duration
}

// NewClient creates new pointer to HTTPClient based on config
func NewClient(s *config.Service) *HTTPClient {
	c := HTTPClient{
		BaseURL: s.Config.Addr,
		Client: resty.New().
			SetBaseURL(s.Config.Addr),
		retryDelays: s.Config.RetryDelays,
	}

	c.Client.SetRetryCount(len(c.retryDelays)).
		AddRetryConditions(retryConditions).
		AddRetryHooks(retryHook)

	c.Client.AddRequestMiddleware(func(client *resty.Client, req *resty.Request) error {
		err := signRequestMiddleware(req)
		if err != nil {
			return err
		}

		err = gzipMiddleware(req)
		if err != nil {
			return err
		}

		err = encryptMiddleware(req)
		if err != nil {
			return err
		}

		return nil
	})

	return &c
}

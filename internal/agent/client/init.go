package agent

import (
	"time"

	"github.com/morzisorn/metrics/config"
	"resty.dev/v3"
)

// HTTPClient contains pointer to resty client and server base url
type HTTPClient struct {
	BaseURL string
	Client  *resty.Client
}

// Delays for retrier
var retryDelays = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

// NewClient creates new pointer to HTTPClient based on config
func NewClient(s *config.Service) *HTTPClient {
	c := HTTPClient{
		BaseURL: s.Config.Addr,
		Client: resty.New().
			SetBaseURL(s.Config.Addr),
	}

	c.Client.SetRetryCount(len(retryDelays)).
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

		return nil
	})

	return &c
}

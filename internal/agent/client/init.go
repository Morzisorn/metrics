package agent

import (
	"time"

	"github.com/morzisorn/metrics/config"
	"resty.dev/v3"
)

type HTTPClient struct {
	BaseURL string
	Client  *resty.Client
}

var RetryDelays = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

func NewClient(s *config.Service) *HTTPClient {
	c := HTTPClient{
		BaseURL: s.Config.Addr,
		Client: resty.New().
			SetBaseURL(s.Config.Addr),
	}

	c.Client.SetRetryCount(len(RetryDelays)).
		AddRetryConditions(retryConditions).
		AddRetryHooks(retryHook)

	c.Client.AddRequestMiddleware(func(client *resty.Client, req *resty.Request) error {
		err := gzipMiddleware(req)
		if err != nil {
			return err
		}

		return nil
	})

	return &c
}

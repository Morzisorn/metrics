package agent

import (
	"net/http"
	"time"

	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/server/logger"
	"go.uber.org/zap"
	"resty.dev/v3"
)

type HTTPClient struct {
	BaseURL string
	Client  *resty.Client
}

func NewClient(s *config.Service) *HTTPClient {
	retryDelays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	c := HTTPClient{
		BaseURL: s.Config.Addr,
		Client: resty.New().
			SetBaseURL(s.Config.Addr),
		}

	
	c.Client.SetRetryCount(len(retryDelays)).
	AddRetryConditions(func(r *resty.Response, err error) bool {
		if err != nil {
			return true
		}

		return r.StatusCode() >= 500 || r.StatusCode() == http.StatusTooManyRequests
	}).
	AddRetryHooks(func(resp *resty.Response, err error) {
			attempt := resp.Request.Attempt
			if attempt-1 < len(retryDelays) {
				delay := retryDelays[attempt-1]
				logger.Log.Info("Request to server error", zap.Int("Retry #", attempt))
				time.Sleep(delay)
			}
		},
	)


	c.Client.AddRequestMiddleware(func(client *resty.Client, req *resty.Request) error {
		err := gzipMiddleware(req)
		if err != nil {
			return err
		}

		return nil
	})

	return &c
}

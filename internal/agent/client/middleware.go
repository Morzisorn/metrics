package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/hash"
	"github.com/morzisorn/metrics/internal/server/logger"
	"go.uber.org/zap"
	"resty.dev/v3"
)

func gzipMiddleware(r *resty.Request) error {
	body, err := getByteBody(r)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err = gz.Write(body)
	if err != nil {
		return err
	}
	gz.Close()

	r.SetBody(buf.Bytes())
	r.SetHeader("Content-Encoding", "gzip")
	return nil
}

func retryConditions(r *resty.Response, err error) bool {
	if err != nil {
		return true
	}

	return r.StatusCode() >= 500 || r.StatusCode() == http.StatusTooManyRequests
}

func retryHook(resp *resty.Response, err error) {
	attempt := resp.Request.Attempt
	if attempt-1 < len(RetryDelays) {
		delay := RetryDelays[attempt-1]
		logger.Log.Info("Request to server error", zap.Int("Retry #", attempt))
		time.Sleep(delay)
	}
}

func signRequestMiddleware(r *resty.Request) error {
	service := config.GetService()
	if service.Config.Key == "" {
		return nil
	}

	body, err := getByteBody(r)
	if err != nil {
		return err
	}

	hash := hash.GetHash(body)
	hashHex := hex.EncodeToString(hash[:])

	r.SetHeader("HashSHA256", hashHex)

	return nil
}

func getByteBody(r *resty.Request) ([]byte, error) {
	body := r.Body
	if body == nil {
		return []byte{}, nil
	}

	var jsonBody []byte
	var err error

	switch b := body.(type) {
	case []byte:
		jsonBody = b
	case string:
		jsonBody = []byte(b)
	default:
		jsonBody, err = json.Marshal(body)
		if err != nil {
			return []byte{}, err
		}
	}
	return jsonBody, nil
}

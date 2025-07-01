package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
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
	cfg := config.GetService()
	if attempt-1 < len(cfg.Config.RetryDelays) {
		delay := cfg.Config.RetryDelays[attempt-1]
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

func encryptMiddleware(r *resty.Request) error {
	service := config.GetService()
	if service.Config.PublicKey == nil {
		return nil
	}

	body, err := getByteBody(r)
	if err != nil {
		return err
	}

	aesKey := make([]byte, 32)
	if _, err := rand.Read(aesKey); err != nil {
		return err
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return err
	}
	ciphertext := gcm.Seal(nil, nonce, body, nil)

	encKeyBytes, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, service.Config.PublicKey, aesKey, nil)
	if err != nil {
		return err
	}

	payload := struct {
		Key  string `json:"key"`
		Data string `json:"data"`
	}{
		Key:  base64.StdEncoding.EncodeToString(encKeyBytes),
		Data: base64.StdEncoding.EncodeToString(append(nonce, ciphertext...)),
	}

	wrapped, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	r.SetBody(wrapped)
	r.SetHeader("X-Encrypted", "1")

	return nil 
}


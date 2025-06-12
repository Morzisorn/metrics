package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/morzisorn/metrics/internal/server/logger"
	"go.uber.org/zap"
)

func loadEnvFile(envPath string) error {
	return godotenv.Load(envPath)
}

func (c *Config) parseEnv(app string) error {
	switch app {
	case "agent":
		c.parseAgentEnvs()
	case "server":
		c.parseServerEnvs()
	}

	return nil
}

func (c *Config) parseAgentEnvs() {
	c.parseAgentFlags()

	addr := os.Getenv("ADDRESS")
	if addr != "" {
		c.Addr = addr
	}

	f, err := getEnvFloat("POLL_INTERVAL")
	if err == nil && f != 0 {
		c.PollInterval = f
	}

	f, err = getEnvFloat("REPORT_INTERVAL")
	if err == nil && f != 0 {
		c.ReportInterval = f
	}

	k, err := getEnvString("KEY")
	if err == nil {
		c.Key = k
	}

	l, err := getEnvInt("RATE_LIMIT")
	if err == nil {
		c.RateLimit = l
	}

	rd, err := getEnvString("RETRY_DELAYS")
	if err == nil {
		c.RetryDelays, err = parseRetryDelays(rd)
		if err != nil {
			logger.Log.Panic("Parse env error ", zap.Error(err))
		}
	}

	cryptoPath, err := getEnvString("CRYPTO_KEY")
	if err == nil {
		c.CryptoKeyPath = cryptoPath
		pemData, err := getStringFromFile(c.CryptoKeyPath)
		if err != nil {
			logger.Log.Panic("get crypto key from file error ", zap.Error(err))
		}
		c.PublicKey, err = getKeyFromPem[*rsa.PublicKey](pemData)
		if err != nil {
			logger.Log.Panic("parse public key error ", zap.Error(err))
		}
	}
}

func (c *Config) parseServerEnvs() {
	err := c.parseServerFlags()
	if err != nil {
		logger.Log.Panic("Parse flags error ", zap.Error(err))
	}

	addr := os.Getenv("ADDRESS")
	if addr != "" {
		c.Addr = addr
	}

	i, err := getEnvInt("STORE_INTERVAL")
	if err == nil {
		c.StoreInterval = i
	}

	s, err := getEnvString("FILE_STORAGE_PATH")
	if err == nil && s != "" {
		c.FileStoragePath = s
	}

	b, err := getEnvBool("RESTORE")
	if err == nil {
		c.Restore = b
	}

	d, err := getEnvString("DATABASE_DSN")
	if err == nil {
		c.DBConnStr = d
	}

	k, err := getEnvString("KEY")
	if err == nil {
		c.Key = k
	}

	cryptoPath, err := getEnvString("CRYPTO_KEY")
	if err == nil {
		c.CryptoKeyPath = cryptoPath
		pemData, err := getStringFromFile(c.CryptoKeyPath)
		if err != nil {
			logger.Log.Panic("get crypto key from file error ", zap.Error(err))
		}
		c.PrivateKey, err = getKeyFromPem[*rsa.PrivateKey](pemData)
		if err != nil {
			logger.Log.Panic("parse private key error ", zap.Error(err))
		}
	}
}

func getEnvFloat(key string) (float64, error) {
	env := os.Getenv(key)
	if env != "" {
		return strconv.ParseFloat(env, 64)
	}
	return 0, fmt.Errorf("env %s not found", key)
}

func getEnvInt(key string) (int64, error) {
	env := os.Getenv(key)
	if env != "" {
		return strconv.ParseInt(env, 10, 64)
	}
	return 0, fmt.Errorf("env %s not found", key)
}

func getEnvString(key string) (string, error) {
	env := os.Getenv(key)
	if env != "" {
		return env, nil
	}
	return "", fmt.Errorf("env %s not found", key)
}

func getEnvBool(key string) (bool, error) {
	env := os.Getenv(key)
	if env != "" {
		return strconv.ParseBool(env)
	}
	return false, fmt.Errorf("env %s not found", key)
}

func parseRetryDelays(s string) ([]time.Duration, error) {
	splited := strings.Split(s, ",")
	res := make([]time.Duration, len(splited))
	for i, v := range splited {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("env retry delays have to contain integers only")
		}
		res[i] = time.Duration(n) * time.Second
	}
	return res, nil
}

func getKeyFromPem[T *rsa.PublicKey | *rsa.PrivateKey](pemData string) (T, error) {
	var zero T
	rest := []byte(pemData)
	for {
		block, remaining := pem.Decode(rest)
		if block == nil {
			return nil, fmt.Errorf("no PUBLIC KEY block found")
		}

		if _, isPub := any(zero).(*rsa.PublicKey); isPub && block.Type == "PUBLIC KEY" {
			pubIfc, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err != nil {
				return nil, err
			}
			pub, ok := pubIfc.(T)
			if !ok {
				return nil, fmt.Errorf("not RSA public key")
			}
			return pub, nil
		}

		if _, isPriv := any(zero).(*rsa.PrivateKey); isPriv && block.Type == "PRIVATE KEY" {
			privIfc, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return nil, err
			}
			priv, ok := privIfc.(T)
			if !ok {
				return nil, fmt.Errorf("not RSA private key")
			}
			return priv, nil
		}

		rest = remaining
	}
}

func getStringFromFile(path string) (string, error) {
	root, err := GetProjectRoot()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(filepath.Join(root, path))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

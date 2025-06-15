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

	var configMap map[string]interface{}

	configFile, err := getEnvString("CONFIG")
	if err == nil && configFile != "" {
		c.ConfigFile = configFile
		configMap, err = getConfigMap(configFile)
		if err != nil {
			logger.Log.Panic("parse json config error ", zap.Error(err))
		}
	}

	addr := os.Getenv("ADDRESS")
	if addr != "" {
		c.Addr = addr
	} else if c.Addr == "" && configMap != nil && configMap["address"].(string) != "" {
		c.Addr = configMap["address"].(string)
	}

	f, err := getEnvFloat("POLL_INTERVAL")
	if err == nil && f != 0 {
		c.PollInterval = f
	} else if c.PollInterval == float64(0) && configMap != nil && configMap["poll_interval"].(float64) != 0 {
		c.PollInterval = configMap["poll_interval"].(float64)
	}

	f, err = getEnvFloat("REPORT_INTERVAL")
	if err == nil && f != 0 {
		c.ReportInterval = f
	} else if c.ReportInterval == float64(0) && configMap != nil && configMap["report_interval"].(float64) != 0 {
		c.ReportInterval = configMap["report_interval"].(float64)
	}

	k, err := getEnvString("KEY")
	if err == nil {
		c.Key = k
	} else if c.Key == "" && configMap != nil && configMap["key"].(string) != "" {
		c.Key = configMap["key"].(string)
	}

	l, err := getEnvInt("RATE_LIMIT")
	if err == nil {
		c.RateLimit = l
	} else if c.RateLimit == 0 && configMap != nil && configMap["rate_limit"].(int64) != 0 {
		c.RateLimit = configMap["rate_limit"].(int64)
	}

	rd, err := getEnvString("RETRY_DELAYS")
	if err == nil {
		c.RetryDelays, err = parseRetryDelays(rd)
		if err != nil {
			logger.Log.Panic("Parse env error ", zap.Error(err))
		}
	} else if c.RetryDelays == nil && configMap != nil && configMap["retry_delays"] != nil {
		for _, r := range configMap["retry_delays"].([]interface{}) {
			v, ok := r.(float64)
			if !ok {
				logger.Log.Panic("parse retry delays error ", zap.Error(err))
			}
			c.RetryDelays = append(c.RetryDelays, time.Duration(v)*time.Second)
		}

	}

	cryptoPath, err := getEnvString("CRYPTO_KEY")
	if err == nil {
		c.CryptoKeyPath = cryptoPath
	} else if c.CryptoKeyPath == "" && configMap != nil && configMap["crypto_key"].(string) != "" {
		c.CryptoKeyPath = configMap["crypto_key"].(string)
	}

	if c.CryptoKeyPath != "" {
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
		logger.Log.Panic("parse flags error ", zap.Error(err))
	}

	var configMap map[string]interface{}

	configFile, err := getEnvString("CONFIG")
	if err == nil && configFile != "" {
		c.ConfigFile = configFile
		configMap, err = getConfigMap(configFile)
		if err != nil {
			logger.Log.Panic("parse json config error ", zap.Error(err))
		}
	}

	addr := os.Getenv("ADDRESS")
	if addr != "" {
		c.Addr = addr
	} else if c.Addr == "" && configMap != nil && configMap["address"].(string) != "" {
		c.Addr = configMap["address"].(string)
	}

	i, err := getEnvInt("STORE_INTERVAL")
	if err == nil {
		c.StoreInterval = i
	} else if c.StoreInterval == 0 && configMap != nil && configMap["store_interval"].(int64) != 0 {
		c.StoreInterval = configMap["store_interval"].(int64)
	}

	s, err := getEnvString("FILE_STORAGE_PATH")
	if err == nil && s != "" {
		c.FileStoragePath = s
	} else if c.FileStoragePath == "" && configMap != nil && configMap["file_storage_path"].(string) != "" {
		c.FileStoragePath = configMap["file_storage_path"].(string)
	}

	b, err := getEnvBool("RESTORE")
	if err == nil {
		c.Restore = b
	} else if !c.Restore && configMap != nil && configMap["restore"].(bool) {
		c.Restore = configMap["restore"].(bool)
	}

	d, err := getEnvString("DATABASE_DSN")
	if err == nil {
		c.DBConnStr = d
	} else if c.DBConnStr == "" && configMap != nil && configMap["database_dsn"].(string) != "" {
		c.DBConnStr = configMap["database_dsn"].(string)
	}

	k, err := getEnvString("KEY")
	if err == nil {
		c.Key = k
	} else if c.Key == "" && configMap != nil && configMap["key"].(string) != "" {
		c.Key = configMap["key"].(string)
	}

	cryptoPath, err := getEnvString("CRYPTO_KEY")
	if err == nil {
		c.CryptoKeyPath = cryptoPath
	} else if c.CryptoKeyPath == "" && configMap != nil && configMap["crypto_key"].(string) != "" {
		c.CryptoKeyPath = configMap["crypto_key"].(string)
	}

	if c.CryptoKeyPath != "" {
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

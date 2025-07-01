package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/morzisorn/metrics/internal/server/logger"
	"go.uber.org/zap"
)

func parseJSONConfig(configPath string) (map[string]interface{}, error) {
	j, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("open json config error: %v. path: %s", err, configPath)
	}

	var config map[string]interface{}

	err = json.Unmarshal(j, &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func getConfigMap(c *Config) (map[string]interface{}, error) {
	configFile, err := getEnvString("CONFIG")
	if err == nil && configFile != "" {
		c.ConfigFile = configFile
		root, err := GetProjectRoot()
		if err != nil {
			logger.Log.Panic("get project root error ", zap.Error(err))
		}
		configPath := filepath.Join(root, "config", configFile)

		configMap, err := parseJSONConfig(configPath)
		if err != nil {
			logger.Log.Panic("parse json config error ", zap.Error(err))
		}
		return configMap, nil
	}
	return nil, nil
}

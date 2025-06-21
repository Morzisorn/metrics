package config

import (
	"encoding/json"
	"os"
)

func getConfigMap(configPath string) (map[string]interface{}, error) {
	j, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config map[string]interface{}

	err = json.Unmarshal(j, &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

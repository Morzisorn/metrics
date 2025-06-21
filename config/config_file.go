package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func getConfigMap(file string) (map[string]interface{}, error) {
	root, err := GetProjectRoot()
	if err != nil {
		return nil, err
	}

	j, err := os.ReadFile(filepath.Join(root, "config", file))
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

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetConfigMap(t *testing.T) {
	_, err := getConfigMap("test_config.json")
	require.NoError(t, err)
}

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseJSONConfig(t *testing.T) {
	_, err := parseJSONConfig("test_config.json")
	require.NoError(t, err)
}

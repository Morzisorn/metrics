package hash

import (
	"crypto/sha256"
	"testing"

	"github.com/morzisorn/metrics/config"
	"github.com/stretchr/testify/assert"
)

func TestGetHash(t *testing.T) {
	service := config.GetService("server")
	body := []byte("test string")

	str := append(body, []byte(service.Config.Key)...)
	expected := sha256.Sum256(str)

	assert.Equal(t,expected, GetHash(body))
}

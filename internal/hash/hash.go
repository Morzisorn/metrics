// Package hash is used to generate hash
package hash

import (
	"crypto/sha256"

	"github.com/morzisorn/metrics/config"
)

// GetHash receives slice of bytes and returns 32 bytes array.
// Uses sha256 algorithm.
func GetHash(body []byte) [32]byte {
	service := config.GetService()
	str := append(body, []byte(service.Config.Key)...)
	return sha256.Sum256(str)
}

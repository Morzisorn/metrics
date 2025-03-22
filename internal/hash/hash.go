package hash

import (
	"crypto/sha256"

	"github.com/morzisorn/metrics/config"
)

func GetHash(body []byte) [32]byte {
	service := config.GetService()
	str := append(body, []byte(service.Config.Key)...)
	return sha256.Sum256(str)
}

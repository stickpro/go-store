package hash

import (
	"crypto/sha256"
	"encoding/hex"
)

func SHA256(plainTextToken string) string {
	hash := sha256.New()
	hash.Write([]byte(plainTextToken))
	hashedBytes := hash.Sum(nil)
	return hex.EncodeToString(hashedBytes)
}

func SHA256Signature(data []byte, secretKey string) string {
	sign := sha256.New()
	sign.Write(append(data, []byte(secretKey)...))
	return hex.EncodeToString(sign.Sum(nil))
}

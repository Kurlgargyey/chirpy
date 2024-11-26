package auth

import (
	"crypto/rand"
	"encoding/hex"
)

func MakeRefreshToken() (string, error) {
	token_bytes := make([]byte, 32)
	_, err := rand.Read(token_bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(token_bytes), nil
}

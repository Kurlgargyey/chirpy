package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("HTTP Headers did not contain any auth info: %s", headers)
	}
	return strings.TrimSpace(strings.TrimPrefix(authHeader, "ApiKey")), nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("HTTP Headers did not contain any auth info: %s", headers)
	}
	return strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer")), nil
}

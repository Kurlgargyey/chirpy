package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(pwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), 0)
	if err != nil {
		return "", fmt.Errorf("error hashing password: %w", err)
	}
	return string(hash), nil
}

func CheckPasswordHash(pwd, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd))
}

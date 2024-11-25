package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	password := "Start2020"
	hash, _ := HashPassword(password)
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		t.Fatalf("TestHashPassword failed: %s", err)
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "Start2020"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), 0)
	if err := CheckPasswordHash(password, string(hash)); err != nil {
		t.Fatalf("TestCheckPasswordHash failed: %s", err)
	}
}

func TestTokenBasic(t *testing.T) {
	tokenSecret := "lollmao"
	userID := uuid.New()
	expiresIn := time.Duration.Minutes(2)

	token, token_err := MakeJWT(userID, tokenSecret, time.Duration(expiresIn))
	tokenID, validation_err := ValidateJWT(token, tokenSecret)
	if token_err != nil || validation_err != nil || userID != tokenID {
		t.Fatalf("TestTokenBasic failed.\nToken error: %s\nValidation error: %s\nuserID: %s\ntokenID:%s", token_err, validation_err, userID, tokenID)
	}
}

func TestExpiredToken(t *testing.T) {
	tokenSecret := "lollmao"
	userID := uuid.New()
	expiresIn := time.Duration.Seconds(0)

	token, token_err := MakeJWT(userID, tokenSecret, time.Duration(expiresIn))
	time.Sleep(time.Second)
	tokenID, validation_err := ValidateJWT(token, tokenSecret)
	if validation_err == nil || token_err != nil || tokenID != uuid.Nil {
		t.Fatalf("TestExpiredToken failed.\nToken error: %snValidation error: %s\nuserID: %s\ntokenID:%s", token_err, validation_err, userID, tokenID)
	}
}

func TestWrongSecret(t *testing.T) {
	tokenSecret := "lollmao"
	userID := uuid.New()
	expiresIn := time.Minute * 5

	token, token_err := MakeJWT(userID, tokenSecret, expiresIn)
	tokenID, validation_err := ValidateJWT(token, "lol")
	if validation_err == nil || token_err != nil || tokenID != uuid.Nil {
		t.Fatalf("TestWrongSecret failed.\nToken error: %snValidation error: %s\nuserID: %s\ntokenID:%s", token_err, validation_err, userID, tokenID)
	}
}

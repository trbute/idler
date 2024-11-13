package auth

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"net/http"
	"testing"
	"time"
)

func TestMakeAndValidateJWT(t *testing.T) {
	id := uuid.New()

	tokenSecret := "testing123"
	tokenString, err := MakeJWT(id, tokenSecret, time.Duration(72*time.Hour))
	if err != nil {
		t.Fatalf("Failed to make secret")
	}

	validated, err := ValidateJWT(tokenString, tokenSecret)
	if err != nil {
		t.Fatalf("Failed to validate JWT")
	}

	if id != validated {
		t.Fatalf("Secrets do not match")
	}
}

func TestExpiredJWT(t *testing.T) {
	id := uuid.New()
	tokenSecret := "testing123"
	tokenString, err := MakeJWT(id, tokenSecret, time.Duration(1*time.Millisecond))
	if err != nil {
		t.Fatalf("Failed to make JWT")
	}

	time.Sleep(1 * time.Millisecond)

	_, err = ValidateJWT(tokenString, tokenSecret)
	if !errors.Is(err, jwt.ErrTokenExpired) {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestInvalidSecret(t *testing.T) {
	id := uuid.New()

	tokenSecret := "testing123"
	tokenString, err := MakeJWT(id, tokenSecret, time.Duration(72*time.Hour))
	if err != nil {
		t.Fatalf("Failed to make secret")
	}

	tokenSecret = "testing1234"
	_, err = ValidateJWT(tokenString, tokenSecret)
	if !errors.Is(err, jwt.ErrTokenSignatureInvalid) {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestValidHeaders(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer Hello")
	_, err := GetBearerToken(headers)
	if err != nil {
		t.Fatalf("%v", err)
	}
}

func TestInvalidHeaders(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Boarer Hello")
	_, err := GetBearerToken(headers)
	if err == nil {
		t.Fatalf("Expected error to occur")
	}
}

func TestNoAuthHeader(t *testing.T) {
	headers := http.Header{}
	_, err := GetBearerToken(headers)
	if err == nil {
		t.Fatalf("Expected error to occur")
	}
}

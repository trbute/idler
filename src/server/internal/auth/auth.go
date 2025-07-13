package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

const (
	tokenIssuer       = "idler"
	blacklistKeyFmt   = "blacklist_token:%s"
	userTokensKeyFmt  = "user_tokens:%s"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func CheckPasswordHash(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return err
	}

	return nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	now := time.Now()
	jwtID := uuid.New().String() // Unique ID for this token
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    tokenIssuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
		Subject:   userID.String(),
		ID:        jwtID,
	})

	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) { return []byte(tokenSecret), nil },
	)
	if err != nil {
		return uuid.Nil, err
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, err
	}

	if issuer != tokenIssuer {
		return uuid.Nil, errors.New("Invalid token issuer")
	}

	idStr, err := uuid.Parse(subject)
	if err != nil {
		return uuid.Nil, err
	}

	return idStr, nil
}

func ValidateJWTWithBlacklist(ctx context.Context, tokenString, tokenSecret string, redis *redis.Client) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) { return []byte(tokenSecret), nil },
	)
	if err != nil {
		return uuid.Nil, err
	}

	claims := token.Claims.(*jwt.RegisteredClaims)
	
	// Check if this specific token is blacklisted
	tokenHash := fmt.Sprintf(blacklistKeyFmt, claims.ID)
	blacklisted, err := redis.Exists(ctx, tokenHash).Result()
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to check token blacklist: %w", err)
	}
	
	if blacklisted > 0 {
		return uuid.Nil, errors.New("token revoked")
	}

	subject, err := claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	issuer, err := claims.GetIssuer()
	if err != nil {
		return uuid.Nil, err
	}

	if issuer != tokenIssuer {
		return uuid.Nil, errors.New("Invalid token issuer")
	}

	userID, err := uuid.Parse(subject)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

func BlacklistToken(ctx context.Context, tokenString string, redis *redis.Client) error {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) { return nil, nil }, // Don't validate signature, just parse
	)
	if err != nil {
		return err
	}

	claims := token.Claims.(*jwt.RegisteredClaims)
	if claims.ID == "" {
		return errors.New("token missing ID claim")
	}

	tokenHash := fmt.Sprintf(blacklistKeyFmt, claims.ID)
	expiration := claims.ExpiresAt.Time.Sub(time.Now())
	if expiration <= 0 {
		return nil // Token already expired, no need to blacklist
	}

	return redis.Set(ctx, tokenHash, "1", expiration).Err()
}

func BlacklistAllUserTokens(ctx context.Context, userID uuid.UUID, redis *redis.Client) ([]string, error) {
	userTokensKey := fmt.Sprintf(userTokensKeyFmt, userID.String())
	tokenIDs, err := redis.SMembers(ctx, userTokensKey).Result()
	if err != nil {
		return nil, err
	}

	pipe := redis.Pipeline()
	for _, tokenID := range tokenIDs {
		blacklistKey := fmt.Sprintf(blacklistKeyFmt, tokenID)
		pipe.Set(ctx, blacklistKey, "1", time.Hour)
	}
	
	pipe.Del(ctx, userTokensKey)
	
	_, err = pipe.Exec(ctx)
	return tokenIDs, err
}

func TrackUserToken(ctx context.Context, userID uuid.UUID, tokenID string, redis *redis.Client) error {
	userTokensKey := fmt.Sprintf(userTokensKeyFmt, userID.String())
	return redis.SAdd(ctx, userTokensKey, tokenID).Err()
}

func ValidateUserSession(ctx context.Context, userID uuid.UUID, redis *redis.Client) error {
	userTokensKey := fmt.Sprintf(userTokensKeyFmt, userID.String())
	tokenIDs, err := redis.SMembers(ctx, userTokensKey).Result()
	if err != nil {
		return fmt.Errorf("failed to check user tokens: %w", err)
	}

	if len(tokenIDs) == 0 {
		return fmt.Errorf("no active session")
	}

	for _, tokenID := range tokenIDs {
		if ValidateSpecificToken(ctx, tokenID, redis) == nil {
			return nil
		}
	}

	return fmt.Errorf("session revoked")
}

func ValidateSpecificToken(ctx context.Context, tokenID string, redis *redis.Client) error {
	blacklistKey := fmt.Sprintf(blacklistKeyFmt, tokenID)
	blacklisted, err := redis.Exists(ctx, blacklistKey).Result()
	if err != nil {
		return fmt.Errorf("failed to check token blacklist: %w", err)
	}
	
	if blacklisted > 0 {
		return fmt.Errorf("token blacklisted")
	}
	
	return nil
}

func GetBearerToken(headers http.Header) (string, error) {
	auth := headers.Get("Authorization")
	if auth == "" {
		return "", errors.New("Authorization header doesn't exist")
	}

	split := strings.Split(auth, " ")
	if len(split) != 2 || split[0] != "Bearer" {
		return "", fmt.Errorf("Invalid, Authorization header: %v", auth)
	}

	return split[1], nil
}

func GetAPIKey(headers http.Header) (string, error) {
	auth := headers.Get("Authorization")
	if auth == "" {
		return "", errors.New("Authorization header doesn't exist")
	}

	split := strings.Split(auth, " ")
	if len(split) != 2 || split[0] != "ApiKey" {
		return "", fmt.Errorf("Invalid, Authorization header: %v", auth)
	}

	return split[1], nil
}

func MakeRefreshToken() (string, error) {
	c := 32
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	encodedStr := hex.EncodeToString(b)

	return encodedStr, nil
}

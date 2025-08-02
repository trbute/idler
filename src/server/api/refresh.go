package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trbute/idler/server/internal/auth"
	"github.com/trbute/idler/server/internal/database"
)

type RefreshResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func (cfg *ApiConfig) handleRefresh(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Failed retrieving token", err)
		return
	}

	tokenRecord, err := cfg.DB.GetRefreshTokenById(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to retrieve token from database", err)
		return
	}

	if tokenRecord.ExpiresAt.Time.Before(time.Now()) || tokenRecord.RevokedAt.Valid {
		respondWithError(
			w,
			http.StatusUnauthorized,
			"Refresh token has expired",
			errors.New("refresh token has expired"),
		)
		return
	}

	expiresInSeconds := 3600
	expireDuration := time.Duration(time.Duration(expiresInSeconds) * time.Second)

	userID := uuid.UUID(tokenRecord.UserID.Bytes)
	
	// Revoke the current refresh token
	err = cfg.DB.RevokeRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to revoke current refresh token", err)
		return
	}

	// Create new JWT
	newJWT, err := auth.MakeJWT(userID, cfg.JwtSecret, expireDuration)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "JWT creation failed", err)
		return
	}

	// Create new refresh token  
	newRefreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Refresh token creation failed", err)
		return
	}

	// Store new refresh token
	day := 24 * time.Hour
	refreshExpire := time.Now().Add(60 * day)
	pgTimestamp := pgtype.Timestamp{
		Time:  refreshExpire,
		Valid: true,
	}

	_, err = cfg.DB.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     newRefreshToken,
		UserID:    tokenRecord.UserID,
		ExpiresAt: pgTimestamp,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Refresh token db insert failed", err)
		return
	}

	token = newJWT

	parsedToken, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JwtSecret), nil
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to parse created token", err)
		return
	}

	claims, ok := parsedToken.Claims.(*jwt.RegisteredClaims)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Invalid token claims type", nil)
		return
	}
	err = auth.TrackUserToken(r.Context(), userID, claims.ID, cfg.Redis)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to track new token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, RefreshResponse{
		Token:        token,
		RefreshToken: newRefreshToken,
	})
}

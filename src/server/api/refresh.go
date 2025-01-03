package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/trbute/idler/server/internal/auth"
)

type RefreshResponse struct {
	Token string `json:"token"`
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
			errors.New("Refresh token has expired"),
		)
		return
	}

	expiresInSeconds := 3600
	expireDuration := time.Duration(time.Duration(expiresInSeconds) * time.Second)

	userID := uuid.UUID(tokenRecord.UserID.Bytes)
	token, err = auth.MakeJWT(userID, cfg.JwtSecret, expireDuration)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "JWT creation failed", err)
		return
	}

	respondWithJSON(w, http.StatusOK, RefreshResponse{
		Token: token,
	})
}

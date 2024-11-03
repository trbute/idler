package main

import (
	"errors"
	"github.com/trbute/idler/internal/auth"
	"net/http"
	"time"
)

type RefreshResponse struct {
	Token string `json:"token"`
}

func (cfg *apiConfig) handleRefresh(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Failed retrieving token", err)
		return
	}

	tokenRecord, err := cfg.db.GetRefreshTokenById(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to retrieve token from database", err)
		return
	}

	if tokenRecord.ExpiresAt.Before(time.Now()) || tokenRecord.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Refresh token has expired", errors.New("Refresh token has expired"))
		return
	}

	expiresInSeconds := 3600
	expireDuration := time.Duration(time.Duration(expiresInSeconds) * time.Second)

	token, err = auth.MakeJWT(tokenRecord.UserID, cfg.jwtSecret, expireDuration)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "JWT creation failed", err)
		return
	}

	respondWithJSON(w, http.StatusOK, RefreshResponse{
		Token: token,
	})
}

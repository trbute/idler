package main

import (
	"github.com/trbute/idler/internal/auth"
	"net/http"
)

func (cfg *apiConfig) handleRevoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Failed retrieving token", err)
		return
	}

	tokenRecord, err := cfg.db.GetRefreshTokenById(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve token from database", err)
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), tokenRecord.Token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to update refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

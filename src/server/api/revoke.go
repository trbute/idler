package api

import (
	"net/http"

	"github.com/trbute/idler/server/internal/auth"
)

func (cfg *ApiConfig) handleRevoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Failed retrieving token", err)
		return
	}

	tokenRecord, err := cfg.DB.GetRefreshTokenById(r.Context(), token)
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"Unable to retrieve token from database",
			err,
		)
		return
	}

	err = cfg.DB.RevokeRefreshToken(r.Context(), tokenRecord.Token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to update refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

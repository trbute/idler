package api

import (
	"net/http"

	"github.com/trbute/idler/server/internal/auth"
)

type Action struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

func (cfg *ApiConfig) handleGetActions(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to retrieve token", err)
		return
	}

	_, err = auth.ValidateJWT(token, cfg.JwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token invalid", err)
		return
	}

	actions, err := cfg.DB.GetAllActions(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve actions", err)
		return
	}

	var actionResponse []Action
	for _, action := range actions {
		actionResponse = append(actionResponse, Action{
			ID:   action.ID,
			Name: action.Name,
		})
	}

	respondWithJSON(w, http.StatusOK, actionResponse)
}

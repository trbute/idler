package main

import (
	"net/http"
)

type Action struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

func (cfg *apiConfig) handleGetActions(w http.ResponseWriter, r *http.Request) {
	actions, err := cfg.db.GetAllActions(r.Context())
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

	respondWithJSON(w, http.StatusCreated, actionResponse)
}

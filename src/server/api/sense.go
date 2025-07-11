package api

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trbute/idler/server/internal/auth"
	"github.com/trbute/idler/server/internal/database"
)

type charData struct {
	CharacterName string `json:"character_name"`
	ActionName    string `json:"action_name"`
	ActionTarget  string `json:"action_target"`
}

type area struct {
	Characters    []charData `json:"characters"`
	ResourceNodes []string   `json:"resource_nodes"`
}

func (cfg *ApiConfig) handleGetArea(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to retrieve token", err)
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.JwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token invalid", err)
		return
	}

	charName := r.PathValue("character")
	char, err := cfg.DB.GetCharacterByName(r.Context(), charName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve character", err)
		return
	}

	pgUserId := pgtype.UUID{
		Bytes: userId,
		Valid: true,
	}

	if char.UserID != pgUserId {
		respondWithError(w, http.StatusUnauthorized, "Character doesn't belong to user", err)
		return
	}

	characters, err := cfg.DB.GetCharactersByCoordinates(r.Context(), database.GetCharactersByCoordinatesParams{
		PositionX: char.PositionX,
		PositionY: char.PositionY,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve characters in area", err)
		return
	}

	resourceNodes, err := cfg.DB.GetResourceNodeSpawnsByCoordinates(r.Context(), database.GetResourceNodeSpawnsByCoordinatesParams{
		PositionX: char.PositionX,
		PositionY: char.PositionY,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve resource nodes in area", err)
		return
	}
	chars := []charData{}
	for _, c := range characters {
		action, err := cfg.DB.GetActionById(r.Context(), c.ActionID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Unable to retrieve action name", err)
			return
		}
		chars = append(chars, charData{
			CharacterName: c.Name,
			ActionName:    action.Name,
			ActionTarget:  string(c.ActionTarget),
		})
	}
	area := area{
		Characters:    characters,
		ResourceNodes: make([]string, 0, len(resourceNodes)),
	}

	respondWithJSON(w, http.StatusOK, area)
}

package api

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trbute/idler/server/internal/auth"
	"github.com/trbute/idler/server/internal/world"
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

	key := world.Coord{
		PositionX: char.PositionX,
		PositionY: char.PositionY,
	}

	cell := cfg.World.Grid[key]
	area := area{}
	area.Characters = []charData{}

	for key, value := range cell.Characters {
		if key != charName {
			action, err := cfg.DB.GetActionById(r.Context(), value.ActionId)
			if err != nil {
				respondWithError(
					w,
					http.StatusInternalServerError,
					"Unable to retrieve action",
					err,
				)
				return
			}

			char := charData{
				CharacterName: key,
				ActionName:    action.Name,
				ActionTarget:  value.ActionTarget,
			}
			area.Characters = append(area.Characters, char)
		}
	}

	for key := range cell.ResourceNodes {
		area.ResourceNodes = append(area.ResourceNodes, key)
	}

	respondWithJSON(w, http.StatusOK, area)
}

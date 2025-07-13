package api

import (
	"net/http"

	"github.com/trbute/idler/server/internal/auth"
	"github.com/trbute/idler/server/internal/database"
)

type charData struct {
	CharacterName string `json:"character_name"`
	ActionName    string `json:"action_name"`
	ActionTarget  string `json:"action_target"`
}

type area struct {
	PositionX     int32      `json:"position_x"`
	PositionY     int32      `json:"position_y"`
	Characters    []charData `json:"characters"`
	ResourceNodes []string   `json:"resource_nodes"`
}

func (cfg *ApiConfig) handleGetArea(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to retrieve token", err)
		return
	}

	userId, err := auth.ValidateJWTWithBlacklist(r.Context(), token, cfg.JwtSecret, cfg.Redis)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token invalid", err)
		return
	}

	charName := r.PathValue("character")
	
	char, err := cfg.GetCharacterWithOwnershipValidation(r.Context(), charName, userId)
	if err != nil {
		if err.Error() == "character doesn't belong to user" {
			respondWithError(w, http.StatusUnauthorized, "Character doesn't belong to user", nil)
		} else {
			respondWithError(w, http.StatusInternalServerError, "Unable to retrieve character", err)
		}
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

	resourceNodes, err := cfg.GetResourceNodeSpawnsByCoordinates(r.Context(), char.PositionX, char.PositionY)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve resource nodes in area", err)
		return
	}

	chars := []charData{}
	for _, c := range characters {
		action, err := cfg.GetActionById(r.Context(), c.ActionID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Unable to retrieve action name", err)
			return
		}
		actionTarget := ""
		if c.ActionTarget.Valid {
			spawn, err := cfg.DB.GetResourceNodeSpawnById(r.Context(), c.ActionTarget.Int32)
			if err == nil {
				targetNode, err := cfg.GetResourceNodeById(r.Context(), spawn.NodeID)
				if err == nil {
					actionTarget = targetNode.Name
				}
			}
			if actionTarget == "" {
				actionTarget = "Unknown Target"
			}
		}

		chars = append(chars, charData{
			CharacterName: c.Name,
			ActionName:    action.Name,
			ActionTarget:  actionTarget,
		})
	}

	nodeNames := make([]string, 0, len(resourceNodes))
	for _, node := range resourceNodes {
		resourceNode, err := cfg.GetResourceNodeById(r.Context(), node.NodeID)
		if err != nil {
			continue
		}
		nodeNames = append(nodeNames, resourceNode.Name)
	}

	area := area{
		PositionX:     char.PositionX,
		PositionY:     char.PositionY,
		Characters:    chars,
		ResourceNodes: nodeNames,
	}

	respondWithJSON(w, http.StatusOK, area)
}

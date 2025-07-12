package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/trbute/idler/server/internal/auth"
	"github.com/trbute/idler/server/internal/database"
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

	ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
	defer cancel()

	const actionsCacheKey = "actions:all"
	cachedData, err := cfg.Redis.Get(ctx, actionsCacheKey).Bytes()
	if err == nil {
		var actionResponse []Action
		if err := json.Unmarshal(cachedData, &actionResponse); err == nil {
			respondWithJSON(w, http.StatusOK, actionResponse)
			return
		}
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

	if jsonData, err := json.Marshal(actionResponse); err == nil {
		err = cfg.Redis.Set(ctx, actionsCacheKey, jsonData, 24*time.Hour).Err()
		if err != nil {
			log.Printf("Failed to cache actions: %v", err)
		}
	}

	respondWithJSON(w, http.StatusOK, actionResponse)
}

func (cfg *ApiConfig) GetActionById(ctx context.Context, actionID int32) (database.Action, error) {
	cacheKey := fmt.Sprintf("action:%d", actionID)
	
	cached, err := cfg.Redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var action database.Action
		if json.Unmarshal([]byte(cached), &action) == nil {
			return action, nil
		}
	}
	
	action, err := cfg.DB.GetActionById(ctx, actionID)
	if err != nil {
		return database.Action{}, err
	}
	
	if data, err := json.Marshal(action); err == nil {
		cfg.Redis.Set(ctx, cacheKey, data, 24*time.Hour)
	}
	
	return action, nil
}

func (cfg *ApiConfig) GetActionByName(ctx context.Context, name string) (database.Action, error) {
	cacheKey := fmt.Sprintf("action:name:%s", name)
	
	cached, err := cfg.Redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var action database.Action
		if json.Unmarshal([]byte(cached), &action) == nil {
			return action, nil
		}
	}
	
	action, err := cfg.DB.GetActionByName(ctx, name)
	if err != nil {
		return database.Action{}, err
	}
	
	if data, err := json.Marshal(action); err == nil {
		cfg.Redis.Set(ctx, cacheKey, data, 24*time.Hour)
	}
	
	return action, nil
}

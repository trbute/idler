package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trbute/idler/internal/auth"
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

	redisKey := "actions"

	cachedData, err := cfg.Redis.Get(r.Context(), redisKey).Result()
	if err == nil {
		decoder := json.NewDecoder(strings.NewReader(cachedData))
		actionResponse := []Action{}
		err := decoder.Decode(&actionResponse)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Unable to decode parameters", err)
			return
		}

		respondWithJSON(w, http.StatusOK, actionResponse)
		return
	} else if err != redis.Nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"Failed to retrieve data from Redis",
			err,
		)
		return
	}

	actions, err := cfg.DB.GetAllActions(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve actions", err)
		return
	}

	dat, err := json.Marshal(actions)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error marshalling JSON: %s", err)
		return
	}

	err = cfg.Redis.Set(r.Context(), redisKey, dat, 5*time.Minute).Err()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to store actions in redis", err)
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

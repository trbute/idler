package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trbute/idler/server/internal/auth"
	"github.com/trbute/idler/server/internal/database"
)

func (cfg *ApiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to decode parameters", err)
		return
	}

	user, err := cfg.DB.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	err = auth.CheckPasswordHash(user.HashedPassword, params.Password)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	expiresInSeconds := 3600

	expireDuration := time.Duration(time.Duration(expiresInSeconds) * time.Second)

	userid := uuid.UUID(user.ID.Bytes)

	token, err := auth.MakeJWT(userid, cfg.JwtSecret, expireDuration)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "JWT creation failed", err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Refresh token creation failed", err)
		return
	}

	day := 24 * time.Hour
	refreshExpire := time.Now().Add(60 * day)

	pgTimestamp := pgtype.Timestamp{
		Time:  refreshExpire,
		Valid: true,
	}

	_, err = cfg.DB.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: pgTimestamp,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Refresh token db insert failed", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
		Token:        token,
		RefreshToken: refreshToken,
	})
}

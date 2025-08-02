package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trbute/idler/server/internal/auth"
	"github.com/trbute/idler/server/internal/database"
	"github.com/trbute/idler/server/internal/validation"
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

	if err := validation.ValidateEmail(params.Email); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if params.Password == "" {
		respondWithError(w, http.StatusBadRequest, "password is required", nil)
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

	userid := uuid.UUID(user.ID.Bytes)

	err = cfg.DB.RevokeAllUserTokens(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to revoke existing refresh tokens", err)
		return
	}

	blacklistedTokens, err := auth.BlacklistAllUserTokens(r.Context(), userid, cfg.Redis)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to revoke existing sessions", err)
		return
	}

	// Disconnect websocket clients with blacklisted tokens
	for _, tokenID := range blacklistedTokens {
		cfg.Hub.DisconnectClientByToken(tokenID)
	}

	expiresInSeconds := 3600
	expireDuration := time.Duration(time.Duration(expiresInSeconds) * time.Second)

	token, err := auth.MakeJWT(userid, cfg.JwtSecret, expireDuration)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "JWT creation failed", err)
		return
	}

	parsedToken, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JwtSecret), nil
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to parse created token", err)
		return
	}

	claims, ok := parsedToken.Claims.(*jwt.RegisteredClaims)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Invalid token claims type", nil)
		return
	}
	err = auth.TrackUserToken(r.Context(), userid, claims.ID, cfg.Redis)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to track new token", err)
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

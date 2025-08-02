package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trbute/idler/server/internal/auth"
	"github.com/trbute/idler/server/internal/database"
	"github.com/trbute/idler/server/internal/validation"
)

type User struct {
	ID        pgtype.UUID      `json:"id"`
	CreatedAt pgtype.Timestamp `json:"created_at"`
	UpdatedAt pgtype.Timestamp `json:"updated_at"`
	Email     string           `json:"email"`
	Surname   string           `json:"surname,omitempty"`
}

func (cfg *ApiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	type Parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Surname  string `json:"surname"`
	}

	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to decode parameters", err)
		return
	}

	if err := validation.ValidateEmail(params.Email); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := validation.ValidatePassword(params.Password); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := validation.ValidateSurname(params.Surname); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Password hash failed", err)
		return
	}

	user, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
		Surname:        pgtype.Text{String: params.Surname, Valid: true},
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "User creation failed", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Surname:   user.Surname.String,
	})
}

func (cfg *ApiConfig) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	type Parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to decode parameters", err)
		return
	}

	if err := validation.ValidateEmail(params.Email); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := validation.ValidatePassword(params.Password); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to retrieve token", err)
		return
	}

	userID, err := auth.ValidateJWTWithBlacklist(r.Context(), token, cfg.JwtSecret, cfg.Redis)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token invalid", err)
		return
	}

	pgUserID := pgtype.UUID{
		Bytes: userID,
		Valid: true,
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Password hash failed", err)
		return
	}

	user, err := cfg.DB.UpdateUserById(r.Context(), database.UpdateUserByIdParams{
		ID:             pgUserID,
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database update failed", err)
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Surname:   user.Surname.String,
	})
}

func (cfg *ApiConfig) GetSurnameById(ctx context.Context, userID uuid.UUID) (string, error) {
	cacheKey := fmt.Sprintf("user:surname:%s", userID.String())

	cached, err := cfg.Redis.Get(ctx, cacheKey).Result()
	if err == nil {
		return cached, nil
	}

	pgUserID := pgtype.UUID{
		Bytes: userID,
		Valid: true,
	}

	surname, err := cfg.DB.GetSurnameById(ctx, pgUserID)
	if err != nil {
		return "", err
	}

	if !surname.Valid {
		return "", sql.ErrNoRows
	}

	cfg.Redis.Set(ctx, cacheKey, surname.String, 5*time.Minute)
	return surname.String, nil
}

package api

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/trbute/idler/server/internal/auth"
	"github.com/trbute/idler/server/internal/websocket"
)

func (cfg *ApiConfig) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	userID, err := auth.ValidateJWTWithBlacklist(r.Context(), token, cfg.JwtSecret, cfg.Redis)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Extract token ID from JWT
	parsedToken, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JwtSecret), nil
	})
	if err != nil {
		http.Error(w, "Invalid token format", http.StatusUnauthorized)
		return
	}

	claims := parsedToken.Claims.(*jwt.RegisteredClaims)
	tokenID := claims.ID

	websocket.ServeWS(cfg.Hub, cfg, w, r, userID, tokenID)
}

func (cfg *ApiConfig) ValidateSpecificToken(ctx context.Context, tokenID string) error {
	return auth.ValidateSpecificToken(ctx, tokenID, cfg.Redis)
}
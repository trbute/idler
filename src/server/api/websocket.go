package api

import (
	"net/http"

	"github.com/trbute/idler/server/internal/auth"
	"github.com/trbute/idler/server/internal/websocket"
)

func (cfg *ApiConfig) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get JWT token from query parameter (since WebSocket doesn't support headers easily)
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.JwtSecret)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	websocket.ServeWS(cfg.Hub, cfg, w, r, userID)
}
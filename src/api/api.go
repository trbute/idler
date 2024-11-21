package api

import (
	"net/http"

	"github.com/trbute/idler/internal/database"
	"github.com/trbute/idler/internal/world"
)

type ApiConfig struct {
	DB        *database.Queries
	Platform  string
	JwtSecret string
	World     *world.World
}

func (cfg *ApiConfig) ServeApi() {
	port := "8080"
	mux := http.NewServeMux()
	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	mux.HandleFunc("POST /api/users", cfg.handleCreateUser)
	mux.HandleFunc("PUT /api/users", cfg.handleUpdateUser)
	mux.HandleFunc("POST /api/characters", cfg.handleCreateCharacter)
	mux.HandleFunc("PUT /api/characters", cfg.handleUpdateCharacter)
	mux.HandleFunc("GET /api/actions", cfg.handleGetActions)
	mux.HandleFunc("POST /api/login", cfg.handleLogin)
	mux.HandleFunc("POST /api/refresh", cfg.handleRefresh)
	mux.HandleFunc("POST /api/revoke", cfg.handleRevoke)

	server.ListenAndServe()
}

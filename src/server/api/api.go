package api

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/trbute/idler/server/internal/database"
	"github.com/trbute/idler/server/internal/websocket"
)

type ApiConfig struct {
	DB        *database.Queries
	JwtSecret string
	Redis     *redis.Client
	Pool      *pgxpool.Pool
	Hub       *websocket.Hub
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
	mux.HandleFunc("GET /api/sense/area/{character}", cfg.handleGetArea)
	mux.HandleFunc("GET /api/inventory/{character}", cfg.handleGetInventory)
	mux.HandleFunc("POST /api/login", cfg.handleLogin)
	mux.HandleFunc("POST /api/refresh", cfg.handleRefresh)
	mux.HandleFunc("POST /api/revoke", cfg.handleRevoke)
	mux.HandleFunc("GET /ws", cfg.handleWebSocket)

	server.ListenAndServe()
}

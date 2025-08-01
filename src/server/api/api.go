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

	mux.Handle("POST /api/users", http.HandlerFunc(cfg.handleCreateUser))
	mux.Handle("PUT /api/users", http.HandlerFunc(cfg.handleUpdateUser))
	mux.Handle("POST /api/characters", http.HandlerFunc(cfg.handleCreateCharacter))
	mux.Handle("PUT /api/characters", http.HandlerFunc(cfg.handleUpdateCharacter))
	mux.Handle("GET /api/characters/{character}/select", http.HandlerFunc(cfg.handleSelectCharacter))
	mux.Handle("GET /api/actions", http.HandlerFunc(cfg.handleGetActions))
	mux.Handle("GET /api/sense/area/{character}", http.HandlerFunc(cfg.handleGetArea))
	mux.Handle("GET /api/inventory/{character}", http.HandlerFunc(cfg.handleGetInventory))
	mux.Handle("POST /api/inventory/drop", http.HandlerFunc(cfg.handleDropItem))
	mux.Handle("POST /api/login", http.HandlerFunc(cfg.handleLogin))
	mux.Handle("POST /api/refresh", http.HandlerFunc(cfg.handleRefresh))
	mux.Handle("POST /api/revoke", http.HandlerFunc(cfg.handleRevoke))
	mux.Handle("GET /ws", http.HandlerFunc(cfg.handleWebSocket))

	server.ListenAndServe()
}

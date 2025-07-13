package api

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/trbute/idler/server/internal/database"
	"github.com/trbute/idler/server/internal/ratelimit"
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

	rateLimiter := ratelimit.NewRateLimiter(cfg.Redis)
	
	authRateLimit := ratelimit.RateLimit{Requests: 5, Window: time.Minute}
	gameRateLimit := ratelimit.RateLimit{Requests: 60, Window: time.Minute}
	readRateLimit := ratelimit.RateLimit{Requests: 120, Window: time.Minute}
	mux.Handle("POST /api/users", rateLimiter.Middleware(authRateLimit)(http.HandlerFunc(cfg.handleCreateUser)))
	mux.Handle("PUT /api/users", rateLimiter.Middleware(gameRateLimit)(http.HandlerFunc(cfg.handleUpdateUser)))
	mux.Handle("POST /api/characters", rateLimiter.Middleware(gameRateLimit)(http.HandlerFunc(cfg.handleCreateCharacter)))
	mux.Handle("PUT /api/characters", rateLimiter.Middleware(gameRateLimit)(http.HandlerFunc(cfg.handleUpdateCharacter)))
	mux.Handle("GET /api/characters/{character}/select", rateLimiter.Middleware(readRateLimit)(http.HandlerFunc(cfg.handleSelectCharacter)))
	mux.Handle("GET /api/actions", rateLimiter.Middleware(readRateLimit)(http.HandlerFunc(cfg.handleGetActions)))
	mux.Handle("GET /api/sense/area/{character}", rateLimiter.Middleware(readRateLimit)(http.HandlerFunc(cfg.handleGetArea)))
	mux.Handle("GET /api/inventory/{character}", rateLimiter.Middleware(readRateLimit)(http.HandlerFunc(cfg.handleGetInventory)))
	mux.Handle("POST /api/inventory/drop", rateLimiter.Middleware(gameRateLimit)(http.HandlerFunc(cfg.handleDropItem)))
	mux.Handle("POST /api/login", rateLimiter.Middleware(authRateLimit)(http.HandlerFunc(cfg.handleLogin)))
	mux.Handle("POST /api/refresh", rateLimiter.Middleware(authRateLimit)(http.HandlerFunc(cfg.handleRefresh)))
	mux.Handle("POST /api/revoke", rateLimiter.Middleware(authRateLimit)(http.HandlerFunc(cfg.handleRevoke)))
	mux.Handle("GET /ws", rateLimiter.Middleware(readRateLimit)(http.HandlerFunc(cfg.handleWebSocket)))

	server.ListenAndServe()
}

package api

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/cors"
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
	Limiter   *ratelimit.Limiter
}

func (cfg *ApiConfig) setupCORS(handler http.Handler) http.Handler {
	allowedOrigins := strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ",")
	if len(allowedOrigins) == 0 || (len(allowedOrigins) == 1 && allowedOrigins[0] == "") {
		allowedOrigins = []string{"http://localhost:*", "https://localhost:*"}
	}

	c := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Content-Type",
			"Authorization",
			"X-Requested-With",
		},
		ExposedHeaders: []string{
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
		},
		AllowCredentials: true,
		MaxAge:           300,
	})

	return c.Handler(handler)
}

func (cfg *ApiConfig) ServeApi() {
	port := "8080"
	mux := http.NewServeMux()

	apiRateLimit := cfg.Limiter.Middleware(cfg.JwtSecret, 100, time.Minute)

	mux.Handle("POST /api/users", apiRateLimit(http.HandlerFunc(cfg.handleCreateUser)))
	mux.Handle("PUT /api/users", apiRateLimit(http.HandlerFunc(cfg.handleUpdateUser)))
	mux.Handle("POST /api/characters", apiRateLimit(http.HandlerFunc(cfg.handleCreateCharacter)))
	mux.Handle("PUT /api/characters", apiRateLimit(http.HandlerFunc(cfg.handleUpdateCharacter)))
	mux.Handle("GET /api/characters/{character}/select", apiRateLimit(http.HandlerFunc(cfg.handleSelectCharacter)))
	mux.Handle("GET /api/actions", apiRateLimit(http.HandlerFunc(cfg.handleGetActions)))
	mux.Handle("GET /api/sense/area/{character}", apiRateLimit(http.HandlerFunc(cfg.handleGetArea)))
	mux.Handle("GET /api/inventory/{character}", apiRateLimit(http.HandlerFunc(cfg.handleGetInventory)))
	mux.Handle("POST /api/inventory/drop", apiRateLimit(http.HandlerFunc(cfg.handleDropItem)))
	mux.Handle("POST /api/login", apiRateLimit(http.HandlerFunc(cfg.handleLogin)))
	mux.Handle("POST /api/refresh", apiRateLimit(http.HandlerFunc(cfg.handleRefresh)))
	mux.Handle("POST /api/revoke", apiRateLimit(http.HandlerFunc(cfg.handleRevoke)))
	mux.Handle("GET /ws", http.HandlerFunc(cfg.handleWebSocket))

	handler := cfg.setupCORS(mux)
	
	server := &http.Server{
		Handler: handler,
		Addr:    ":" + port,
	}

	server.ListenAndServe()
}

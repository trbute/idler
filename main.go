package main

import (
	"database/sql"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/trbute/idler/internal/database"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	jwtSecret      string
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env")
	}

	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	jwtSecret := os.Getenv("JWT_SECRET")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Unable to connect to db")
	}

	port := "8080"
	mux := http.NewServeMux()
	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	apiCfg := apiConfig{
		db:        database.New(db),
		platform:  platform,
		jwtSecret: jwtSecret,
	}

	go mux.HandleFunc("POST /api/users", apiCfg.handleCreateUser)
	go mux.HandleFunc("PUT /api/users", apiCfg.handleUpdateUser)
	go mux.HandleFunc("POST /api/characters", apiCfg.handleCreateCharacter)
	go mux.HandleFunc("POST /api/login", apiCfg.handleLogin)
	go mux.HandleFunc("POST /api/refresh", apiCfg.handleRefresh)
	go mux.HandleFunc("POST /api/revoke", apiCfg.handleRevoke)

	server.ListenAndServe()
}

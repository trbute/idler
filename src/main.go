package main

import (
	"database/sql"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/trbute/idler/api"
	"github.com/trbute/idler/internal/database"
	"github.com/trbute/idler/internal/world"
)

func main() {
	if os.Getenv("IS_DOCKER") == "" {
		if err := godotenv.Load("../.env"); err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	jwtSecret := os.Getenv("JWT_SECRET")

	tickInt, err := strconv.Atoi(os.Getenv("TICK_MS"))
	tickRate := time.Duration(time.Duration(tickInt) * time.Millisecond)

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Unable to connect to db")
	}

	conn := database.New(db)

	apiCfg := api.ApiConfig{
		DB:        conn,
		Platform:  platform,
		JwtSecret: jwtSecret,
	}

	serverCfg := world.ServerConfig{
		DB:       conn,
		Platform: platform,
		TickRate: tickRate,
	}

	apiCfg.ServeApi()
	serverCfg.ProcessTicks()
}

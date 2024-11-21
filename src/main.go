package main

import (
	"database/sql"
	"fmt"
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
	if os.Getenv("PLATFORM") != "docker" {
		if err := godotenv.Load("../.env"); err != nil {
			log.Fatal(err)
		}
	}

	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser,
		dbPass,
		dbHost,
		dbPort,
		dbName,
	)
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Unable to connect to db")
	}
	DbConn := database.New(db)

	platform := os.Getenv("PLATFORM")
	jwtSecret := os.Getenv("JWT_SECRET")

	tickInt, err := strconv.Atoi(os.Getenv("TICK_MS"))
	tickRate := time.Duration(time.Duration(tickInt) * time.Millisecond)

	worldCfg := world.WorldConfig{
		DB:       DbConn,
		Platform: platform,
		TickRate: tickRate,
	}

	world := worldCfg.GetWorld()
	worldCfg.World = world

	apiCfg := api.ApiConfig{
		DB:        DbConn,
		Platform:  platform,
		JwtSecret: jwtSecret,
		World:     world,
	}

	go worldCfg.ProcessTicks()
	apiCfg.ServeApi()
}

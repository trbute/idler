package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/trbute/idler/server/api"
	"github.com/trbute/idler/server/internal/database"
	"github.com/trbute/idler/server/internal/world"
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

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to the database: %v", err)
	}
	defer pool.Close()

	// Use the connection pool with sqlc
	DbConn := database.New(pool)

	platform := os.Getenv("PLATFORM")
	jwtSecret := os.Getenv("JWT_SECRET")

	tickInt, err := strconv.Atoi(os.Getenv("TICK_MS"))
	tickRate := time.Duration(time.Duration(tickInt) * time.Millisecond)
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))

	worldCfg := world.WorldConfig{
		DB:       DbConn,
		Platform: platform,
		TickRate: tickRate,
		Seed:     seed,
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

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
	_ "github.com/lib/pq"
	"github.com/trbute/idler/server/api"
	"github.com/trbute/idler/server/data"
	"github.com/trbute/idler/server/internal/database"
	"github.com/trbute/idler/server/internal/world"
)

func main() {

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

	DbConn := database.New(pool)

	jwtSecret := os.Getenv("JWT_SECRET")

	dataCfg := data.DataConfig{
		DB: DbConn,
	}

	dataCfg.InitData()

	tickInt, err := strconv.Atoi(os.Getenv("TICK_MS"))
	if err != nil {
		log.Fatalf("Unable to convert TICK_MS to int: %v", err)
	}

	tickRate := time.Duration(time.Duration(tickInt) * time.Millisecond)
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))

	worldCfg := world.WorldConfig{
		DB:       DbConn,
		TickRate: tickRate,
		Seed:     seed,
	}

	world := worldCfg.GetWorld()
	worldCfg.World = world

	apiCfg := api.ApiConfig{
		DB:        DbConn,
		JwtSecret: jwtSecret,
		World:     world,
	}

	go worldCfg.ProcessTicks()
	apiCfg.ServeApi()
}

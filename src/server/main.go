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
	"github.com/redis/go-redis/v9"
	"github.com/trbute/idler/server/api"
	"github.com/trbute/idler/server/data"
	"github.com/trbute/idler/server/internal/database"
	"github.com/trbute/idler/server/internal/websocket"
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

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Unable to parse database config: %v", err)
	}
	
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 15 * time.Minute
	config.HealthCheckPeriod = 30 * time.Second
	
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
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

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDRESS"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Unable to connect to Redis: %v", err)
	}
	defer rdb.Close()

	hub := websocket.NewHub()
	go hub.Run()

	apiCfg := api.ApiConfig{
		DB:        DbConn,
		JwtSecret: jwtSecret,
		Redis:     rdb,
		Pool:      pool,
		Hub:       hub,
	}

	worldCfg := world.WorldConfig{
		DB:        DbConn,
		Redis:     rdb,
		TickRate:  tickRate,
		Seed:      seed,
		ApiConfig: &apiCfg,
	}

	go worldCfg.ProcessTicks()
	apiCfg.ServeApi()
}

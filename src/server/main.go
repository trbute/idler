package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
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
	"github.com/trbute/idler/server/internal/ratelimit"
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
	if err := validateJWTSecret(jwtSecret); err != nil {
		log.Fatalf("JWT_SECRET validation failed: %v", err)
	}

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

	limiter := ratelimit.NewLimiter(rdb)

	apiCfg := api.ApiConfig{
		DB:        DbConn,
		JwtSecret: jwtSecret,
		Redis:     rdb,
		Pool:      pool,
		Hub:       hub,
		Limiter:   limiter,
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

func validateJWTSecret(secret string) error {
	if secret == "" {
		return errors.New("JWT_SECRET environment variable is required")
	}
	
	if len(secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long, got %d", len(secret))
	}
	
	if len(secret) > 512 {
		return fmt.Errorf("JWT_SECRET should not exceed 512 characters, got %d", len(secret))
	}
	
	entropy := calculateEntropy(secret)
	minEntropy := 4.0
	if entropy < minEntropy {
		return fmt.Errorf("JWT_SECRET has insufficient entropy: %.2f (minimum: %.2f)", entropy, minEntropy)
	}
	
	return nil
}

func calculateEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}
	
	charCount := make(map[rune]int)
	for _, char := range s {
		charCount[char]++
	}
	
	var entropy float64
	length := float64(len(s))
	
	for _, count := range charCount {
		probability := float64(count) / length
		if probability > 0 {
			entropy -= probability * math.Log2(probability)
		}
	}
	
	return entropy
}

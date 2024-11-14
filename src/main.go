package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
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

	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisURL := fmt.Sprintf(
		"redis://:@%s:%s/0?protocol=3",
		redisHost,
		redisPort,
	)
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatal("Unable to connect to redis")
	}
	redisConn := redis.NewClient(opts)
	defer redisConn.Close()

	err = redisConn.FlushAll(context.Background()).Err()
	if err != nil {
		log.Fatalf("Failed to flush Redis cache: %v", err)
	}

	platform := os.Getenv("PLATFORM")
	jwtSecret := os.Getenv("JWT_SECRET")

	tickInt, err := strconv.Atoi(os.Getenv("TICK_MS"))
	tickRate := time.Duration(time.Duration(tickInt) * time.Millisecond)

	apiCfg := api.ApiConfig{
		DB:        DbConn,
		Redis:     redisConn,
		Platform:  platform,
		JwtSecret: jwtSecret,
	}

	serverCfg := world.ServerConfig{
		DB:       DbConn,
		Redis:    redisConn,
		Platform: platform,
		TickRate: tickRate,
	}

	apiCfg.ServeApi()
	serverCfg.ProcessTicks()
}

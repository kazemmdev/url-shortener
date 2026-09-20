package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/microsoft/go-mssqldb"
	"github.com/redis/go-redis/v9"
)

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func main() {
	httpPort := env("HTTP_PORT", "8080")
	sqlHost := env("SQLSERVER_HOST", "sqlserver")
	sqlPort := env("SQLSERVER_PORT", "1433")
	sqlUser := env("SQLSERVER_USER", "sa")
	sqlPassword := os.Getenv("SQLSERVER_PASSWORD")
	sqlDatabase := env("SQLSERVER_DATABASE", "UrlShortener")
	redisAddr := env("REDIS_ADDR", "redis:6379")
	expireDays := envInt("EXPIRE_DAYS", 30)

	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s&TrustServerCertificate=true",
		sqlUser, sqlPassword, sqlHost, sqlPort, sqlDatabase)

	db, err := sql.Open("sqlserver", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	if err := bootstrapSchema(ctx, db); err != nil {
		log.Fatalf("bootstrap schema: %v", err)
	}

	cache := redis.NewClient(&redis.Options{Addr: redisAddr})
	if err := cache.Ping(ctx).Err(); err != nil {
		log.Fatalf("ping redis: %v", err)
	}
	defer cache.Close()

	s := &server{db: db, cache: cache, expireDays: expireDays}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/url", s.handleCreate)
	mux.HandleFunc("GET /api/url/{shortCode}", s.handleRedirect)

	log.Printf("listening on :%s", httpPort)
	log.Fatal(http.ListenAndServe(":"+httpPort, mux))
}

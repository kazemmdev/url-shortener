package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	HTTPPort     string
	SQLServerDSN string
	RedisAddr    string
	ExpireDays   int
}

func Load() Config {
	sqlHost := env("SQLSERVER_HOST", "sqlserver")
	sqlPort := env("SQLSERVER_PORT", "1433")
	sqlUser := env("SQLSERVER_USER", "sa")
	sqlPassword := os.Getenv("SQLSERVER_PASSWORD")
	sqlDatabase := env("SQLSERVER_DATABASE", "UrlShortener")

	return Config{
		HTTPPort: env("HTTP_PORT", "8080"),
		SQLServerDSN: fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s&TrustServerCertificate=true",
			sqlUser, sqlPassword, sqlHost, sqlPort, sqlDatabase),
		RedisAddr:  env("REDIS_ADDR", "redis:6379"),
		ExpireDays: envInt("EXPIRE_DAYS", 30),
	}
}

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

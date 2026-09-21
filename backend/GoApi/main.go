package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"goapi/internal/cache"
	"goapi/internal/config"
	"goapi/internal/handler"
	"goapi/internal/store"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	setupCtx, cancelSetup := context.WithTimeout(ctx, 30*time.Second)
	defer cancelSetup()

	db, err := store.Open(setupCtx, cfg.SQLServerDSN)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer db.Close()

	st := store.New(db, cfg.ExpireDays)
	if err := st.Bootstrap(setupCtx); err != nil {
		log.Fatalf("bootstrap schema: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer redisClient.Close()
	if err := redisClient.Ping(setupCtx).Err(); err != nil {
		log.Fatalf("ping redis: %v", err)
	}

	srv := handler.New(st, cache.New(redisClient))
	httpServer := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: srv.Routes(),
	}

	go func() {
		<-ctx.Done()
		log.Println("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown: %v", err)
		}
	}()

	log.Printf("listening on :%s", cfg.HTTPPort)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %v", err)
	}
}

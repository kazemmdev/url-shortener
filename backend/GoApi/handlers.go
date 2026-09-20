package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/redis/go-redis/v9"
)

type server struct {
	db         *sql.DB
	cache      *redis.Client
	expireDays int
}

type createRequest struct {
	LongUrl string `json:"longUrl"`
}

func (s *server) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.LongUrl == "" {
		http.Error(w, "longUrl is required", http.StatusBadRequest)
		return
	}

	shortCode, err := createUrl(r.Context(), s.db, req.LongUrl, s.expireDays)
	if err != nil {
		log.Printf("create url: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shortCode)
}

func (s *server) handleRedirect(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("shortCode")
	ctx := r.Context()

	if longUrl, ok := s.getFromCache(ctx, shortCode); ok {
		http.Redirect(w, r, longUrl, http.StatusFound)
		return
	}

	longUrl, found, err := getLongUrlFromDb(ctx, s.db, shortCode)
	if err != nil {
		log.Printf("get url: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !found {
		http.NotFound(w, r)
		return
	}

	s.setCache(ctx, shortCode, longUrl)
	http.Redirect(w, r, longUrl, http.StatusFound)
}

func (s *server) getFromCache(ctx context.Context, shortCode string) (string, bool) {
	longUrl, err := s.cache.Get(ctx, shortCode).Result()
	if err == redis.Nil {
		return "", false
	}
	if err != nil {
		log.Printf("cache get: %v", err)
		return "", false
	}
	return longUrl, true
}

func (s *server) setCache(ctx context.Context, shortCode, longUrl string) {
	// No expiration, matching DotnetApi: links never change once created, so
	// the cache entry is valid for as long as the link itself is.
	if err := s.cache.Set(ctx, shortCode, longUrl, 0).Err(); err != nil {
		log.Printf("cache set: %v", err)
	}
}

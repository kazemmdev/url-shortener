package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
)

const maxLongURLLength = 2048

type Store interface {
	CreateURL(ctx context.Context, longURL string) (string, error)
	GetLongURL(ctx context.Context, shortCode string) (string, bool, error)
}

type Cache interface {
	Get(ctx context.Context, shortCode string) (string, bool)
	Set(ctx context.Context, shortCode, longURL string)
}

type Server struct {
	store Store
	cache Cache
}

func New(store Store, cache Cache) *Server {
	return &Server{store: store, cache: cache}
}

func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/url", s.handleCreate)
	mux.HandleFunc("GET /api/url/{shortCode}", s.handleRedirect)
	return mux
}

type createRequest struct {
	LongUrl string `json:"longUrl"`
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "longUrl is required", http.StatusBadRequest)
		return
	}
	if !isValidLongURL(req.LongUrl) {
		http.Error(w, "longUrl must be an absolute http(s) URL", http.StatusBadRequest)
		return
	}

	shortCode, err := s.store.CreateURL(r.Context(), req.LongUrl)
	if err != nil {
		log.Printf("create url: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(shortCode); err != nil {
		log.Printf("encode response: %v", err)
	}
}

func (s *Server) handleRedirect(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("shortCode")
	ctx := r.Context()

	if longURL, ok := s.cache.Get(ctx, shortCode); ok {
		http.Redirect(w, r, longURL, http.StatusFound)
		return
	}

	longURL, found, err := s.store.GetLongURL(ctx, shortCode)
	if err != nil {
		log.Printf("get url: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !found {
		http.NotFound(w, r)
		return
	}

	s.cache.Set(ctx, shortCode, longURL)
	http.Redirect(w, r, longURL, http.StatusFound)
}

func isValidLongURL(raw string) bool {
	if raw == "" || len(raw) > maxLongURLLength {
		return false
	}
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

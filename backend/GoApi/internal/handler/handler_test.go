package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeStore struct {
	createCode string
	createErr  error
	getURL     string
	getFound   bool
	getErr     error
}

func (f *fakeStore) CreateURL(ctx context.Context, longURL string) (string, error) {
	return f.createCode, f.createErr
}

func (f *fakeStore) GetLongURL(ctx context.Context, shortCode string) (string, bool, error) {
	return f.getURL, f.getFound, f.getErr
}

type fakeCache struct {
	hitURL string
	hit    bool
}

func (f *fakeCache) Get(ctx context.Context, shortCode string) (string, bool) {
	return f.hitURL, f.hit
}

func (f *fakeCache) Set(ctx context.Context, shortCode, longURL string) {}

func TestHandleCreate_InvalidBody(t *testing.T) {
	srv := New(&fakeStore{}, &fakeCache{})
	req := httptest.NewRequest(http.MethodPost, "/api/url", strings.NewReader(`not json`))
	rec := httptest.NewRecorder()

	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandleCreate_InvalidLongURL(t *testing.T) {
	cases := []string{"", "not-a-url", "ftp://example.com/file", strings.Repeat("a", 3000)}

	for _, longURL := range cases {
		srv := New(&fakeStore{}, &fakeCache{})
		body, _ := json.Marshal(map[string]string{"longUrl": longURL})
		req := httptest.NewRequest(http.MethodPost, "/api/url", strings.NewReader(string(body)))
		rec := httptest.NewRecorder()

		srv.Routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("longUrl %q: status = %d, want %d", longURL, rec.Code, http.StatusBadRequest)
		}
	}
}

func TestHandleCreate_Success(t *testing.T) {
	srv := New(&fakeStore{createCode: "abc1234"}, &fakeCache{})
	body, _ := json.Marshal(map[string]string{"longUrl": "https://example.com/page"})
	req := httptest.NewRequest(http.MethodPost, "/api/url", strings.NewReader(string(body)))
	rec := httptest.NewRecorder()

	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got != "abc1234" {
		t.Errorf("shortCode = %q, want %q", got, "abc1234")
	}
}

func TestHandleRedirect_CacheHit(t *testing.T) {
	srv := New(&fakeStore{}, &fakeCache{hitURL: "https://example.com/cached", hit: true})
	req := httptest.NewRequest(http.MethodGet, "/api/url/abc1234", nil)
	rec := httptest.NewRecorder()

	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if loc := rec.Header().Get("Location"); loc != "https://example.com/cached" {
		t.Errorf("Location = %q, want %q", loc, "https://example.com/cached")
	}
}

func TestHandleRedirect_NotFound(t *testing.T) {
	srv := New(&fakeStore{getFound: false}, &fakeCache{})
	req := httptest.NewRequest(http.MethodGet, "/api/url/missing", nil)
	rec := httptest.NewRecorder()

	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

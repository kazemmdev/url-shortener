package cache

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
}

func New(client *redis.Client) *Cache {
	return &Cache{client: client}
}

func (c *Cache) Get(ctx context.Context, shortCode string) (string, bool) {
	longURL, err := c.client.Get(ctx, shortCode).Result()
	if err == redis.Nil {
		return "", false
	}
	if err != nil {
		log.Printf("cache get: %v", err)
		return "", false
	}
	return longURL, true
}

func (c *Cache) Set(ctx context.Context, shortCode, longURL string) {
	// No expiration, matching DotnetApi: links never change once created, so
	// the cache entry is valid for as long as the link itself is.
	if err := c.client.Set(ctx, shortCode, longURL, 0).Err(); err != nil {
		log.Printf("cache set: %v", err)
	}
}

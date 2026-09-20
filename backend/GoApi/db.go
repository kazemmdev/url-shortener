package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Matches the schema DotnetApi's EF Core migration creates (snake_case
// table/columns), so both APIs can run against the same database.
const bootstrapSchemaSQL = `
IF OBJECT_ID('urls', 'U') IS NULL
CREATE TABLE urls (
    short_code nvarchar(450) NOT NULL PRIMARY KEY,
    long_url nvarchar(max) NOT NULL,
    created_at datetime2 NOT NULL,
    expires_at datetime2 NOT NULL
);`

func bootstrapSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, bootstrapSchemaSQL)
	return err
}

func createUrl(ctx context.Context, db *sql.DB, longUrl string, expireDays int) (string, error) {
	var count int64
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM urls").Scan(&count); err != nil {
		return "", fmt.Errorf("count urls: %w", err)
	}

	shortCode, err := generateShortCode(count)
	if err != nil {
		return "", fmt.Errorf("generate short code: %w", err)
	}

	now := time.Now().UTC()
	expiresAt := now.AddDate(0, 0, expireDays)

	_, err = db.ExecContext(ctx,
		"INSERT INTO urls (short_code, long_url, created_at, expires_at) VALUES (@p1, @p2, @p3, @p4)",
		shortCode, longUrl, now, expiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("insert url: %w", err)
	}

	return shortCode, nil
}

func getLongUrlFromDb(ctx context.Context, db *sql.DB, shortCode string) (string, bool, error) {
	var longUrl string
	var expiresAt time.Time

	err := db.QueryRowContext(ctx,
		"SELECT long_url, expires_at FROM urls WHERE short_code = @p1", shortCode,
	).Scan(&longUrl, &expiresAt)

	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("select url: %w", err)
	}
	if expiresAt.Before(time.Now().UTC()) {
		return "", false, nil
	}

	return longUrl, true, nil
}

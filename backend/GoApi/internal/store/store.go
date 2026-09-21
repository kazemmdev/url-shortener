package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	// Registers the "sqlserver" driver used by Open below; also gives us
	// mssql.Error to detect duplicate-key collisions in isDuplicateShortCode.
	mssql "github.com/microsoft/go-mssqldb"

	"goapi/internal/shortcode"
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

// Two concurrent creates can read the same row count before either inserts,
// so the short code they generate can collide. That surfaces as a duplicate
// primary key on insert; retry a few times with a fresh count rather than
// failing the request outright.
const maxCreateAttempts = 5

// Open dials SQL Server and configures the connection pool. database/sql
// defaults to unlimited open connections but only 2 kept idle, so under
// concurrent load it was opening a fresh TCP+auth connection per request
// instead of reusing a pool. Cap it at 100, matching ADO.NET SqlClient's
// default Max Pool Size, so this isn't an artificially different limit from
// the .NET side.
func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlserver", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(100)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return db, nil
}

type Store struct {
	db         *sql.DB
	expireDays int
}

func New(db *sql.DB, expireDays int) *Store {
	return &Store{db: db, expireDays: expireDays}
}

func (s *Store) Bootstrap(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, bootstrapSchemaSQL)
	return err
}

func (s *Store) CreateURL(ctx context.Context, longURL string) (string, error) {
	var lastErr error
	for attempt := 0; attempt < maxCreateAttempts; attempt++ {
		code, err := s.tryCreateURL(ctx, longURL)
		if err == nil {
			return code, nil
		}
		if !isDuplicateShortCode(err) {
			return "", err
		}
		lastErr = err
	}
	return "", fmt.Errorf("insert url: gave up after %d short code collisions: %w", maxCreateAttempts, lastErr)
}

func (s *Store) tryCreateURL(ctx context.Context, longURL string) (string, error) {
	var count int64
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM urls").Scan(&count); err != nil {
		return "", fmt.Errorf("count urls: %w", err)
	}

	code, err := shortcode.Generate(count)
	if err != nil {
		return "", fmt.Errorf("generate short code: %w", err)
	}

	now := time.Now().UTC()
	expiresAt := now.AddDate(0, 0, s.expireDays)

	_, err = s.db.ExecContext(ctx,
		"INSERT INTO urls (short_code, long_url, created_at, expires_at) VALUES (@p1, @p2, @p3, @p4)",
		code, longURL, now, expiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("insert url: %w", err)
	}

	return code, nil
}

// SQL Server error 2627 is a primary key violation, 2601 a duplicate key in
// a unique index; either means another request already claimed this code.
func isDuplicateShortCode(err error) bool {
	var sqlErr mssql.Error
	return errors.As(err, &sqlErr) && (sqlErr.Number == 2627 || sqlErr.Number == 2601)
}

func (s *Store) GetLongURL(ctx context.Context, shortCode string) (string, bool, error) {
	var longURL string
	var expiresAt time.Time

	err := s.db.QueryRowContext(ctx,
		"SELECT long_url, expires_at FROM urls WHERE short_code = @p1", shortCode,
	).Scan(&longURL, &expiresAt)

	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("select url: %w", err)
	}
	if expiresAt.Before(time.Now().UTC()) {
		return "", false, nil
	}

	return longURL, true, nil
}

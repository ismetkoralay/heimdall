// Package sql manages db connection and migrations
package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"

	"github.com/ismetkoralay/heimdall/internal/config"
)

var (
	ErrConnectingDatabase = errors.New("error creating db connection")
	ErrApplyingMigrations = errors.New("error applying db migrations")
)

// New opens a connection pool and, if configured, applies pending migrations.
// sql.Open does not dial the database eagerly, so when ApplyMigrations is
// false, a bad host/port/credentials won't surface here — only on the first
// real query. This keeps New usable without a live database in unit tests;
// TestNewIntegration is the one place connectivity is actually exercised.
func New(ctx context.Context, dbConfig config.Database) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.Name, dbConfig.SSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrConnectingDatabase, err)
	}

	if dbConfig.ApplyMigrations {
		if err := goose.UpContext(ctx, db, dbConfig.MigrationsDir); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrApplyingMigrations, err)
		}
	}

	return db, nil
}

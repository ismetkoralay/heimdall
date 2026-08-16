// Package repository manages auth related stores.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ismetkoralay/heimdall/internal/auth"
)

type Repository struct {
	db *sql.DB
}

func New(
	db *sql.DB,
) *Repository {
	return &Repository{
		db: db,
	}
}

const insertAPIKeyQuery = "insert into api_keys(hashed_key, name, rpm_limit, daily_token_quota, revoked, created_at) values ($1, $2, $3, $4, $5, $6) " +
	"returning id, hashed_key, name, rpm_limit, daily_token_quota, revoked, created_at"

var ErrInsertingAPIKeyQuery = errors.New("error inserting new api key")

// CreateAPIKey insert a new api_key record.
func (r *Repository) CreateAPIKey(
	ctx context.Context,
	apiKey auth.APIKey,
) (auth.APIKey, error) {
	row := r.db.QueryRowContext(ctx, insertAPIKeyQuery, apiKey.HashedKey, apiKey.Name, apiKey.RPMLimit, apiKey.DailyTokenQuota, apiKey.Revoked, apiKey.CreatedAt)

	var (
		apiKeyID        string
		hashedKey       string
		name            string
		rpmLimit        int
		dailyTokenQuota int
		revoked         bool
		createdAt       time.Time
	)
	if err := row.Scan(&apiKeyID, &hashedKey, &name, &rpmLimit, &dailyTokenQuota, &revoked, &createdAt); err != nil {
		return auth.APIKey{}, fmt.Errorf("%w: %w", ErrInsertingAPIKeyQuery, err)
	}

	return auth.APIKey{
		ID:              apiKeyID,
		HashedKey:       hashedKey,
		Name:            name,
		RPMLimit:        rpmLimit,
		DailyTokenQuota: dailyTokenQuota,
		Revoked:         revoked,
		CreatedAt:       createdAt,
	}, nil
}

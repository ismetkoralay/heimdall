// Package repository manages auth related stores.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ismetkoralay/heimdall/internal/auth"
)

type Repository struct {
	db     *sql.DB
	logger *slog.Logger
}

func New(
	db *sql.DB,
	logger *slog.Logger,
) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
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

const getAPIKeyByHashedKeyQuery = "select id, hashed_key, name, rpm_limit, daily_token_quota, revoked, created_at from api_keys where hashed_key = $1"

var ErrGettingAPIKeyByHashedKey = errors.New("error getting api key by hashed key")

func (r *Repository) GetAPIKeyByHashedKey(ctx context.Context, hashedKey string) (auth.APIKey, error) {
	row := r.db.QueryRowContext(ctx, getAPIKeyByHashedKeyQuery, hashedKey)

	res, err := scanAPIKey(row)
	if errors.Is(err, sql.ErrNoRows) {
		return auth.APIKey{}, fmt.Errorf("%w: %w", ErrGettingAPIKeyByHashedKey, &auth.AuthError{
			StatusCode: http.StatusUnauthorized,
			Err:        errors.New("key not found"),
		})
	} else if err != nil {
		r.logger.ErrorContext(ctx, "failed to get api key by hashed key", "error", err)
		return auth.APIKey{}, fmt.Errorf("%w: %w", ErrGettingAPIKeyByHashedKey, &auth.AuthError{
			StatusCode: http.StatusInternalServerError,
			Err:        errors.New("something went wrong"),
		})
	}

	return res, nil
}

func scanAPIKey(row *sql.Row) (auth.APIKey, error) {
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
		return auth.APIKey{}, err
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

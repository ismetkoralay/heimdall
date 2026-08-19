package repository

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"github.com/ismetkoralay/heimdall/internal/auth"
	"github.com/ismetkoralay/heimdall/internal/config"
	"github.com/ismetkoralay/heimdall/internal/sql"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestCreateAPIKey(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)

	ctx := context.Background()
	pgContainer, err := postgres.Run(ctx, "postgres:17-alpine",
		postgres.WithDatabase("heimdall"),
		postgres.WithUsername("heimdall"),
		postgres.WithPassword("password"),
		postgres.BasicWaitStrategies(),
	)
	t.Cleanup(func() {
		assert.NoError(t, testcontainers.TerminateContainer(pgContainer))
	})
	assert.NoError(t, err)

	host, err := pgContainer.Host(ctx)
	assert.NoError(t, err)
	port, err := pgContainer.MappedPort(ctx, "5432/tcp")
	assert.NoError(t, err)

	dbConfig := config.Database{
		Host:            host,
		Port:            port.Port(),
		User:            "heimdall",
		Password:        "password",
		Name:            "heimdall",
		SSLMode:         "disable",
		ApplyMigrations: true,
		MigrationsDir:   "../../../migrations",
	}

	db, err := sql.New(ctx, dbConfig)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	repo := New(db, slog.New(slog.NewTextHandler(io.Discard, nil)))

	var (
		now      = time.Now().UTC().Truncate(time.Microsecond)
		dummyRec = auth.APIKey{
			HashedKey:       "existing-key",
			Name:            "existing-key-name",
			RPMLimit:        1,
			DailyTokenQuota: 2,
			Revoked:         false,
			CreatedAt:       now,
		}
	)

	_, err = repo.CreateAPIKey(ctx, dummyRec)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		input          auth.APIKey
		expectedResult auth.APIKey
		expectedErr    error
	}{
		{
			name: "successful call",
			input: auth.APIKey{
				HashedKey:       "new-key-1",
				Name:            "new-key-1-name",
				RPMLimit:        1,
				DailyTokenQuota: 2,
				Revoked:         false,
				CreatedAt:       now,
			},
			expectedResult: auth.APIKey{
				HashedKey:       "new-key-1",
				Name:            "new-key-1-name",
				RPMLimit:        1,
				DailyTokenQuota: 2,
				Revoked:         false,
				CreatedAt:       now,
			},
		},
		{
			name: "fails if hashed key exists",
			input: auth.APIKey{
				HashedKey: "existing-key",
				Name:      "test",
			},
			expectedErr: ErrInsertingAPIKeyQuery,
		},
		{
			name: "fails if hashed key is empty",
			input: auth.APIKey{
				HashedKey: "",
				Name:      "test",
			},
			expectedErr: ErrInsertingAPIKeyQuery,
		},
		{
			name: "fails if name is empty",
			input: auth.APIKey{
				HashedKey: "new-key-2",
				Name:      "",
			},
			expectedErr: ErrInsertingAPIKeyQuery,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			res, err := repo.CreateAPIKey(ctx, tt.input)

			// Assert
			assert.ErrorIs(t, err, tt.expectedErr)
			assert.Equal(t, tt.expectedResult.HashedKey, res.HashedKey)
			assert.Equal(t, tt.expectedResult.Name, res.Name)
			assert.Equal(t, tt.expectedResult.RPMLimit, res.RPMLimit)
			assert.Equal(t, tt.expectedResult.DailyTokenQuota, res.DailyTokenQuota)
			assert.Equal(t, tt.expectedResult.Revoked, res.Revoked)
			assert.Equal(t, tt.expectedResult.CreatedAt, res.CreatedAt)
		})
	}
}

func TestGetAPIKeyByHashedKey(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)

	ctx := context.Background()
	pgContainer, err := postgres.Run(ctx, "postgres:17-alpine",
		postgres.WithDatabase("heimdall"),
		postgres.WithUsername("heimdall"),
		postgres.WithPassword("password"),
		postgres.BasicWaitStrategies(),
	)
	t.Cleanup(func() {
		assert.NoError(t, testcontainers.TerminateContainer(pgContainer))
	})
	assert.NoError(t, err)

	host, err := pgContainer.Host(ctx)
	assert.NoError(t, err)
	port, err := pgContainer.MappedPort(ctx, "5432/tcp")
	assert.NoError(t, err)

	dbConfig := config.Database{
		Host:            host,
		Port:            port.Port(),
		User:            "heimdall",
		Password:        "password",
		Name:            "heimdall",
		SSLMode:         "disable",
		ApplyMigrations: true,
		MigrationsDir:   "../../../migrations",
	}

	db, err := sql.New(ctx, dbConfig)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	repo := New(db, slog.New(slog.NewTextHandler(io.Discard, nil)))

	var (
		now      = time.Now().UTC().Truncate(time.Microsecond)
		dummyRec = auth.APIKey{
			HashedKey:       "existing-key",
			Name:            "existing-key-name",
			RPMLimit:        1,
			DailyTokenQuota: 2,
			Revoked:         false,
			CreatedAt:       now,
		}
	)

	_, err = repo.CreateAPIKey(ctx, dummyRec)
	assert.NoError(t, err)

	tests := []struct {
		name            string
		input           string
		expectedResult  auth.APIKey
		expectedErr     error
		expectedAuthErr *auth.AuthError
	}{
		{
			name:  "successful call - existing key",
			input: "existing-key",
			expectedResult: auth.APIKey{
				HashedKey:       "existing-key",
				Name:            "existing-key-name",
				RPMLimit:        1,
				DailyTokenQuota: 2,
				Revoked:         false,
				CreatedAt:       now,
			},
		},
		{
			name:           "fails with non-existing key",
			input:          "non-existing-key",
			expectedResult: auth.APIKey{},
			expectedErr:    errors.New("key not found"),
			expectedAuthErr: &auth.AuthError{
				StatusCode: http.StatusUnauthorized,
				Err:        errors.New("key not found"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			res, err := repo.GetAPIKeyByHashedKey(ctx, tt.input)

			// Assert
			assert.Equal(t, tt.expectedResult.HashedKey, res.HashedKey)
			assert.Equal(t, tt.expectedResult.Name, res.Name)
			assert.Equal(t, tt.expectedResult.RPMLimit, res.RPMLimit)
			assert.Equal(t, tt.expectedResult.DailyTokenQuota, res.DailyTokenQuota)
			assert.Equal(t, tt.expectedResult.Revoked, res.Revoked)
			assert.Equal(t, tt.expectedResult.CreatedAt, res.CreatedAt)

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			if tt.expectedAuthErr != nil {
				var authErr *auth.AuthError
				assert.ErrorAs(t, err, &authErr)
				assert.Equal(t, tt.expectedAuthErr.StatusCode, authErr.StatusCode)
			}
		})
	}
}

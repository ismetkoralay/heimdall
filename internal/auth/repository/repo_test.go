package repository

import (
	"context"
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

	repo := New(db)

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

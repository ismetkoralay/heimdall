package sql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/ismetkoralay/heimdall/internal/config"
)

// TestNewIntegration applies the real migrations/ directory to a disposable
// Postgres container started via testcontainers-go. It skips cleanly
// whenever Docker isn't reachable, so `make test` never needs Docker or a
// live database.
func TestNewIntegration(t *testing.T) {
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
		MigrationsDir:   "../../migrations",
	}

	conn, err := New(ctx, dbConfig)
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	t.Cleanup(func() { assert.NoError(t, conn.Close()) })

	var exists bool
	err = conn.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'api_keys')`,
	).Scan(&exists)
	assert.NoError(t, err)
	assert.True(t, exists, "api_keys table should exist after goose applies migrations/")
}

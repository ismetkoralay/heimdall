package sql

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ismetkoralay/heimdall/internal/config"
)

const validMigration = `-- +goose Up
CREATE TABLE probe (id int);
-- +goose Down
DROP TABLE probe;
`

func createDBConfig(migrationsDir string, applyMigrations bool) config.Database {
	return config.Database{
		Host:            "127.0.0.1",
		Port:            "1",
		User:            "user",
		Password:        "password",
		Name:            "db",
		SSLMode:         "disable",
		ApplyMigrations: applyMigrations,
		MigrationsDir:   migrationsDir,
	}
}

func TestNew(t *testing.T) {
	missingDir := filepath.Join(t.TempDir(), "does-not-exist")

	migrationsDir := t.TempDir()
	assert.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "00001_probe.sql"), []byte(validMigration), 0o600))

	tests := []struct {
		name           string
		dbConfig       config.Database
		expectsNilConn bool
		expectedError  error
	}{
		{
			name:     "successful - nonexiting migration folder, not applies migration",
			dbConfig: createDBConfig(missingDir, false),
		},
		{
			name:           "fails with missing migrations folder",
			dbConfig:       createDBConfig(missingDir, true),
			expectsNilConn: true,
			expectedError:  ErrApplyingMigrations,
		},
		{
			name:           "fails - migrations fail because of nonexisting connection",
			dbConfig:       createDBConfig(migrationsDir, true),
			expectsNilConn: true,
			expectedError:  ErrApplyingMigrations,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := New(context.Background(), tt.dbConfig)
			assert.ErrorIs(t, err, tt.expectedError)

			if tt.expectsNilConn {
				assert.Nil(t, conn)
			} else {
				assert.NotNil(t, conn)
				assert.NoError(t, conn.Close())
			}
		})
	}
}

// Package config loads and validates the configuration for heimdall project.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	ErrLoadingPort                = errors.New("error loading port")
	ErrInvalidPort                = errors.New("error invalid port")
	ErrInvalidModelProviderConfig = errors.New("error invalid model provider config")
	ErrMissingDbHost              = errors.New("error missing db host")
	ErrMissingDbPort              = errors.New("error missing db port")
	ErrMissingDbUser              = errors.New("error missing db user")
	ErrMissingDbPassword          = errors.New("error missing db password")
	ErrMissingDbName              = errors.New("error missing db name")
	ErrMissingDbSSLMode           = errors.New("error missing db ssl mode")
	ErrMissingAdminKey            = errors.New("error missing admin key")
)

// Config holds the configuration for the project.
type Config struct {
	Port             string
	ModelProviderMap map[string]ModelProviderConfig
	ProviderMap      map[string]string
	LogLevel         string
	Database         Database
	AdminKey         string
}

type Database struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	ApplyMigrations bool
	MigrationsDir   string
}

// Load loads the configuration from environment variables.
// Sets default values where possible.
func Load() (Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if p, err := strconv.Atoi(port); err != nil {
		return Config{}, fmt.Errorf("%w: %w", ErrLoadingPort, err)
	} else if p < 1 || p > 65535 {
		return Config{}, fmt.Errorf("%w: %q: port must be between 1 and 65535", ErrInvalidPort, port)
	}

	modelConfig, providerMap, err := LoadModelProviderConfig()
	if err != nil {
		return Config{}, fmt.Errorf("%w: %w", ErrInvalidModelProviderConfig, err)
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		return Config{}, ErrMissingDbHost
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		return Config{}, ErrMissingDbPort
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		return Config{}, ErrMissingDbUser
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		return Config{}, ErrMissingDbPassword
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		return Config{}, ErrMissingDbName
	}

	dbSSLMode := os.Getenv("DB_SSL_MODE")
	if strings.TrimSpace(dbSSLMode) == "" {
		dbSSLMode = "disable"
	}

	dbApplyMigrationsEnv := os.Getenv("DB_APPLY_MIGRATIONS")
	dbApplyMigrations, _ := strconv.ParseBool(dbApplyMigrationsEnv)
	if strings.TrimSpace(dbSSLMode) == "" {
		dbApplyMigrations = false
	}

	dbMigrationsDir := os.Getenv("DB_MIGRATIONS_DIR")
	if strings.TrimSpace(dbMigrationsDir) == "" {
		dbMigrationsDir = "migrations"
	}

	adminKey := os.Getenv("HEIMDALL_ADMIN_KEY")
	if adminKey == "" {
		return Config{}, ErrMissingAdminKey
	}

	return Config{
		Port:             port,
		ModelProviderMap: modelConfig,
		ProviderMap:      providerMap,
		LogLevel:         logLevel,
		Database: Database{
			Host:            dbHost,
			Port:            dbPort,
			User:            dbUser,
			Password:        dbPassword,
			Name:            dbName,
			SSLMode:         dbSSLMode,
			ApplyMigrations: dbApplyMigrations,
			MigrationsDir:   dbMigrationsDir,
		},
		AdminKey: adminKey,
	}, nil
}

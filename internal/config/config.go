// Package config loads and validates the configuration for heimdall project.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

var (
	ErrLoadingPort                = errors.New("error loading port")
	ErrInvalidPort                = errors.New("error invalid port")
	ErrInvalidModelProviderConfig = errors.New("error invalid model provider config")
)

// Config holds the configuration for the project.
type Config struct {
	Port             string
	ModelProviderMap map[string]ModelProviderConfig
	ProviderMap      map[string]string
	LogLevel         string
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

	return Config{
		Port:             port,
		ModelProviderMap: modelConfig,
		ProviderMap:      providerMap,
		LogLevel:         logLevel,
	}, nil
}

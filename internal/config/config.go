// Package config loads and validates the configuration for heimdall project.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
)

var (
	ErrLoadingPort                = errors.New("error loading port")
	ErrInvalidPort                = errors.New("error invalid port")
	ErrInvalidOllamaBaseURL       = errors.New("error invalid ollama base url")
	ErrInvalidModelProviderConfig = errors.New("error invalid model provider config")
)

// Config holds the configuration for the project.
type Config struct {
	Port             string
	OllamaBaseURL    string
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

	ollamaBaseURL := os.Getenv("OLLAMA_BASE_URL")
	if ollamaBaseURL == "" {
		ollamaBaseURL = "http://localhost:11434"
	}

	if u, err := url.Parse(ollamaBaseURL); err != nil {
		return Config{}, fmt.Errorf("%w: %w", ErrInvalidOllamaBaseURL, err)
	} else if u.Scheme == "" || u.Host == "" {
		return Config{}, fmt.Errorf("%w: %q: either scheme or host is empty", ErrInvalidOllamaBaseURL, ollamaBaseURL)
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
		OllamaBaseURL:    ollamaBaseURL,
		ModelProviderMap: modelConfig,
		ProviderMap:      providerMap,
		LogLevel:         logLevel,
	}, nil
}

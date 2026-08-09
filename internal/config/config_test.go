package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name           string
		env            map[string]string
		expectedError  error
		expectedConfig Config
	}{
		{
			name: "valid config",
			env: map[string]string{
				"PORT":                 "3000",
				"OLLAMA_BASE_URL":      "http://testollama.com",
				"OLLAMA_DEFAULT_MODEL": "test-model",
				"LOG_LEVEL":            "error",
			},
			expectedError: nil,
			expectedConfig: Config{
				Port:               "3000",
				OllamaBaseURL:      "http://testollama.com",
				OllamaDefaultModel: "test-model",
				LogLevel:           "error",
			},
		},
		{
			name: "invalid port",
			env: map[string]string{
				"PORT": "invalid",
			},
			expectedError:  ErrLoadingPort,
			expectedConfig: Config{},
		},
		{
			name: "invalid ollama base url",
			env: map[string]string{
				"OLLAMA_BASE_URL": "://invalid-url",
			},
			expectedError:  ErrInvalidOllamaBaseURL,
			expectedConfig: Config{},
		},
		{
			name:          "missing env vars",
			env:           map[string]string{},
			expectedError: nil,
			expectedConfig: Config{
				Port:               "8080",
				OllamaBaseURL:      "http://localhost:11434",
				OllamaDefaultModel: "qwen2.5-coder",
				LogLevel:           "info",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			cfg, err := Load()
			assert.Equal(t, tt.expectedConfig, cfg)
			assert.ErrorIsf(t, err, tt.expectedError, "")
		})
	}
}

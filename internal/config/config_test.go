package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name                   string
		env                    map[string]string
		modelConfigPathEnv     string
		modelConfigFiletoWrite string
		modelConfigFileContent string
		expectedError          error
		expectedConfig         Config
	}{
		{
			name: "successful call - valid config",
			env: map[string]string{
				"PORT":            "3000",
				"OLLAMA_BASE_URL": "http://testollama.com",
				"LOG_LEVEL":       "error",
			},
			modelConfigPathEnv:     "models.yaml",
			modelConfigFiletoWrite: "models.yaml",
			modelConfigFileContent: validFileContent,
			expectedError:          nil,
			expectedConfig: Config{
				Port:          "3000",
				OllamaBaseURL: "http://testollama.com",
				LogLevel:      "error",
				ModelProviderMap: map[string]ModelProviderConfig{
					"test_model": {
						ProviderName:    "test_provider",
						ProviderBaseURL: "http://test_provider.com:3000",
						UpstreamModel:   "test_upstream_model",
					},
				},
			},
		},
		{
			name: "fails with invalid port",
			env: map[string]string{
				"PORT": "invalid",
			},
			modelConfigPathEnv:     "models.yaml",
			modelConfigFiletoWrite: "models.yaml",
			modelConfigFileContent: validFileContent,
			expectedError:          ErrLoadingPort,
			expectedConfig:         Config{},
		},
		{
			name: "fails with invalid ollama base url",
			env: map[string]string{
				"OLLAMA_BASE_URL": "://invalid-url",
			},
			modelConfigPathEnv:     "models.yaml",
			modelConfigFiletoWrite: "models.yaml",
			modelConfigFileContent: validFileContent,
			expectedError:          ErrInvalidOllamaBaseURL,
			expectedConfig:         Config{},
		},
		{
			name: "fails if LoadModelProviderConfig fails",
			env: map[string]string{
				"PORT":            "3000",
				"OLLAMA_BASE_URL": "http://testollama.com",
				"LOG_LEVEL":       "error",
			},
			modelConfigPathEnv:     "dummy.yaml",
			modelConfigFiletoWrite: "models.yaml",
			modelConfigFileContent: validFileContent,
			expectedError:          ErrInvalidModelProviderConfig,
			expectedConfig:         Config{},
		},
		{
			name:           "fails with missing env vars",
			env:            map[string]string{},
			expectedError:  ErrInvalidModelProviderConfig,
			expectedConfig: Config{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			dir := t.TempDir()
			if tt.modelConfigPathEnv != "" {
				envVarPath := filepath.Join(dir, tt.modelConfigPathEnv)
				t.Setenv("MODEL_CONFIG_PATH", envVarPath)
			}

			if tt.modelConfigFiletoWrite != "" {
				path := filepath.Join(dir, tt.modelConfigFiletoWrite)
				if err := os.WriteFile(path, []byte(tt.modelConfigFileContent), os.ModePerm); err != nil {
					t.Fatal(err)
				}
			}

			// Act
			cfg, err := Load()

			// Assert
			assert.Equal(t, tt.expectedConfig, cfg)
			assert.ErrorIsf(t, err, tt.expectedError, "")
		})
	}
}

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
				"PORT":                "3000",
				"LOG_LEVEL":           "error",
				"DB_HOST":             "localhost",
				"DB_PORT":             "5432",
				"DB_USER":             "user",
				"DB_PASSWORD":         "password",
				"DB_NAME":             "database",
				"DB_SSL_MODE":         "disable",
				"DB_APPLY_MIGRATIONS": "true",
				"DB_MIGRATIONS_DIR":   "custom_migrations",
			},
			modelConfigPathEnv:     "models.yaml",
			modelConfigFiletoWrite: "models.yaml",
			modelConfigFileContent: validFileContent,
			expectedError:          nil,
			expectedConfig: Config{
				Port:     "3000",
				LogLevel: "error",
				ModelProviderMap: map[string]ModelProviderConfig{
					"test_model": {
						ProviderName:    "test_provider",
						ProviderBaseURL: "http://test_provider.com:3000",
						UpstreamModel:   "test_upstream_model",
					},
				},
				ProviderMap: map[string]string{
					"test_provider": "http://test_provider.com:3000",
				},
				Database: Database{
					Host:            "localhost",
					Port:            "5432",
					User:            "user",
					Password:        "password",
					Name:            "database",
					SSLMode:         "disable",
					ApplyMigrations: true,
					MigrationsDir:   "custom_migrations",
				},
			},
		},
		{
			name: "successful call - with db vars and model config path env var",
			env: map[string]string{
				"DB_HOST":             "localhost",
				"DB_PORT":             "5432",
				"DB_USER":             "user",
				"DB_PASSWORD":         "password",
				"DB_NAME":             "database",
				"DB_SSL_MODE":         "disable",
				"DB_APPLY_MIGRATIONS": "false",
			},
			modelConfigPathEnv:     "models.yaml",
			modelConfigFiletoWrite: "models.yaml",
			modelConfigFileContent: validFileContent,
			expectedError:          nil,
			expectedConfig: Config{
				Port:     "8080",
				LogLevel: "info",
				ModelProviderMap: map[string]ModelProviderConfig{
					"test_model": {
						ProviderName:    "test_provider",
						ProviderBaseURL: "http://test_provider.com:3000",
						UpstreamModel:   "test_upstream_model",
					},
				},
				ProviderMap: map[string]string{
					"test_provider": "http://test_provider.com:3000",
				},
				Database: Database{
					Host:            "localhost",
					Port:            "5432",
					User:            "user",
					Password:        "password",
					Name:            "database",
					SSLMode:         "disable",
					ApplyMigrations: false,
					MigrationsDir:   "migrations",
				},
			},
		},
		{
			name: "successful call - with missing db ssl mode, fallbacks to disable",
			env: map[string]string{
				"PORT":        "3000",
				"LOG_LEVEL":   "error",
				"DB_HOST":     "localhost",
				"DB_PORT":     "5432",
				"DB_USER":     "user",
				"DB_PASSWORD": "password",
				"DB_NAME":     "database",
			},
			modelConfigPathEnv:     "models.yaml",
			modelConfigFiletoWrite: "models.yaml",
			modelConfigFileContent: validFileContent,
			expectedError:          nil,
			expectedConfig: Config{
				Port:     "3000",
				LogLevel: "error",
				ModelProviderMap: map[string]ModelProviderConfig{
					"test_model": {
						ProviderName:    "test_provider",
						ProviderBaseURL: "http://test_provider.com:3000",
						UpstreamModel:   "test_upstream_model",
					},
				},
				ProviderMap: map[string]string{
					"test_provider": "http://test_provider.com:3000",
				},
				Database: Database{
					Host:          "localhost",
					Port:          "5432",
					User:          "user",
					Password:      "password",
					Name:          "database",
					SSLMode:       "disable",
					MigrationsDir: "migrations",
				},
			},
		},
		{
			name: "successful call - with missing db apply migrations, fallbacks to false",
			env: map[string]string{
				"PORT":        "3000",
				"LOG_LEVEL":   "error",
				"DB_HOST":     "localhost",
				"DB_PORT":     "5432",
				"DB_USER":     "user",
				"DB_PASSWORD": "password",
				"DB_NAME":     "database",
				"DB_SSL_MODE": "disable",
			},
			modelConfigPathEnv:     "models.yaml",
			modelConfigFiletoWrite: "models.yaml",
			modelConfigFileContent: validFileContent,
			expectedError:          nil,
			expectedConfig: Config{
				Port:     "3000",
				LogLevel: "error",
				ModelProviderMap: map[string]ModelProviderConfig{
					"test_model": {
						ProviderName:    "test_provider",
						ProviderBaseURL: "http://test_provider.com:3000",
						UpstreamModel:   "test_upstream_model",
					},
				},
				ProviderMap: map[string]string{
					"test_provider": "http://test_provider.com:3000",
				},
				Database: Database{
					Host:            "localhost",
					Port:            "5432",
					User:            "user",
					Password:        "password",
					Name:            "database",
					SSLMode:         "disable",
					ApplyMigrations: false,
					MigrationsDir:   "migrations",
				},
			},
		},
		{
			name: "successful call - with non-boolean db apply migrations, fallback to false",
			env: map[string]string{
				"PORT":                "3000",
				"LOG_LEVEL":           "error",
				"DB_HOST":             "localhost",
				"DB_PORT":             "5432",
				"DB_USER":             "user",
				"DB_PASSWORD":         "password",
				"DB_NAME":             "database",
				"DB_SSL_MODE":         "disable",
				"DB_APPLY_MIGRATIONS": "dummy",
			},
			modelConfigPathEnv:     "models.yaml",
			modelConfigFiletoWrite: "models.yaml",
			modelConfigFileContent: validFileContent,
			expectedError:          nil,
			expectedConfig: Config{
				Port:     "3000",
				LogLevel: "error",
				ModelProviderMap: map[string]ModelProviderConfig{
					"test_model": {
						ProviderName:    "test_provider",
						ProviderBaseURL: "http://test_provider.com:3000",
						UpstreamModel:   "test_upstream_model",
					},
				},
				ProviderMap: map[string]string{
					"test_provider": "http://test_provider.com:3000",
				},
				Database: Database{
					Host:            "localhost",
					Port:            "5432",
					User:            "user",
					Password:        "password",
					Name:            "database",
					SSLMode:         "disable",
					ApplyMigrations: false,
					MigrationsDir:   "migrations",
				},
			},
		},
		{
			name: "fails if db host missing",
			env: map[string]string{
				"PORT":        "3000",
				"LOG_LEVEL":   "error",
				"DB_PORT":     "5432",
				"DB_USER":     "user",
				"DB_PASSWORD": "password",
				"DB_NAME":     "database",
				"DB_SSL_MODE": "disable",
			},
			modelConfigPathEnv:     "models.yaml",
			modelConfigFiletoWrite: "models.yaml",
			modelConfigFileContent: validFileContent,
			expectedError:          ErrMissingDbHost,
			expectedConfig:         Config{},
		},
		{
			name: "fails if db port missing",
			env: map[string]string{
				"PORT":        "3000",
				"LOG_LEVEL":   "error",
				"DB_HOST":     "localhost",
				"DB_USER":     "user",
				"DB_PASSWORD": "password",
				"DB_NAME":     "database",
				"DB_SSL_MODE": "disable",
			},
			modelConfigPathEnv:     "models.yaml",
			modelConfigFiletoWrite: "models.yaml",
			modelConfigFileContent: validFileContent,
			expectedError:          ErrMissingDbPort,
			expectedConfig:         Config{},
		},
		{
			name: "fails if db user missing",
			env: map[string]string{
				"PORT":        "3000",
				"LOG_LEVEL":   "error",
				"DB_HOST":     "localhost",
				"DB_PORT":     "5432",
				"DB_PASSWORD": "password",
				"DB_NAME":     "database",
				"DB_SSL_MODE": "disable",
			},
			modelConfigPathEnv:     "models.yaml",
			modelConfigFiletoWrite: "models.yaml",
			modelConfigFileContent: validFileContent,
			expectedError:          ErrMissingDbUser,
			expectedConfig:         Config{},
		},
		{
			name: "fails if db password missing",
			env: map[string]string{
				"PORT":        "3000",
				"LOG_LEVEL":   "error",
				"DB_HOST":     "localhost",
				"DB_PORT":     "5432",
				"DB_USER":     "user",
				"DB_NAME":     "database",
				"DB_SSL_MODE": "disable",
			},
			modelConfigPathEnv:     "models.yaml",
			modelConfigFiletoWrite: "models.yaml",
			modelConfigFileContent: validFileContent,
			expectedError:          ErrMissingDbPassword,
			expectedConfig:         Config{},
		},
		{
			name: "fails if db name missing",
			env: map[string]string{
				"PORT":        "3000",
				"LOG_LEVEL":   "error",
				"DB_HOST":     "localhost",
				"DB_PORT":     "5432",
				"DB_USER":     "user",
				"DB_PASSWORD": "password",
				"DB_SSL_MODE": "disable",
			},
			modelConfigPathEnv:     "models.yaml",
			modelConfigFiletoWrite: "models.yaml",
			modelConfigFileContent: validFileContent,
			expectedError:          ErrMissingDbName,
			expectedConfig:         Config{},
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

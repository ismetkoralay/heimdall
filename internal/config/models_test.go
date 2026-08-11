package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	validFileContent = `providers:
  - name: test_provider
    base_url: http://test_provider.com:3000
models:
  - name: test_model
    provider: test_provider
    upstream_model: test_upstream_model
`
	invalidFileContent = `providers
  - name: test
    base_uri: http://test.com:1234
models:
  - name: test_model
    provider: test
	upstream_model: test_upstream_model`
)

func TestLoadModelProviderConfig(t *testing.T) {
	tests := []struct {
		name          string
		configPathEnv string
		fileToWrite   string
		fileContent   string
		expectedRes   map[string]ModelProviderConfig
		expectedErr   error
	}{
		{
			name:          "successful call",
			configPathEnv: "models.yaml",
			fileToWrite:   "models.yaml",
			fileContent:   validFileContent,
			expectedRes: map[string]ModelProviderConfig{
				"test_model": {
					ProviderName:    "test_provider",
					ProviderBaseURL: "http://test_provider.com:3000",
					UpstreamModel:   "test_upstream_model",
				},
			},
			expectedErr: nil,
		},
		{
			name:        "fails if env is missing the model config map path",
			fileToWrite: "models.yaml",
			fileContent: validFileContent,
			expectedRes: nil,
			expectedErr: ErrMissingModelConfigMapEnvVar,
		},
		{
			name:          "fails if file doesn't exists in the path",
			configPathEnv: "wrong.yaml",
			fileToWrite:   "models.yaml",
			fileContent:   "dummy",
			expectedRes:   nil,
			expectedErr:   ErrReadingModelConfigFile,
		},
		{
			name:          "fails if the file content is not valid yaml",
			configPathEnv: "models.yaml",
			fileToWrite:   "models.yaml",
			fileContent:   "dummy",
			expectedRes:   nil,
			expectedErr:   ErrUnmarshalingModelConfigFile,
		},
		{
			name:          "fails if the file content is not in the exptected structure",
			configPathEnv: "models.yaml",
			fileToWrite:   "models.yaml",
			fileContent:   invalidFileContent,
			expectedRes:   nil,
			expectedErr:   ErrUnmarshalingModelConfigFile,
		},
		{
			name:          "fails if a provider is defined more than once",
			configPathEnv: "models.yaml",
			fileToWrite:   "models.yaml",
			fileContent: `providers:
  - name: test
    base_url: http://test.com:1234
  - name: test
    base_url: http://test.com:1234`,
			expectedRes: nil,
			expectedErr: ErrProviderDefinedManyTimes,
		},
		{
			name:          "fails if the provider base url is invalid",
			configPathEnv: "models.yaml",
			fileToWrite:   "models.yaml",
			fileContent: `providers:
  - name: test
    base_url: bad-url`,
			expectedRes: nil,
			expectedErr: ErrInvalidProviderURL,
		},
		{
			name:          "fails if a model is defined more than once",
			configPathEnv: "models.yaml",
			fileToWrite:   "models.yaml",
			fileContent: `providers:
  - name: test
    base_url: http://test.com:1234
models:
  - name: test_model
    provider: test
    upstream_model: test
  - name: test_model`,
			expectedRes: nil,
			expectedErr: ErrModelDefinedManyTimes,
		},
		{
			name:          "fails if the provider is not listed in the providers",
			configPathEnv: "models.yaml",
			fileToWrite:   "models.yaml",
			fileContent: `providers:
  - name: test_provider
    base_url: http://test_provider.com:3000
models:
  - name: test_model
    provider: none
    upstream_model: test_upstream_model`,
			expectedRes: nil,
			expectedErr: ErrUnknownProvider,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if tt.configPathEnv != "" {
				envVarPath := filepath.Join(dir, tt.configPathEnv)
				t.Setenv("MODEL_CONFIG_PATH", envVarPath)
			}

			path := filepath.Join(dir, tt.fileToWrite)
			if err := os.WriteFile(path, []byte(tt.fileContent), os.ModePerm); err != nil {
				t.Fatal(err)
			}

			res, err := LoadModelProviderConfig()
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedRes, res)
		})
	}
}

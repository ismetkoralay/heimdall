package provider

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProviderError(t *testing.T) {
	tests := []struct {
		name       string
		provider   string
		statusCode int
		retryable  bool
		err        error
	}{
		{
			name:       "Error() returns underlying error message",
			provider:   "TestProvider",
			statusCode: 500,
			retryable:  true,
			err:        errors.New("test error"),
		},
		{
			name:       "Error() returns empty string when underlying error is nil",
			provider:   "TestProvider",
			statusCode: 500,
			retryable:  true,
			err:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providerErr := &ProviderError{
				Provider:   tt.provider,
				StatusCode: tt.statusCode,
				Retryable:  tt.retryable,
				Err:        tt.err,
			}

			err := providerErr.Error()
			if tt.err != nil {
				assert.Equal(t, tt.err.Error(), err)
			} else {
				assert.Equal(t, "", err)
			}

			assert.Equal(t, tt.provider, providerErr.Provider)
			assert.Equal(t, tt.statusCode, providerErr.StatusCode)
			assert.Equal(t, tt.retryable, providerErr.IsRetryable())
		})
	}
}

type fakeUpstreamError struct{ msg string }

func (e *fakeUpstreamError) Error() string { return e.msg }

func TestProviderErrorUnwrap(t *testing.T) {
	t.Run("errors.Is sees through to a wrapped sentinel", func(t *testing.T) {
		sentinel := errors.New("model not found")
		err := &ProviderError{Provider: "ollama", StatusCode: 404, Err: sentinel}

		assert.ErrorIs(t, err, sentinel)
	})

	t.Run("errors.As finds a wrapped concrete error type", func(t *testing.T) {
		wrapped := &fakeUpstreamError{msg: "connection refused"}
		err := &ProviderError{Provider: "ollama", StatusCode: 502, Retryable: true, Err: wrapped}

		var target *fakeUpstreamError
		assert.ErrorAs(t, err, &target)
		assert.Same(t, wrapped, target)
	})
}

package fake

import (
	"context"
	"net/http"
	"testing"

	"github.com/ismetkoralay/heimdall/internal/provider"
	"github.com/stretchr/testify/assert"
)

func TestChat(t *testing.T) {
	response := CreateFakeResponse()

	tests := []struct {
		name              string
		returnErr         bool
		expectedResponse  provider.ChatResponse
		expectErr         bool
		expectedStatus    int
		expectedRetryable bool
	}{
		{
			name:             "successful call returns dummy reponse",
			returnErr:        false,
			expectedResponse: response,
		},
		{
			name:              "returns error when requested",
			returnErr:         true,
			expectedResponse:  provider.ChatResponse{},
			expectErr:         true,
			expectedStatus:    http.StatusBadRequest,
			expectedRetryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			fp := NewFakeProvider(tt.returnErr)

			// Act
			res, err := fp.Chat(context.Background(), provider.ChatRequest{})

			// Assert
			assert.Equal(t, tt.expectedResponse, res)

			if !tt.expectErr {
				assert.NoError(t, err)
				return
			}

			assert.ErrorIs(t, err, ErrFakeProvider)

			var providerErr *provider.ProviderError
			if assert.ErrorAs(t, err, &providerErr) {
				assert.Equal(t, "fake", providerErr.Provider)
				assert.Equal(t, tt.expectedStatus, providerErr.StatusCode)
				assert.Equal(t, tt.expectedRetryable, providerErr.Retryable)
			}
		})
	}
}

func TestSetBaseURL(t *testing.T) {
	// Arrange
	fp := NewFakeProvider(false)

	// Act
	fp.SetBaseURL("test")

	// Assert
	assert.Equal(t, "test", fp.baseURL)
}

func TestName(t *testing.T) {
	// Arrange
	fp := NewFakeProvider(false)

	// Act
	res := fp.Name()

	// Assert
	assert.Equal(t, "fake", res)
}

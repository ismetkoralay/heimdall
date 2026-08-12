package provider

import (
	"context"
	"iter"
	"testing"

	"github.com/ismetkoralay/heimdall/internal/config"
	"github.com/stretchr/testify/assert"
)

type DummyProvider struct{}

func (*DummyProvider) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	return ChatResponse{}, nil
}

func (*DummyProvider) ChatStream(ctx context.Context, req ChatRequest) iter.Seq2[Chunk, error] {
	return nil
}

func (*DummyProvider) Embed(ctx context.Context, req EmbedRequest) (EmbedResponse, error) {
	return EmbedResponse{}, nil
}

func (*DummyProvider) SetBaseURL(baseURL string) {}

func (*DummyProvider) Name() string {
	return ""
}

func TestRoute(t *testing.T) {
	tests := []struct {
		name                  string
		givenModelMap         map[string]config.ModelProviderConfig
		givenProviders        map[string]Provider
		requestedModel        string
		expectedProvider      Provider
		expectedUpstreamModel string
		expectedErr           error
	}{
		{
			name: "successful call",
			givenModelMap: map[string]config.ModelProviderConfig{
				"test-model": {
					ProviderName:    "test-provider",
					ProviderBaseURL: "base-url",
					UpstreamModel:   "test-model",
				},
			},
			givenProviders: map[string]Provider{
				"test-provider": &DummyProvider{},
			},
			requestedModel:        "test-model",
			expectedProvider:      &DummyProvider{},
			expectedUpstreamModel: "test-model",
			expectedErr:           nil,
		},
		{
			name: "fails if the requested model is unknown",
			givenModelMap: map[string]config.ModelProviderConfig{
				"test-model": {
					ProviderName:    "test-provider",
					ProviderBaseURL: "base-url",
					UpstreamModel:   "test-model",
				},
			},
			givenProviders: map[string]Provider{
				"test-provider": &DummyProvider{},
			},
			requestedModel:   "none",
			expectedProvider: nil,
			expectedErr:      ErrUnknownModel,
		},
		{
			name: "fails if the requested model's provider is unknown",
			givenModelMap: map[string]config.ModelProviderConfig{
				"test-model": {
					ProviderName:    "none",
					ProviderBaseURL: "base-url",
					UpstreamModel:   "test-model",
				},
			},
			givenProviders: map[string]Provider{
				"test-provider": &DummyProvider{},
			},
			requestedModel:   "test-model",
			expectedProvider: nil,
			expectedErr:      ErrUnknownProvider,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			router := NewRouter(tt.givenModelMap, tt.givenProviders)

			// Act
			res, uM, err := router.Resolve(tt.requestedModel)

			// Assert
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedProvider, res)
			assert.Equal(t, tt.expectedUpstreamModel, uM)
		})
	}
}

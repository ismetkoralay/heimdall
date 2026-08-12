// Package fake defines a fake provider. It'll be useful for CI and testing.
package fake

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"net/http"

	"github.com/ismetkoralay/heimdall/internal/provider"
)

var ErrFakeProvider = errors.New("error from fake provider")

type FakeProvider struct {
	returnErr bool
	baseURL   string
}

func NewFakeProvider(
	returnErr bool,
) *FakeProvider {
	return &FakeProvider{
		returnErr: returnErr,
	}
}

func CreateFakeResponse() provider.ChatResponse {
	return provider.ChatResponse{
		Model:        "fake-model",
		Content:      "fake-content",
		FinishReason: "done",
		Usage: provider.Usage{
			PromptTokens:     1,
			CompletionTokens: 2,
			TotalTokens:      3,
		},
	}
}

func (p *FakeProvider) Chat(ctx context.Context, request provider.ChatRequest) (provider.ChatResponse, error) {
	if p.returnErr {
		return provider.ChatResponse{}, fmt.Errorf("%w", &provider.ProviderError{
			Provider:   "fake",
			StatusCode: http.StatusBadRequest,
			Retryable:  false,
			Err:        ErrFakeProvider,
		})
	}

	return CreateFakeResponse(), nil
}

func (p *FakeProvider) ChatStream(ctx context.Context, request provider.ChatRequest) iter.Seq2[provider.Chunk, error] {
	return nil
}

func (p *FakeProvider) Embed(ctx context.Context, request provider.EmbedRequest) (provider.EmbedResponse, error) {
	return provider.EmbedResponse{}, nil
}

func (p *FakeProvider) SetBaseURL(baseURL string) {
	p.baseURL = baseURL
}

func (p *FakeProvider) Name() string {
	return "fake"
}

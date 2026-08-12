// Package ollama provides an implementation of the Ollama provider for the internal/provider package.
package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"net/http"
	"net/url"

	"github.com/ismetkoralay/heimdall/internal/provider"
)

var ErrProviderFailed = errors.New("error making ollama request")

type OllamaProvider struct {
	httpClient *http.Client
	baseURL    string
}

func NewOllamaProvider(
	client *http.Client,
) *OllamaProvider {
	return &OllamaProvider{
		httpClient: client,
	}
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaChatRequestOptions struct {
	Temperature *float64 `json:"temperature,omitempty"`
	MaxTokens   *int     `json:"num_predict,omitempty"`
}

type ollamaChatRequest struct {
	Model string `json:"model"`
	// Messages is the ordered conversation history, including any system prompt.
	Messages []ollamaMessage           `json:"messages"`
	Options  *ollamaChatRequestOptions `json:"options,omitempty"`
	Stream   bool                      `json:"stream"`
}

type ollamaChatResponse struct {
	Model         string        `json:"model"`
	CreatedAt     string        `json:"created_at"`
	Message       ollamaMessage `json:"message"`
	Done          bool          `json:"done"`
	DoneReason    string        `json:"done_reason"`
	TotalDuration int           `json:"total_duration"`
	// LoadDuration is the time taken to load the model, in nanoseconds.
	LoadDuration int `json:"load_duration"`
	// Number of tokens in the prompt
	PromptEvalCount int `json:"prompt_eval_count"`
	// Time spent evaluating the prompt in nanoseconds
	PromptEvalDuration int `json:"prompt_eval_duration"`
	// Number of tokens generated in the response
	EvalCount int `json:"eval_count"`
	// Time spent generating tokens in nanoseconds
	EvalDuration int `json:"eval_duration"`
}

type ollamaErrorResponse struct {
	Error string `json:"error"`
}

func mapProviderRole(role provider.Role) string {
	var res string
	switch role {
	case provider.RoleSystem:
		res = "system"
	case provider.RoleAssistant:
		res = "assistant"
	case provider.RoleDeveloper:
		res = "developer"
	case provider.RoleUser:
		res = "user"
	default:
		res = "user"
	}

	return res
}

func (p *OllamaProvider) Chat(ctx context.Context, request provider.ChatRequest) (provider.ChatResponse, error) {
	ollamaMessages := make([]ollamaMessage, len(request.Messages))
	for i, msg := range request.Messages {
		ollamaMessages[i] = ollamaMessage{
			Role:    mapProviderRole(msg.Role),
			Content: msg.Content,
		}
	}
	inputJSON, err := json.Marshal(ollamaChatRequest{
		Model:    request.Model,
		Messages: ollamaMessages,
		Options: &ollamaChatRequestOptions{
			Temperature: request.Temperature,
			MaxTokens:   request.MaxTokens,
		},
		Stream: false,
	})
	if err != nil {
		return provider.ChatResponse{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	jsonBody := bytes.NewReader(inputJSON)

	reqURL, err := url.JoinPath(p.baseURL, "/api/chat")
	if err != nil {
		return provider.ChatResponse{}, fmt.Errorf("failed to build request url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, jsonBody)
	if err != nil {
		return provider.ChatResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")

	var providerError provider.ProviderError
	res, err := p.httpClient.Do(req)
	if err != nil {
		providerError = provider.ProviderError{
			Provider:  "ollama",
			Retryable: true,
			Err:       err,
		}
		return provider.ChatResponse{}, fmt.Errorf("%w: %w", ErrProviderFailed, &providerError)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		retryable := res.StatusCode >= http.StatusInternalServerError

		var ollamaErrorResponse ollamaErrorResponse
		if err := json.NewDecoder(res.Body).Decode(&ollamaErrorResponse); err != nil {
			providerError = provider.ProviderError{
				Provider:   "ollama",
				StatusCode: res.StatusCode,
				Retryable:  retryable,
				Err:        fmt.Errorf("error decoding error response: %w", err),
			}
			return provider.ChatResponse{}, fmt.Errorf("%w: %w", ErrProviderFailed, &providerError)
		}
		providerError = provider.ProviderError{
			Provider:   "ollama",
			StatusCode: res.StatusCode,
			Retryable:  retryable,
			Err:        fmt.Errorf("error returned from ollama: %s", ollamaErrorResponse.Error),
		}
		return provider.ChatResponse{}, fmt.Errorf("%w: %w", ErrProviderFailed, &providerError)
	}

	var ollamaResponse ollamaChatResponse
	if err := json.NewDecoder(res.Body).Decode(&ollamaResponse); err != nil {
		return provider.ChatResponse{}, fmt.Errorf("error decoding response: %w", err)
	}

	return provider.ChatResponse{
		Model:        ollamaResponse.Model,
		Content:      ollamaResponse.Message.Content,
		FinishReason: ollamaResponse.DoneReason,
		Usage: provider.Usage{
			PromptTokens:     ollamaResponse.PromptEvalCount,
			CompletionTokens: ollamaResponse.EvalCount,
			TotalTokens:      ollamaResponse.PromptEvalCount + ollamaResponse.EvalCount,
		},
	}, nil
}

func (p *OllamaProvider) ChatStream(ctx context.Context, request provider.ChatRequest) iter.Seq2[provider.Chunk, error] {
	return nil
}

func (p *OllamaProvider) Embed(ctx context.Context, request provider.EmbedRequest) (provider.EmbedResponse, error) {
	return provider.EmbedResponse{}, nil
}

func (p *OllamaProvider) SetBaseURL(baseURL string) {
	p.baseURL = baseURL
}

func (p *OllamaProvider) Name() string {
	return "ollama"
}

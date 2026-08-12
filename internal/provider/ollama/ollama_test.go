package ollama

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ismetkoralay/heimdall/internal/provider"
	"github.com/stretchr/testify/assert"
)

// errRoundTripper simulates a transport-level failure (e.g. connection refused)
// without needing a real network dial.
type errRoundTripper struct{ err error }

func (rt errRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, rt.err
}

var (
	badTemp = math.Inf(1)
)

func TestMapProviderRole(t *testing.T) {
	tests := []struct {
		name     string
		given    provider.Role
		expected string
	}{
		{
			name:     "maps system role",
			given:    provider.RoleSystem,
			expected: "system",
		},
		{
			name:     "maps assistant role",
			given:    provider.RoleAssistant,
			expected: "assistant",
		},
		{
			name:     "maps user role",
			given:    provider.RoleUser,
			expected: "user",
		},
		{
			name:     "maps developer role",
			given:    provider.RoleDeveloper,
			expected: "developer",
		},
		{
			name:     "falls back to user for unknown role",
			given:    provider.RoleUnkown,
			expected: "user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			res := mapProviderRole(tt.given)

			// Assert
			assert.Equal(t, tt.expected, res)
		})
	}
}

func TestChat(t *testing.T) {
	tests := []struct {
		name                string
		ctx                 context.Context
		roundTripper        *errRoundTripper
		baseURL             string
		ollamaRequest       ollamaChatRequest
		ollamaResponse      ollamaChatResponse
		ollamaErrorResponse *ollamaErrorResponse
		statusCode          int
		chatRequest         provider.ChatRequest
		chatResponse        provider.ChatResponse
		expectedHttpMethod  string
		expectedPath        string
		expectedContentType string
		err                 error
		expectedProviderErr *provider.ProviderError
	}{
		{
			name: "successful call",
			ctx:  context.Background(),
			ollamaRequest: ollamaChatRequest{
				Model: "test-model",
				Messages: []ollamaMessage{
					{
						Role:    "user",
						Content: "test user content",
					},
				},
				Options: &ollamaChatRequestOptions{
					Temperature: new(1.7),
					MaxTokens:   new(100),
				},
				Stream: false,
			},
			ollamaResponse: ollamaChatResponse{
				Model:     "test-model",
				CreatedAt: "2026-08-10",
				Message: ollamaMessage{
					Role:    "assistant",
					Content: "test content",
				},
				Done:               true,
				DoneReason:         "stop",
				TotalDuration:      10,
				LoadDuration:       20,
				PromptEvalCount:    30,
				PromptEvalDuration: 40,
				EvalCount:          50,
				EvalDuration:       60,
			},
			statusCode: http.StatusOK,
			chatRequest: provider.ChatRequest{
				Model: "test-model",
				Messages: []provider.Message{
					{
						Role:    "user",
						Content: "test user content",
					},
				},
				Temperature: new(1.7),
				MaxTokens:   new(100),
			},
			chatResponse: provider.ChatResponse{
				Model:        "test-model",
				Content:      "test content",
				FinishReason: "stop",
				Usage: provider.Usage{
					PromptTokens:     30,
					CompletionTokens: 50,
					TotalTokens:      80,
				},
			},
			expectedHttpMethod:  http.MethodPost,
			expectedPath:        "/api/chat",
			expectedContentType: "application/json",
			err:                 nil,
		},
		{
			name: "successful call when temperature is not set",
			ctx:  context.Background(),
			ollamaRequest: ollamaChatRequest{
				Model: "test-model",
				Messages: []ollamaMessage{
					{
						Role:    "user",
						Content: "test user content",
					},
				},
				Options: &ollamaChatRequestOptions{
					MaxTokens: new(100),
				},
				Stream: false,
			},
			ollamaResponse: ollamaChatResponse{
				Model:     "test-model",
				CreatedAt: "2026-08-10",
				Message: ollamaMessage{
					Role:    "assistant",
					Content: "test content",
				},
				Done:               true,
				DoneReason:         "stop",
				TotalDuration:      10,
				LoadDuration:       20,
				PromptEvalCount:    30,
				PromptEvalDuration: 40,
				EvalCount:          50,
				EvalDuration:       60,
			},
			statusCode: http.StatusOK,
			chatRequest: provider.ChatRequest{
				Model: "test-model",
				Messages: []provider.Message{
					{
						Role:    "user",
						Content: "test user content",
					},
				},
				MaxTokens: new(100),
			},
			chatResponse: provider.ChatResponse{
				Model:        "test-model",
				Content:      "test content",
				FinishReason: "stop",
				Usage: provider.Usage{
					PromptTokens:     30,
					CompletionTokens: 50,
					TotalTokens:      80,
				},
			},
			expectedHttpMethod:  http.MethodPost,
			expectedPath:        "/api/chat",
			expectedContentType: "application/json",
			err:                 nil,
		},
		{
			name: "successful call when max tokens is not set",
			ctx:  context.Background(),
			ollamaRequest: ollamaChatRequest{
				Model: "test-model",
				Messages: []ollamaMessage{
					{
						Role:    "user",
						Content: "test user content",
					},
				},
				Options: &ollamaChatRequestOptions{
					Temperature: new(1.7),
				},
				Stream: false,
			},
			ollamaResponse: ollamaChatResponse{
				Model:     "test-model",
				CreatedAt: "2026-08-10",
				Message: ollamaMessage{
					Role:    "assistant",
					Content: "test content",
				},
				Done:               true,
				DoneReason:         "stop",
				TotalDuration:      10,
				LoadDuration:       20,
				PromptEvalCount:    30,
				PromptEvalDuration: 40,
				EvalCount:          50,
				EvalDuration:       60,
			},
			statusCode: http.StatusOK,
			chatRequest: provider.ChatRequest{
				Model: "test-model",
				Messages: []provider.Message{
					{
						Role:    "user",
						Content: "test user content",
					},
				},
				Temperature: new(1.7),
			},
			chatResponse: provider.ChatResponse{
				Model:        "test-model",
				Content:      "test content",
				FinishReason: "stop",
				Usage: provider.Usage{
					PromptTokens:     30,
					CompletionTokens: 50,
					TotalTokens:      80,
				},
			},
			expectedHttpMethod:  http.MethodPost,
			expectedPath:        "/api/chat",
			expectedContentType: "application/json",
			err:                 nil,
		},
		{
			name: "successful call when both temperature and max tokens are not set",
			ctx:  context.Background(),
			ollamaRequest: ollamaChatRequest{
				Model: "test-model",
				Messages: []ollamaMessage{
					{
						Role:    "user",
						Content: "test user content",
					},
				},
				Options: &ollamaChatRequestOptions{},
				Stream:  false,
			},
			ollamaResponse: ollamaChatResponse{
				Model:     "test-model",
				CreatedAt: "2026-08-10",
				Message: ollamaMessage{
					Role:    "assistant",
					Content: "test content",
				},
				Done:               true,
				DoneReason:         "stop",
				TotalDuration:      10,
				LoadDuration:       20,
				PromptEvalCount:    30,
				PromptEvalDuration: 40,
				EvalCount:          50,
				EvalDuration:       60,
			},
			statusCode: http.StatusOK,
			chatRequest: provider.ChatRequest{
				Model: "test-model",
				Messages: []provider.Message{
					{
						Role:    "user",
						Content: "test user content",
					},
				},
			},
			chatResponse: provider.ChatResponse{
				Model:        "test-model",
				Content:      "test content",
				FinishReason: "stop",
				Usage: provider.Usage{
					PromptTokens:     30,
					CompletionTokens: 50,
					TotalTokens:      80,
				},
			},
			expectedHttpMethod:  http.MethodPost,
			expectedPath:        "/api/chat",
			expectedContentType: "application/json",
			err:                 nil,
		},
		{
			name:           "fails when bad temperature value is set",
			ctx:            context.Background(),
			ollamaRequest:  ollamaChatRequest{},
			ollamaResponse: ollamaChatResponse{},
			statusCode:     http.StatusOK,
			chatRequest: provider.ChatRequest{
				Model: "test-model",
				Messages: []provider.Message{
					{
						Role:    "user",
						Content: "test user content",
					},
				},
				Temperature: &badTemp,
				MaxTokens:   new(100),
			},
			chatResponse:        provider.ChatResponse{},
			expectedHttpMethod:  "",
			expectedPath:        "",
			expectedContentType: "",
			err:                 errors.New("failed to marshal request"),
		},
		{
			name:           "fails when nil context provided",
			ctx:            nil,
			ollamaRequest:  ollamaChatRequest{},
			ollamaResponse: ollamaChatResponse{},
			statusCode:     http.StatusOK,
			chatRequest: provider.ChatRequest{
				Model: "test-model",
				Messages: []provider.Message{
					{
						Role:    "user",
						Content: "test user content",
					},
				},
				Temperature: new(1.7),
				MaxTokens:   new(100),
			},
			chatResponse:        provider.ChatResponse{},
			expectedHttpMethod:  "",
			expectedPath:        "",
			expectedContentType: "",
			err:                 errors.New("failed to create request"),
		},
		{
			name:          "fails when bad base url provided",
			ctx:           context.Background(),
			baseURL:       ":://bad-url",
			ollamaRequest: ollamaChatRequest{},
			statusCode:    http.StatusOK,
			chatRequest: provider.ChatRequest{
				Model: "test-model",
				Messages: []provider.Message{
					{
						Role:    "user",
						Content: "test user content",
					},
				},
				Temperature: new(1.7),
				MaxTokens:   new(100),
			},
			chatResponse:        provider.ChatResponse{},
			expectedHttpMethod:  "",
			expectedPath:        "",
			expectedContentType: "",
			err:                 errors.New("failed to build request url"),
		},
		{
			name: "fails with connection refused",
			ctx:  context.Background(),
			roundTripper: &errRoundTripper{
				err: errors.New("connection refused"),
			},
			ollamaRequest: ollamaChatRequest{},
			statusCode:    http.StatusOK,
			chatRequest: provider.ChatRequest{
				Model: "test-model",
				Messages: []provider.Message{
					{
						Role:    "user",
						Content: "test user content",
					},
				},
				Temperature: new(1.7),
				MaxTokens:   new(100),
			},
			chatResponse:        provider.ChatResponse{},
			expectedHttpMethod:  "",
			expectedPath:        "",
			expectedContentType: "",
			err:                 errors.New("connection refused"),
			expectedProviderErr: &provider.ProviderError{
				Provider:  "ollama",
				Retryable: true,
			},
		},
		{
			name: "fails with not found status",
			ctx:  context.Background(),
			ollamaRequest: ollamaChatRequest{
				Model: "test-model",
				Messages: []ollamaMessage{
					{
						Role:    "user",
						Content: "test user content",
					},
				},
				Options: &ollamaChatRequestOptions{
					Temperature: new(1.7),
					MaxTokens:   new(100),
				},
			},
			ollamaErrorResponse: &ollamaErrorResponse{
				Error: "fails with not found",
			},
			statusCode: http.StatusNotFound,
			chatRequest: provider.ChatRequest{
				Model: "test-model",
				Messages: []provider.Message{
					{
						Role:    "user",
						Content: "test user content",
					},
				},
				Temperature: new(1.7),
				MaxTokens:   new(100),
			},
			chatResponse:        provider.ChatResponse{},
			expectedHttpMethod:  http.MethodPost,
			expectedPath:        "/api/chat",
			expectedContentType: "application/json",
			err:                 errors.New("fails with not found"),
			expectedProviderErr: &provider.ProviderError{
				Provider:   "ollama",
				StatusCode: http.StatusNotFound,
				Retryable:  false,
			},
		},
		{
			name: "fails with internal server error status",
			ctx:  context.Background(),
			ollamaRequest: ollamaChatRequest{
				Model: "test-model",
				Messages: []ollamaMessage{
					{
						Role:    "user",
						Content: "test user content",
					},
				},
				Options: &ollamaChatRequestOptions{
					Temperature: new(1.7),
					MaxTokens:   new(100),
				},
			},
			ollamaErrorResponse: &ollamaErrorResponse{
				Error: "fails with internal server error",
			},
			statusCode: http.StatusInternalServerError,
			chatRequest: provider.ChatRequest{
				Model: "test-model",
				Messages: []provider.Message{
					{
						Role:    "user",
						Content: "test user content",
					},
				},
				Temperature: new(1.7),
				MaxTokens:   new(100),
			},
			chatResponse:        provider.ChatResponse{},
			expectedHttpMethod:  http.MethodPost,
			expectedPath:        "/api/chat",
			expectedContentType: "application/json",
			err:                 errors.New("fails with internal server error"),
			expectedProviderErr: &provider.ProviderError{
				Provider:   "ollama",
				StatusCode: http.StatusInternalServerError,
				Retryable:  true,
			},
		},
		{
			name: "fails with bad gateway status",
			ctx:  context.Background(),
			ollamaRequest: ollamaChatRequest{
				Model: "test-model",
				Messages: []ollamaMessage{
					{
						Role:    "user",
						Content: "test user content",
					},
				},
				Options: &ollamaChatRequestOptions{
					Temperature: new(1.7),
					MaxTokens:   new(100),
				},
			},
			ollamaErrorResponse: &ollamaErrorResponse{
				Error: "fails with bad gateway",
			},
			statusCode: http.StatusBadGateway,
			chatRequest: provider.ChatRequest{
				Model: "test-model",
				Messages: []provider.Message{
					{
						Role:    "user",
						Content: "test user content",
					},
				},
				Temperature: new(1.7),
				MaxTokens:   new(100),
			},
			chatResponse:        provider.ChatResponse{},
			expectedHttpMethod:  http.MethodPost,
			expectedPath:        "/api/chat",
			expectedContentType: "application/json",
			err:                 errors.New("fails with bad gateway"),
			expectedProviderErr: &provider.ProviderError{
				Provider:   "ollama",
				StatusCode: http.StatusBadGateway,
				Retryable:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			var (
				capturedRequest     ollamaChatRequest
				capturedMethod      string
				capturedPath        string
				capturedContentType string
			)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				err := json.NewDecoder(r.Body).Decode(&capturedRequest)
				assert.NoError(t, err)

				capturedMethod = r.Method
				capturedPath = r.URL.Path
				capturedContentType = r.Header.Get("Content-Type")

				w.WriteHeader(tt.statusCode)
				if tt.ollamaErrorResponse != nil {
					err = json.NewEncoder(w).Encode(*tt.ollamaErrorResponse)
					assert.NoError(t, err)
				} else {
					err = json.NewEncoder(w).Encode(tt.ollamaResponse)
					assert.NoError(t, err)
				}
			}))
			defer srv.Close()

			client := srv.Client()
			if tt.roundTripper != nil {
				client.Transport = tt.roundTripper
			}

			baseURL := srv.URL
			if tt.baseURL != "" {
				baseURL = tt.baseURL
			}
			p := NewOllamaProvider(client)
			p.SetBaseURL(baseURL)

			// Act
			res, err := p.Chat(tt.ctx, tt.chatRequest)

			// Assert
			if tt.err == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.err.Error())
			}

			if tt.expectedProviderErr != nil {
				assert.ErrorIs(t, err, ErrProviderFailed)
				var providerErr *provider.ProviderError
				assert.ErrorAs(t, err, &providerErr)
				assert.Equal(t, tt.expectedProviderErr.Provider, providerErr.Provider)
				assert.Equal(t, tt.expectedProviderErr.StatusCode, providerErr.StatusCode)
				assert.Equal(t, tt.expectedProviderErr.Retryable, providerErr.Retryable)
			}

			assert.EqualValues(t, tt.chatResponse, res)
			assert.Equal(t, tt.expectedHttpMethod, capturedMethod)
			assert.Equal(t, tt.expectedPath, capturedPath)
			assert.Equal(t, tt.expectedContentType, capturedContentType)
			assert.EqualValues(t, tt.ollamaRequest, capturedRequest)
		})
	}
}

func TestOllamaProvider_Chat_InvalidJSONResponse(t *testing.T) {
	// Arrange
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	p := NewOllamaProvider(srv.Client())
	p.SetBaseURL(srv.URL)

	// Act
	_, err := p.Chat(context.Background(), provider.ChatRequest{
		Model: "test-model",
		Messages: []provider.Message{
			{
				Role:    "user",
				Content: "test content",
			},
		},
	})

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error decoding response")
}

func TestOllamaProvider_Chat_InvalidJSONErrorResponse(t *testing.T) {
	// Arrange
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("<html>bad gateway</html>"))
	}))
	defer srv.Close()

	p := NewOllamaProvider(srv.Client())
	p.SetBaseURL(srv.URL)

	// Act
	_, err := p.Chat(context.Background(), provider.ChatRequest{
		Model: "test-model",
		Messages: []provider.Message{
			{
				Role:    "user",
				Content: "test content",
			},
		},
	})

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrProviderFailed)

	var providerErr *provider.ProviderError
	assert.ErrorAs(t, err, &providerErr)
	assert.Equal(t, "ollama", providerErr.Provider)
	assert.Equal(t, http.StatusBadGateway, providerErr.StatusCode)
	assert.True(t, providerErr.Retryable)
}

func TestOllamaProvider_Chat_ContextCanceled(t *testing.T) {
	// Arrange
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ollamaChatResponse{})
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := NewOllamaProvider(srv.Client())
	p.SetBaseURL(srv.URL)

	// Act
	_, err := p.Chat(ctx, provider.ChatRequest{
		Model: "test-model",
		Messages: []provider.Message{
			{
				Role:    "user",
				Content: "test content",
			},
		},
	})

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestSetBaseURL(t *testing.T) {
	// Arrange
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ollamaChatResponse{})
	}))
	defer srv.Close()

	p := NewOllamaProvider(srv.Client())

	// Act
	p.SetBaseURL(srv.URL)

	// Assert
	assert.Equal(t, srv.URL, p.baseURL)
}

func TestName(t *testing.T) {
	// Arrange
	p := NewOllamaProvider(nil)

	// Act
	name := p.Name()

	// Assert
	assert.Equal(t, "ollama", name)
}

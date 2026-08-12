package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ismetkoralay/heimdall/internal/provider"
	"github.com/ismetkoralay/heimdall/internal/provider/fake"
	"github.com/stretchr/testify/assert"
)

type MockProviderRouter struct {
	resolve func(model string) (provider.Provider, string, error)
}

func (m *MockProviderRouter) Resolve(model string) (provider.Provider, string, error) {
	return m.resolve(model)
}

type MockProvider struct {
	chat func(ctx context.Context, req provider.ChatRequest) (provider.ChatResponse, error)
	name func() string
}

func (m *MockProvider) Chat(ctx context.Context, req provider.ChatRequest) (provider.ChatResponse, error) {
	return m.chat(ctx, req)
}

func (m *MockProvider) ChatStream(ctx context.Context, req provider.ChatRequest) iter.Seq2[provider.Chunk, error] {
	return nil
}

func (m *MockProvider) Embed(ctx context.Context, req provider.EmbedRequest) (provider.EmbedResponse, error) {
	return provider.EmbedResponse{}, nil
}

func (m *MockProvider) SetBaseURL(baseURL string) {}

func (m *MockProvider) Name() string {
	return m.name()
}

func createChatErrorResponse(t *testing.T, message string) []byte {
	res := ChatErrorResponseModel{
		Error: message,
	}

	json, err := json.Marshal(res)
	assert.NoError(t, err)

	return append(json, []byte("\n")...)
}

func TestChatHandler(t *testing.T) {
	var (
		uuidStr = "f47ac10b-58cc-4372-a567-0e02b2c3d479"
		uuidFn  = func() uuid.UUID {
			return uuid.MustParse(uuidStr)
		}
		now   = time.Now()
		nowFn = func() time.Time {
			return now
		}
		fullChatRequest = ChatRequestModel{
			Model: "test-model",
			Messages: []ChatMessage{
				{
					Role:    "user",
					Content: "content",
				},
			},
			Temperature: new(1.7),
			MaxTokens:   new(1),
		}
		chatRequestWithoutTempMax = ChatRequestModel{
			Model: "test-model",
			Messages: []ChatMessage{
				{
					Role:    "user",
					Content: "content",
				},
			},
		}
		chatRequestFakeModelWithoutTempMax = ChatRequestModel{
			Model: "fake-model",
			Messages: []ChatMessage{
				{
					Role:    "user",
					Content: "content",
				},
			},
		}
		chatRequestWithoutModel = ChatRequestModel{
			Messages: []ChatMessage{
				{
					Role:    "user",
					Content: "content",
				},
			},
		}
		chatRequestWithoutMessages = ChatRequestModel{
			Model: "test-model",
		}
		chatResponse = ChatResponse{
			ID:      uuidStr,
			Object:  "chat.completion",
			Created: now.Unix(),
			Model:   "test-upstream-model",
			Choices: []ChatChoice{
				{
					Index: 0,
					Message: ChatMessage{
						Role:    "assistant",
						Content: "test content",
					},
					FinishReason: "done",
				},
			},
			Usage: Usage{
				PromptTokens:     15,
				CompletionTokens: 20,
				TotalTokens:      35,
			},
		}
		fakeProviderChatResponse = fake.CreateFakeResponse()
		fakeChatResponse         = ChatResponse{
			ID:      uuidStr,
			Object:  "chat.completion",
			Created: now.Unix(),
			Model:   fakeProviderChatResponse.Model,
			Choices: []ChatChoice{
				{
					Index: 0,
					Message: ChatMessage{
						Role:    "assistant",
						Content: fakeProviderChatResponse.Content,
					},
					FinishReason: fakeProviderChatResponse.FinishReason,
				},
			},
			Usage: Usage{
				PromptTokens:     fakeProviderChatResponse.Usage.PromptTokens,
				CompletionTokens: fakeProviderChatResponse.Usage.CompletionTokens,
				TotalTokens:      fakeProviderChatResponse.Usage.TotalTokens,
			},
		}
	)

	fullChatRequestJson, err := json.Marshal(fullChatRequest)
	assert.NoError(t, err)

	chatRequestWithoutTempMaxJson, err := json.Marshal(chatRequestWithoutTempMax)
	assert.NoError(t, err)

	chatRequestFakeModelWithoutTempMaxJson, err := json.Marshal(chatRequestFakeModelWithoutTempMax)
	assert.NoError(t, err)

	chatResponseJson, err := json.Marshal(chatResponse)
	chatResponseJson = append(chatResponseJson, []byte("\n")...)
	assert.NoError(t, err)

	chatRequestWithoutModelJson, err := json.Marshal(chatRequestWithoutModel)
	assert.NoError(t, err)

	chatRequestWithoutMessagesJson, err := json.Marshal(chatRequestWithoutMessages)
	assert.NoError(t, err)

	chatResponseFakeModel, err := json.Marshal(fakeChatResponse)
	chatResponseFakeModel = append(chatResponseFakeModel, []byte("\n")...)
	assert.NoError(t, err)

	tests := []struct {
		name                 string
		providerRouter       ProviderRouter
		method               string
		requestBody          []byte
		expectedResponseBody []byte
		expectedStatusCode   int
	}{
		{
			name: "successful call",
			providerRouter: &MockProviderRouter{
				resolve: func(model string) (provider.Provider, string, error) {
					return &MockProvider{
						chat: func(ctx context.Context, req provider.ChatRequest) (provider.ChatResponse, error) {
							return provider.ChatResponse{
								Model:        "test-model",
								Content:      "test content",
								FinishReason: "done",
								Usage: provider.Usage{
									PromptTokens:     15,
									CompletionTokens: 20,
									TotalTokens:      35,
								},
							}, nil
						},
						name: func() string {
							return "test-provider"
						},
					}, "test-upstream-model", nil
				},
			},
			method:               http.MethodPost,
			requestBody:          fullChatRequestJson,
			expectedResponseBody: chatResponseJson,
			expectedStatusCode:   http.StatusOK,
		},
		{
			name: "successful call - fake model, fake provider",
			providerRouter: &MockProviderRouter{
				resolve: func(model string) (provider.Provider, string, error) {
					return fake.NewFakeProvider(false), "fake-model", nil
				},
			},
			method:               http.MethodPost,
			requestBody:          chatRequestFakeModelWithoutTempMaxJson,
			expectedResponseBody: chatResponseFakeModel,
			expectedStatusCode:   http.StatusOK,
		},
		{
			name: "successful call - with temperature and max tokens",
			providerRouter: &MockProviderRouter{
				resolve: func(model string) (provider.Provider, string, error) {
					return &MockProvider{
						chat: func(ctx context.Context, req provider.ChatRequest) (provider.ChatResponse, error) {
							return provider.ChatResponse{
								Model:        "test-model",
								Content:      "test content",
								FinishReason: "done",
								Usage: provider.Usage{
									PromptTokens:     15,
									CompletionTokens: 20,
									TotalTokens:      35,
								},
							}, nil
						},
						name: func() string {
							return "test-provider"
						},
					}, "test-upstream-model", nil
				},
			},
			method:               http.MethodPost,
			requestBody:          chatRequestWithoutTempMaxJson,
			expectedResponseBody: chatResponseJson,
			expectedStatusCode:   http.StatusOK,
		},
		{
			name: "fails when fake provider fails",
			providerRouter: &MockProviderRouter{
				resolve: func(model string) (provider.Provider, string, error) {
					return fake.NewFakeProvider(true), "fake-model", nil
				},
			},
			method:               http.MethodPost,
			requestBody:          chatRequestFakeModelWithoutTempMaxJson,
			expectedResponseBody: createChatErrorResponse(t, "error from fake provider"),
			expectedStatusCode:   http.StatusBadRequest,
		},
		{
			name: "fails when request is not valid",
			providerRouter: &MockProviderRouter{
				resolve: func(model string) (provider.Provider, string, error) {
					return &MockProvider{
						chat: func(ctx context.Context, req provider.ChatRequest) (provider.ChatResponse, error) {
							return provider.ChatResponse{
								Model:        "test-model",
								Content:      "test content",
								FinishReason: "done",
								Usage: provider.Usage{
									PromptTokens:     15,
									CompletionTokens: 20,
									TotalTokens:      35,
								},
							}, nil
						},
						name: func() string {
							return "test-provider"
						},
					}, "test-upstream-model", nil
				},
			},
			method:               http.MethodPost,
			requestBody:          []byte("invalid"),
			expectedResponseBody: createChatErrorResponse(t, "invalid request"),
			expectedStatusCode:   http.StatusBadRequest,
		},
		{
			name: "fails when model is not set in request",
			providerRouter: &MockProviderRouter{
				resolve: func(model string) (provider.Provider, string, error) {
					return &MockProvider{
						chat: func(ctx context.Context, req provider.ChatRequest) (provider.ChatResponse, error) {
							return provider.ChatResponse{
								Model:        "test-model",
								Content:      "test content",
								FinishReason: "done",
								Usage: provider.Usage{
									PromptTokens:     15,
									CompletionTokens: 20,
									TotalTokens:      35,
								},
							}, nil
						},
						name: func() string {
							return "test-provider"
						},
					}, "test-upstream-model", nil
				},
			},
			method:               http.MethodPost,
			requestBody:          chatRequestWithoutModelJson,
			expectedResponseBody: createChatErrorResponse(t, fmt.Errorf("%w: model must be provided", ErrInvalidRequest).Error()),
			expectedStatusCode:   http.StatusBadRequest,
		},
		{
			name: "fails when messages are empty in request",
			providerRouter: &MockProviderRouter{
				resolve: func(model string) (provider.Provider, string, error) {
					return &MockProvider{
						chat: func(ctx context.Context, req provider.ChatRequest) (provider.ChatResponse, error) {
							return provider.ChatResponse{
								Model:        "test-model",
								Content:      "test content",
								FinishReason: "done",
								Usage: provider.Usage{
									PromptTokens:     15,
									CompletionTokens: 20,
									TotalTokens:      35,
								},
							}, nil
						},
						name: func() string {
							return "test-provider"
						},
					}, "test-upstream-model", nil
				},
			},
			method:               http.MethodPost,
			requestBody:          chatRequestWithoutMessagesJson,
			expectedResponseBody: createChatErrorResponse(t, fmt.Errorf("%w: at least 1 message must be provided", ErrInvalidRequest).Error()),
			expectedStatusCode:   http.StatusBadRequest,
		},
		{
			name: "fails when provider cannot be resolved",
			providerRouter: &MockProviderRouter{
				resolve: func(model string) (provider.Provider, string, error) {
					return nil, "", fmt.Errorf("provider not found")
				},
			},
			method:               http.MethodPost,
			requestBody:          fullChatRequestJson,
			expectedResponseBody: createChatErrorResponse(t, "unsupported model"),
			expectedStatusCode:   http.StatusNotFound,
		},
		{
			name: "fails when provider returns regular error",
			providerRouter: &MockProviderRouter{
				resolve: func(model string) (provider.Provider, string, error) {
					return &MockProvider{
						chat: func(ctx context.Context, req provider.ChatRequest) (provider.ChatResponse, error) {
							return provider.ChatResponse{}, fmt.Errorf("error from provider")
						},
						name: func() string {
							return "test-provider"
						},
					}, "test-upstream-model", nil
				},
			},
			method:               http.MethodPost,
			requestBody:          fullChatRequestJson,
			expectedResponseBody: createChatErrorResponse(t, "internal server error"),
			expectedStatusCode:   http.StatusInternalServerError,
		},
		{
			name: "fails when provider returns ProviderError with badrequest",
			providerRouter: &MockProviderRouter{
				resolve: func(model string) (provider.Provider, string, error) {
					return &MockProvider{
						chat: func(ctx context.Context, req provider.ChatRequest) (provider.ChatResponse, error) {
							return provider.ChatResponse{}, fmt.Errorf("error from provider: %w", &provider.ProviderError{
								Provider:   "test-provider",
								StatusCode: http.StatusBadRequest,
								Retryable:  false,
								Err:        fmt.Errorf("provider error"),
							})
						},
						name: func() string {
							return "test-provider"
						},
					}, "test-upstream-model", nil
				},
			},
			method:               http.MethodPost,
			requestBody:          fullChatRequestJson,
			expectedResponseBody: createChatErrorResponse(t, "provider error"),
			expectedStatusCode:   http.StatusBadRequest,
		},
		{
			name: "fails when provider returns ProviderError with 0 status code - fallbacks to internal server error",
			providerRouter: &MockProviderRouter{
				resolve: func(model string) (provider.Provider, string, error) {
					return &MockProvider{
						chat: func(ctx context.Context, req provider.ChatRequest) (provider.ChatResponse, error) {
							return provider.ChatResponse{}, fmt.Errorf("error from provider: %w", &provider.ProviderError{
								Provider:  "test-provider",
								Retryable: false,
								Err:       fmt.Errorf("provider error"),
							})
						},
						name: func() string {
							return "test-provider"
						},
					}, "test-upstream-model", nil
				},
			},
			method:               http.MethodPost,
			requestBody:          fullChatRequestJson,
			expectedResponseBody: createChatErrorResponse(t, "provider error"),
			expectedStatusCode:   http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			h := ChatHandler(tt.providerRouter, nowFn, uuidFn)
			request := httptest.NewRequest(tt.method, "/", bytes.NewReader(tt.requestBody))
			recorder := httptest.NewRecorder()
			h.ServeHTTP(recorder, request)

			// Act
			response := recorder.Result()
			defer func() {
				_ = response.Body.Close()
			}()

			responseBody, err := io.ReadAll(response.Body)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedResponseBody, responseBody)
			assert.Equal(t, tt.expectedStatusCode, response.StatusCode)
		})
	}
}

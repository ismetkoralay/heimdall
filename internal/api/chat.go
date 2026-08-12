package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ismetkoralay/heimdall/internal/provider"
)

var ErrInvalidRequest = errors.New("error invalid request")

func ChatHandler(
	providerRouter ProviderRouter,
	now func() time.Time,
	newUUID func() uuid.UUID,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			_ = r.Body.Close()
		}()

		w.Header().Set("Content-Type", "application/json")

		var chatRequest ChatRequestModel
		if err := json.NewDecoder(r.Body).Decode(&chatRequest); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			errResponse := ChatErrorResponseModel{
				Error: "invalid request",
			}
			_ = json.NewEncoder(w).Encode(errResponse)
			return
		}

		if err := validateChatRequest(chatRequest); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			errResponse := ChatErrorResponseModel{
				Error: err.Error(),
			}
			_ = json.NewEncoder(w).Encode(errResponse)
			return
		}

		p, upstreamModel, err := providerRouter.Resolve(chatRequest.Model)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			errResponse := ChatErrorResponseModel{
				Error: "unsupported model",
			}
			_ = json.NewEncoder(w).Encode(errResponse)
			return
		}

		messages := make([]provider.Message, len(chatRequest.Messages))
		for i, message := range chatRequest.Messages {
			messages[i] = provider.Message{
				Role:    mapProviderRole(message.Role),
				Content: message.Content,
			}
		}

		chatResponse, err := p.Chat(r.Context(), provider.ChatRequest{
			Model:       upstreamModel,
			Messages:    messages,
			Temperature: chatRequest.Temperature,
			MaxTokens:   chatRequest.MaxTokens,
		})
		if err != nil {
			var providerError *provider.ProviderError
			if errors.As(err, &providerError) {
				statusCode := providerError.StatusCode
				if statusCode == 0 {
					statusCode = http.StatusInternalServerError
				}
				w.WriteHeader(statusCode)
				_ = json.NewEncoder(w).Encode(ChatErrorResponseModel{
					Error: providerError.Error(),
				})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			errResponse := ChatErrorResponseModel{
				Error: "internal server error",
			}
			_ = json.NewEncoder(w).Encode(errResponse)
			return
		}

		response := ChatResponse{
			ID:      newUUID().String(),
			Object:  "chat.completion",
			Created: now().Unix(),
			Model:   upstreamModel,
			Choices: []ChatChoice{
				{
					Index: 0,
					Message: ChatMessage{
						Role:    "assistant",
						Content: chatResponse.Content,
					},
					FinishReason: chatResponse.FinishReason,
				},
			},
			Usage: Usage{
				PromptTokens:     chatResponse.Usage.PromptTokens,
				CompletionTokens: chatResponse.Usage.CompletionTokens,
				TotalTokens:      chatResponse.Usage.TotalTokens,
			},
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	})

	return mux
}

func validateChatRequest(req ChatRequestModel) error {
	if req.Model == "" {
		return fmt.Errorf("%w: model must be provided", ErrInvalidRequest)
	}

	if len(req.Messages) < 1 {
		return fmt.Errorf("%w: at least 1 message must be provided", ErrInvalidRequest)
	}

	return nil
}

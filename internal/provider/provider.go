// Package provider defines the interfaces and types for the provider layer of the Heimdall application.
package provider

import (
	"context"
	"iter"
)

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleDeveloper Role = "developer"
)

type Message struct {
	Role    Role
	Content string
}

type Usage struct {
	PromptTokens     int
	CompletionTokens int
	// TotalTokens is PromptTokens + CompletionTokens.
	TotalTokens int
}

type ChatRequest struct {
	Model string
	// Messages is the ordered conversation history, including any system prompt.
	Messages []Message
	// Temperature is nil when unset, so a caller's zero value isn't sent upstream as an explicit 0.
	Temperature *float64
	// MaxTokens is nil when unset, so a caller's zero value isn't sent upstream as an explicit 0.
	MaxTokens *int
}

type ChatResponse struct {
	Model   string
	Content string
	// FinishReason is provider-defined (e.g. "stop", "length").
	FinishReason string
	Usage        Usage
}

type Chunk struct {
	// Delta is the incremental content for this chunk, not the accumulated response so far.
	Delta string
	// FinishReason is nil except on the final chunk.
	FinishReason *string
	// Usage is nil except on the final chunk.
	Usage *Usage
}

type EmbedRequest struct {
	Model string
	// Input is the batch of texts to embed.
	Input []string
}

type EmbedResponse struct {
	Model string
	// Embeddings[i] is the embedding for Input[i].
	Embeddings [][]float64
	Usage      Usage
}

type Provider interface {
	// Chat processes a chat request and returns a chat response or an error.
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
	// ChatStream processes a chat request and returns a stream of chunks or an error. The stream is represented as an iterator that yields chunks until the chat response is complete.
	ChatStream(ctx context.Context, req ChatRequest) iter.Seq2[Chunk, error]
	// Embed processes an embedding request and returns an embedding response or an error.
	Embed(ctx context.Context, req EmbedRequest) (EmbedResponse, error)
	// Name returns the name of the provider.
	Name() string
}

// Package api provides chat/chat stream endpoints for the Heimdall project.
package api

import (
	"github.com/ismetkoralay/heimdall/internal/provider"
)

type ChatRequestRole string

const (
	ChatRequestRoleUnknown   ChatRequestRole = "unknown"
	ChatRequestRoleAssistant ChatRequestRole = "assistant"
	ChatRequestRoleUser      ChatRequestRole = "user"
	ChatRequestRoleDeveloper ChatRequestRole = "developer"
	ChatRequestRoleSystem    ChatRequestRole = "system"
)

type ProviderRouter interface {
	Resolve(model string) (provider.Provider, string, error)
}

type ChatRequestModel struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature *float64      `json:"temperature,omitempty"`
	MaxTokens   *int          `json:"max_tokens,omitempty"`
}

type ChatErrorResponseModel struct {
	Error string `json:"error"`
}

type ChatMessage struct {
	Role    ChatRequestRole `json:"role"`
	Content string          `json:"content"`
	Refusal *string         `json:"refusal"`
}

type ChatChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type ChatResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []ChatChoice `json:"choices"`
	Usage   Usage        `json:"usage"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	// TotalTokens is PromptTokens + CompletionTokens.
	TotalTokens int `json:"total_tokens"`
}

func mapProviderRole(role ChatRequestRole) provider.Role {
	switch role {
	case ChatRequestRoleAssistant:
		return provider.RoleAssistant
	case ChatRequestRoleUser:
		return provider.RoleUser
	case ChatRequestRoleDeveloper:
		return provider.RoleDeveloper
	case ChatRequestRoleSystem:
		return provider.RoleSystem
	}
	return provider.RoleUnkown
}

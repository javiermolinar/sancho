// Package llm provides interfaces and implementations for LLM-based task planning.
package llm

import (
	"context"
)

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Client defines the interface for LLM providers.
type Client interface {
	// Chat sends messages to the LLM and returns the response.
	Chat(ctx context.Context, messages []Message) (string, error)

	// ChatJSON sends messages and parses the response as JSON into the provided type.
	ChatJSON(ctx context.Context, messages []Message, result any) error
}

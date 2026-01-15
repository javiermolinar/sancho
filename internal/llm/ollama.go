package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

const defaultOllamaBaseURL = "http://localhost:11434"

// OllamaClient implements the Client interface using an Ollama backend.
type OllamaClient struct {
	client  *ollama.LLM
	model   string
	baseURL string
}

// NewOllamaClient creates a new Ollama client.
func NewOllamaClient(model, baseURL string) (*OllamaClient, error) {
	if model == "" {
		return nil, errors.New("ollama model is required")
	}
	if baseURL == "" {
		baseURL = defaultOllamaBaseURL
	}

	client, err := ollama.New(
		ollama.WithModel(model),
		ollama.WithServerURL(baseURL),
	)
	if err != nil {
		return nil, fmt.Errorf("creating ollama client: %w", err)
	}

	return &OllamaClient{
		client:  client,
		model:   model,
		baseURL: baseURL,
	}, nil
}

// Chat sends messages to the LLM and returns the response.
func (c *OllamaClient) Chat(ctx context.Context, messages []Message) (string, error) {
	resp, err := c.client.GenerateContent(ctx, toLangChainMessages(messages), llms.WithModel(c.model))
	if err != nil {
		return "", fmt.Errorf("ollama chat: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}
	return resp.Choices[0].Content, nil
}

// ChatJSON sends messages and parses the response as JSON into the provided type.
func (c *OllamaClient) ChatJSON(ctx context.Context, messages []Message, result any) error {
	resp, err := c.client.GenerateContent(
		ctx,
		toLangChainMessages(messages),
		llms.WithModel(c.model),
		llms.WithJSONMode(),
	)
	if err != nil {
		return fmt.Errorf("ollama chat json: %w", err)
	}
	if len(resp.Choices) == 0 {
		return fmt.Errorf("no response choices returned")
	}

	content := extractJSON(resp.Choices[0].Content)
	if err := json.Unmarshal([]byte(content), result); err != nil {
		return fmt.Errorf("parsing JSON response: %w (content: %s)", err, resp.Choices[0].Content)
	}
	return nil
}

func toLangChainMessages(messages []Message) []llms.MessageContent {
	result := make([]llms.MessageContent, 0, len(messages))
	for _, msg := range messages {
		role := llms.ChatMessageTypeHuman
		switch strings.ToLower(msg.Role) {
		case "system":
			role = llms.ChatMessageTypeSystem
		case "assistant":
			role = llms.ChatMessageTypeAI
		case "user":
			role = llms.ChatMessageTypeHuman
		}
		result = append(result, llms.TextParts(role, msg.Content))
	}
	return result
}

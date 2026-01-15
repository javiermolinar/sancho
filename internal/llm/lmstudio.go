package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const defaultLMStudioBaseURL = "http://localhost:1234/v1"

// LMStudioClient implements the Client interface using LM Studio's OpenAI-compatible API.
type LMStudioClient struct {
	client  openai.Client
	model   string
	baseURL string
}

// NewLMStudioClient creates a new LM Studio client.
func NewLMStudioClient(model, baseURL string) (*LMStudioClient, error) {
	if strings.TrimSpace(model) == "" {
		return nil, errors.New("lm studio model is required")
	}
	if baseURL == "" {
		baseURL = defaultLMStudioBaseURL
	}

	apiKey := os.Getenv("LMSTUDIO_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if apiKey == "" {
		apiKey = "lm-studio"
	}

	client := openai.NewClient(
		option.WithBaseURL(baseURL),
		option.WithAPIKey(apiKey),
	)

	return &LMStudioClient{
		client:  client,
		model:   model,
		baseURL: baseURL,
	}, nil
}

// Chat sends messages to the LLM and returns the response.
func (c *LMStudioClient) Chat(ctx context.Context, messages []Message) (string, error) {
	openaiMessages := make([]openai.ChatCompletionMessageParamUnion, len(messages))
	for i, msg := range messages {
		switch msg.Role {
		case "system":
			openaiMessages[i] = openai.SystemMessage(msg.Content)
		case "user":
			openaiMessages[i] = openai.UserMessage(msg.Content)
		case "assistant":
			openaiMessages[i] = openai.AssistantMessage(msg.Content)
		default:
			openaiMessages[i] = openai.UserMessage(msg.Content)
		}
	}

	resp, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    c.model,
		Messages: openaiMessages,
	})
	if err != nil {
		return "", fmt.Errorf("lm studio chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return resp.Choices[0].Message.Content, nil
}

// ChatJSON sends messages and parses the response as JSON into the provided type.
func (c *LMStudioClient) ChatJSON(ctx context.Context, messages []Message, result any) error {
	content, err := c.Chat(ctx, messages)
	if err != nil {
		return err
	}

	jsonContent := extractJSON(content)
	if err := json.Unmarshal([]byte(jsonContent), result); err != nil {
		return fmt.Errorf("parsing JSON response: %w (content: %s)", err, content)
	}

	return nil
}

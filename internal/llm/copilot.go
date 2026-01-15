package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const (
	copilotTokenURL = "https://api.github.com/copilot_internal/v2/token"
	copilotBaseURL  = "https://api.githubcopilot.com"

	// DefaultModel is the default model to use for planning.
	DefaultModel = "gpt-4o"
)

// CopilotClient implements the Client interface using GitHub Copilot's API.
type CopilotClient struct {
	client     openai.Client
	model      string
	httpClient *http.Client
}

// tokenResponse represents the response from GitHub's token exchange endpoint.
type tokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

// NewCopilotClient creates a new Copilot client.
// It loads the GitHub token and exchanges it for a Copilot bearer token.
func NewCopilotClient(model string) (*CopilotClient, error) {
	if model == "" {
		model = DefaultModel
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}

	// Load GitHub token
	githubToken, err := LoadGitHubToken()
	if err != nil {
		return nil, fmt.Errorf("loading GitHub token: %w", err)
	}

	// Exchange for Copilot bearer token
	bearerToken, err := exchangeToken(httpClient, githubToken)
	if err != nil {
		return nil, fmt.Errorf("exchanging token: %w", err)
	}

	// Create OpenAI client configured for Copilot
	client := openai.NewClient(
		option.WithBaseURL(copilotBaseURL),
		option.WithAPIKey(bearerToken),
		option.WithHeader("Editor-Version", "Sancho/1.0"),
		option.WithHeader("Editor-Plugin-Version", "Sancho/1.0"),
		option.WithHeader("Copilot-Integration-Id", "vscode-chat"),
	)

	return &CopilotClient{
		client:     client,
		model:      model,
		httpClient: httpClient,
	}, nil
}

// exchangeToken exchanges a GitHub OAuth token for a Copilot bearer token.
func exchangeToken(httpClient *http.Client, githubToken string) (string, error) {
	req, err := http.NewRequest("GET", copilotTokenURL, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+githubToken)
	req.Header.Set("User-Agent", "Sancho/1.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("making request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token exchange failed (status %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	return tokenResp.Token, nil
}

// Chat sends messages to the LLM and returns the response.
func (c *CopilotClient) Chat(ctx context.Context, messages []Message) (string, error) {
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
		return "", fmt.Errorf("chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return resp.Choices[0].Message.Content, nil
}

// ChatJSON sends messages and parses the response as JSON into the provided type.
func (c *CopilotClient) ChatJSON(ctx context.Context, messages []Message, result any) error {
	content, err := c.Chat(ctx, messages)
	if err != nil {
		return err
	}

	// Try to extract JSON from the response (may be wrapped in markdown code blocks)
	jsonContent := extractJSON(content)

	if err := json.Unmarshal([]byte(jsonContent), result); err != nil {
		return fmt.Errorf("parsing JSON response: %w (content: %s)", err, content)
	}

	return nil
}

// extractJSON attempts to extract JSON from a string that may contain markdown formatting.
func extractJSON(s string) string {
	// Try to find ```json ... ``` block
	jsonStart := "```json"
	if idx := indexOf(s, jsonStart); idx != -1 {
		start := idx + len(jsonStart)
		// Skip newline after ```json
		for start < len(s) && (s[start] == '\n' || s[start] == '\r') {
			start++
		}
		// Find closing ```
		if end := indexOf(s[start:], "```"); end != -1 {
			result := s[start : start+end]
			// Trim trailing newlines
			for len(result) > 0 && (result[len(result)-1] == '\n' || result[len(result)-1] == '\r') {
				result = result[:len(result)-1]
			}
			return result
		}
	}

	// Try to find ``` ... ``` block (plain code block)
	codeStart := "```"
	if idx := indexOf(s, codeStart); idx != -1 {
		start := idx + len(codeStart)
		// Skip newline
		for start < len(s) && (s[start] == '\n' || s[start] == '\r') {
			start++
		}
		// Find closing ```
		if end := indexOf(s[start:], "```"); end != -1 {
			result := s[start : start+end]
			// Trim trailing newlines
			for len(result) > 0 && (result[len(result)-1] == '\n' || result[len(result)-1] == '\r') {
				result = result[:len(result)-1]
			}
			return result
		}
	}

	// Try to find raw JSON (starts with { or [)
	for i := 0; i < len(s); i++ {
		if s[i] == '{' || s[i] == '[' {
			// Find matching closing bracket
			depth := 0
			for j := i; j < len(s); j++ {
				switch s[j] {
				case '{', '[':
					depth++
				case '}', ']':
					depth--
					if depth == 0 {
						return s[i : j+1]
					}
				}
			}
		}
	}

	return s
}

// indexOf returns the index of the first occurrence of substr in s, or -1.
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

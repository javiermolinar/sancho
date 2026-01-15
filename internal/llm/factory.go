package llm

import (
	"fmt"
	"strings"
)

const (
	ProviderCopilot  = "copilot"
	ProviderOllama   = "ollama"
	ProviderLMStudio = "lmstudio"
)

// NewClient creates an LLM client based on provider configuration.
func NewClient(provider, model, baseURL string) (Client, error) {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "", ProviderCopilot:
		return NewCopilotClient(model)
	case ProviderOllama:
		return NewOllamaClient(model, baseURL)
	case ProviderLMStudio, "lm-studio", "llmstudio":
		return NewLMStudioClient(model, baseURL)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", provider)
	}
}

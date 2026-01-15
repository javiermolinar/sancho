package llm

import "testing"

func TestNewClient_Ollama(t *testing.T) {
	client, err := NewClient("ollama", "llama3", "")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	ollamaClient, ok := client.(*OllamaClient)
	if !ok {
		t.Fatalf("expected OllamaClient, got %T", client)
	}
	if ollamaClient.baseURL != defaultOllamaBaseURL {
		t.Errorf("baseURL = %q, want %q", ollamaClient.baseURL, defaultOllamaBaseURL)
	}
}

func TestNewClient_LMStudio(t *testing.T) {
	client, err := NewClient("lmstudio", "llama3", "")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	lmStudioClient, ok := client.(*LMStudioClient)
	if !ok {
		t.Fatalf("expected LMStudioClient, got %T", client)
	}
	if lmStudioClient.baseURL != defaultLMStudioBaseURL {
		t.Errorf("baseURL = %q, want %q", lmStudioClient.baseURL, defaultLMStudioBaseURL)
	}
}

func TestNewClient_UnsupportedProvider(t *testing.T) {
	_, err := NewClient("unknown", "model", "")
	if err == nil {
		t.Fatal("expected error for unsupported provider")
	}
}

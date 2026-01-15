package llm

import "testing"

func TestNewOllamaClient_DefaultBaseURL(t *testing.T) {
	client, err := NewOllamaClient("llama3", "")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if client.baseURL != defaultOllamaBaseURL {
		t.Errorf("baseURL = %q, want %q", client.baseURL, defaultOllamaBaseURL)
	}
}

func TestNewOllamaClient_EmptyModel(t *testing.T) {
	_, err := NewOllamaClient("", "")
	if err == nil {
		t.Fatal("expected error for empty model")
	}
}

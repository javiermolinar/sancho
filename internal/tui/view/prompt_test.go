package view

import "testing"

func TestPromptLinesIncludesSuggestions(t *testing.T) {
	state := PromptState{Value: "/p", Cursor: "_", ModePrompt: true}
	commands := []PromptCommand{{Name: "/plan", Description: "Plan tasks"}}
	lines := PromptLines(state, 40, commands)

	found := false
	for _, line := range lines {
		if line == "  /plan Plan tasks" {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected suggestion line, got %v", lines)
	}
}

func TestClampPromptLinesAddsEllipsis(t *testing.T) {
	lines := []string{"one", "two", "three"}
	clamped := ClampPromptLines(lines, 2, 5)
	if len(clamped) != 2 {
		t.Fatalf("clamped length = %d, want 2", len(clamped))
	}
	if clamped[1] == "two" {
		t.Fatalf("expected ellipsis on last line, got %q", clamped[1])
	}
}

package input

import "testing"

func TestPromptMatchingCommands(t *testing.T) {
	commands := []PromptCommand{
		{Name: "/plan", Description: "Plan"},
		{Name: "/week", Description: "Week"},
	}

	tests := []struct {
		name  string
		input string
		want  int
	}{
		{name: "no_slash", input: "plan", want: 0},
		{name: "empty", input: "", want: 0},
		{name: "full", input: "/plan", want: 1},
		{name: "prefix", input: "/p", want: 1},
		{name: "with_space", input: "/plan x", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PromptMatchingCommands(tt.input, commands)
			if len(got) != tt.want {
				t.Fatalf("matches = %d, want %d", len(got), tt.want)
			}
		})
	}
}

func TestPromptAutocomplete(t *testing.T) {
	commands := []PromptCommand{
		{Name: "/plan", Description: "Plan"},
		{Name: "/week", Description: "Week"},
	}

	value, ok := PromptAutocomplete("/p", commands)
	if !ok {
		t.Fatal("expected autocomplete")
	}
	if value != "/plan " {
		t.Fatalf("value = %q, want %q", value, "/plan ")
	}
}

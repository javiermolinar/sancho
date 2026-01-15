package theme

import (
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name      string
		themeName string
		wantName  string
		wantErr   bool
	}{
		{
			name:      "load mocha theme",
			themeName: "mocha",
			wantName:  "mocha",
			wantErr:   false,
		},
		{
			name:      "load macchiato theme",
			themeName: "macchiato",
			wantName:  "macchiato",
			wantErr:   false,
		},
		{
			name:      "load frappe theme",
			themeName: "frappe",
			wantName:  "frappe",
			wantErr:   false,
		},
		{
			name:      "load latte theme",
			themeName: "latte",
			wantName:  "latte",
			wantErr:   false,
		},
		{
			name:      "load light theme",
			themeName: "light",
			wantName:  "light",
			wantErr:   false,
		},
		{
			name:      "empty name defaults to mocha",
			themeName: "",
			wantName:  "mocha",
			wantErr:   false,
		},
		{
			name:      "invalid theme falls back to mocha",
			themeName: "nonexistent",
			wantName:  "mocha",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme, err := Load(tt.themeName)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Load(%q) expected error, got nil", tt.themeName)
				}
				return
			}
			if err != nil {
				t.Fatalf("Load(%q) unexpected error: %v", tt.themeName, err)
			}
			if theme.Name != tt.wantName {
				t.Errorf("Load(%q).Name = %q, want %q", tt.themeName, theme.Name, tt.wantName)
			}
		})
	}
}

func TestLoad_ThemeColors(t *testing.T) {
	theme, err := Load("mocha")
	if err != nil {
		t.Fatalf("Load(mocha) unexpected error: %v", err)
	}

	// Verify all required colors are present and valid hex format
	colors := map[string]string{
		"Bg":          theme.Bg,
		"BgHighlight": theme.BgHighlight,
		"BgSelection": theme.BgSelection,
		"Fg":          theme.Fg,
		"FgMuted":     theme.FgMuted,
		"Accent":      theme.Accent,
		"Deep":        theme.Deep,
		"Shallow":     theme.Shallow,
		"Current":     theme.Current,
		"Warning":     theme.Warning,
		"BaseBg":      theme.BaseBg,
		"ModalBorder": theme.ModalBorder,
		"TextPrimary": theme.TextPrimary,
		"TextMuted":   theme.TextMuted,
		"Highlight":   theme.Highlight,
	}

	for name, hex := range colors {
		if len(hex) != 7 {
			t.Errorf("theme.%s = %q, want 7-char hex string", name, hex)
			continue
		}
		if hex[0] != '#' {
			t.Errorf("theme.%s = %q, want hex string starting with #", name, hex)
		}
	}
}

func TestAvailable(t *testing.T) {
	available := Available()

	expected := []string{"mocha", "macchiato", "frappe", "latte", "light"}
	if len(available) != len(expected) {
		t.Errorf("Available() returned %d themes, want %d", len(available), len(expected))
	}

	for i, want := range expected {
		if i >= len(available) {
			break
		}
		if available[i] != want {
			t.Errorf("Available()[%d] = %q, want %q", i, available[i], want)
		}
	}
}

func TestIsAvailable(t *testing.T) {
	tests := []struct {
		name     string
		theme    string
		expected bool
	}{
		{name: "exact match", theme: "mocha", expected: true},
		{name: "case insensitive", theme: "Mocha", expected: true},
		{name: "missing theme", theme: "unknown", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAvailable(tt.theme); got != tt.expected {
				t.Errorf("IsAvailable(%q) = %t, want %t", tt.theme, got, tt.expected)
			}
		})
	}
}

func TestColor(t *testing.T) {
	hex := "#89b4fa"
	c := Color(hex)
	if string(c) != hex {
		t.Errorf("Color(%q) = %q, want %q", hex, string(c), hex)
	}
}

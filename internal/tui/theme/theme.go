// Package theme provides color themes for the TUI.
package theme

import (
	"embed"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pelletier/go-toml/v2"
)

//go:embed embedded/*.toml
var embeddedThemes embed.FS

// Theme holds all colors for a TUI theme.
type Theme struct {
	Name        string `toml:"name"`
	Bg          string `toml:"bg"`           // Base background
	BgHighlight string `toml:"bg_highlight"` // Task blocks, subtle highlight
	BgSelection string `toml:"bg_selection"` // Cursor, selection
	Fg          string `toml:"fg"`           // Primary foreground
	FgMuted     string `toml:"fg_muted"`     // Past tasks, muted elements
	Accent      string `toml:"accent"`       // Title, primary accent, borders
	Deep        string `toml:"deep"`         // Deep work tasks
	Shallow     string `toml:"shallow"`      // Shallow work tasks
	Current     string `toml:"current"`      // Current task border (time-based)
	Warning     string `toml:"warning"`      // Warnings, move mode

	// Modal palette (can override base theme values)
	BaseBg      string `toml:"base_bg"`
	ModalBorder string `toml:"modal_border"`
	TextPrimary string `toml:"text_primary"`
	TextMuted   string `toml:"text_muted"`
	Highlight   string `toml:"highlight"`
}

// Color returns a lipgloss.Color for the given hex string.
func Color(hex string) lipgloss.Color {
	return lipgloss.Color(hex)
}

// Load loads a theme by name from embedded files.
// Falls back to mocha if the theme is not found.
func Load(name string) (*Theme, error) {
	if name == "" {
		name = "mocha"
	}
	name = strings.ToLower(name)

	path := "embedded/" + name + ".toml"
	data, err := embeddedThemes.ReadFile(path)
	if err != nil {
		// Fallback to mocha
		if name != "mocha" {
			return Load("mocha")
		}
		return nil, fmt.Errorf("loading theme %q: %w", name, err)
	}

	var t Theme
	if err := toml.Unmarshal(data, &t); err != nil {
		return nil, fmt.Errorf("parsing theme %q: %w", name, err)
	}
	t.applyDefaults()

	return &t, nil
}

// ModalPalette provides the modal-specific colors derived from the theme.
type ModalPalette struct {
	BaseBg      string
	ModalBorder string
	TextPrimary string
	TextMuted   string
	Highlight   string
}

// Modal returns the modal palette, falling back to base theme colors when needed.
func (t *Theme) Modal() ModalPalette {
	return ModalPalette{
		BaseBg:      coalesce(t.BaseBg, t.BgHighlight, t.Bg),
		ModalBorder: coalesce(t.ModalBorder, t.Accent),
		TextPrimary: coalesce(t.TextPrimary, t.Fg),
		TextMuted:   coalesce(t.TextMuted, t.FgMuted),
		Highlight:   coalesce(t.Highlight, t.BgSelection, t.Accent),
	}
}

func (t *Theme) applyDefaults() {
	if t.BaseBg == "" {
		t.BaseBg = coalesce(t.BgHighlight, t.Bg)
	}
	if t.ModalBorder == "" {
		t.ModalBorder = t.Accent
	}
	if t.TextPrimary == "" {
		t.TextPrimary = t.Fg
	}
	if t.TextMuted == "" {
		t.TextMuted = t.FgMuted
	}
	if t.Highlight == "" {
		t.Highlight = coalesce(t.BgSelection, t.Accent)
	}
}

func coalesce(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// Available returns a list of available theme names.
func Available() []string {
	return []string{"mocha", "macchiato", "frappe", "latte", "light"}
}

// IsAvailable reports whether a theme name is available.
func IsAvailable(name string) bool {
	name = strings.ToLower(name)
	for _, themeName := range Available() {
		if themeName == name {
			return true
		}
	}
	return false
}

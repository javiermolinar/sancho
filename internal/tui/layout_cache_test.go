// Package tui provides the terminal user interface for sancho.
package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"github.com/javiermolinar/sancho/internal/tui/theme"
)

func TestBuildLayoutCache_FooterAuxBackground(t *testing.T) {
	palette := &theme.Theme{
		Bg:          "#101010",
		BgHighlight: "#202020",
		BgSelection: "#303030",
		Fg:          "#ffffff",
		FgMuted:     "#aaaaaa",
		Accent:      "#ff0000",
		Deep:        "#00ff00",
		Shallow:     "#0000ff",
		Current:     "#ffff00",
		Warning:     "#ff00ff",
	}
	styles := NewStyles(palette)
	m := Model{styles: styles, prompt: textinput.New()}

	layout := m.buildLayoutCache(100, 40)
	appH, _ := styles.AppStyle.GetFrameSize()
	innerW := 100 - appH
	if innerW < 0 {
		innerW = 0
	}

	footerBg, ok := layout.FooterAuxStyle.GetBackground().(lipgloss.Color)
	if !ok {
		t.Fatalf("FooterAuxStyle background type = %T, want lipgloss.Color", layout.FooterAuxStyle.GetBackground())
	}
	if footerBg != lipgloss.Color(palette.Bg) {
		t.Fatalf("FooterAuxStyle background = %q, want %q", footerBg, palette.Bg)
	}
	if got := layout.FooterAuxStyle.GetWidth(); got != max(0, innerW) {
		t.Fatalf("FooterAuxStyle width = %d, want %d", got, max(0, innerW))
	}
	if got := layout.StatsBarStyle.GetWidth(); got != max(0, innerW) {
		t.Fatalf("StatsBarStyle width = %d, want %d", got, max(0, innerW))
	}
}

func TestBuildLayoutCache_PromptContentWidth(t *testing.T) {
	palette := &theme.Theme{
		Bg:          "#101010",
		BgHighlight: "#202020",
		BgSelection: "#303030",
		Fg:          "#ffffff",
		FgMuted:     "#aaaaaa",
		Accent:      "#ff0000",
		Deep:        "#00ff00",
		Shallow:     "#0000ff",
		Current:     "#ffff00",
		Warning:     "#ff00ff",
	}
	styles := NewStyles(palette)
	promptFrameW, _ := styles.PromptStyle.GetFrameSize()
	m := Model{styles: styles, prompt: textinput.New()}

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{name: "wide", width: 100, height: 40},
		{name: "narrow", width: 18, height: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layout := m.buildLayoutCache(tt.width, tt.height)
			appH, _ := styles.AppStyle.GetFrameSize()
			innerW := tt.width - appH
			if innerW < 0 {
				innerW = 0
			}

			expected := innerW - promptFrameW
			if expected < 0 {
				expected = 0
			}
			if expected < 20 && innerW >= promptFrameW+20 {
				expected = 20
			}

			if layout.PromptContentWidth != expected {
				t.Fatalf("PromptContentWidth = %d, want %d", layout.PromptContentWidth, expected)
			}
		})
	}
}

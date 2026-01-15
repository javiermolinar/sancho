package theme

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestNewPalette_TaskShades(t *testing.T) {
	base := &Theme{
		Bg:          "#101010",
		BgHighlight: "#202020",
		BgSelection: "#303030",
		Fg:          "#ffffff",
		FgMuted:     "#aaaaaa",
		Accent:      "#ff0000",
		Deep:        "#112233",
		Shallow:     "#445566",
		Current:     "#777777",
		Warning:     "#888888",
	}

	palette := NewPalette(base)

	if palette.DeepBg != lipgloss.Color(darkenColor(base.Deep)) {
		t.Fatalf("DeepBg = %q, want %q", palette.DeepBg, darkenColor(base.Deep))
	}
	if palette.ShallowBg != lipgloss.Color(darkenColor(base.Shallow)) {
		t.Fatalf("ShallowBg = %q, want %q", palette.ShallowBg, darkenColor(base.Shallow))
	}
	if palette.DeepPastBgAlt != lipgloss.Color(alternateShade(muteColor(base.Deep), false)) {
		t.Fatalf("DeepPastBgAlt = %q, want %q", palette.DeepPastBgAlt, alternateShade(muteColor(base.Deep), false))
	}
	if palette.ShallowPastBgAlt != lipgloss.Color(alternateShade(muteColor(base.Shallow), false)) {
		t.Fatalf("ShallowPastBgAlt = %q, want %q", palette.ShallowPastBgAlt, alternateShade(muteColor(base.Shallow), false))
	}
}

func TestNewPalette_ModalFallbacks(t *testing.T) {
	base := &Theme{
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

	palette := NewPalette(base)
	if palette.Modal.Bg != lipgloss.Color(base.BgHighlight) {
		t.Fatalf("Modal.Bg = %q, want %q", palette.Modal.Bg, base.BgHighlight)
	}
	if palette.Modal.Border.Dark != base.Accent {
		t.Fatalf("Modal.Border.Dark = %q, want %q", palette.Modal.Border.Dark, base.Accent)
	}
	if palette.Modal.Backdrop != lipgloss.Color(base.BgSelection) {
		t.Fatalf("Modal.Backdrop = %q, want %q", palette.Modal.Backdrop, base.BgSelection)
	}
}

func TestNewPalette_LightThemeInvertsShades(t *testing.T) {
	base := &Theme{
		Bg:          "#f5f5f5",
		BgHighlight: "#eeeeee",
		BgSelection: "#e0e0e0",
		Fg:          "#222222",
		FgMuted:     "#555555",
		Accent:      "#2f6feb",
		Deep:        "#1d8a8a",
		Shallow:     "#2f8f2f",
		Current:     "#c97b00",
		Warning:     "#c2410c",
	}

	palette := NewPalette(base)
	if relativeLuminance(string(palette.DeepBg)) <= relativeLuminance(base.Deep) {
		t.Fatalf("DeepBg luminance = %f, want greater than Deep", relativeLuminance(string(palette.DeepBg)))
	}
	if relativeLuminance(string(palette.ShallowBg)) <= relativeLuminance(base.Shallow) {
		t.Fatalf("ShallowBg luminance = %f, want greater than Shallow", relativeLuminance(string(palette.ShallowBg)))
	}
}

func TestChooseTextColorPrefersContrast(t *testing.T) {
	bg := "#f0f0f0"
	lightText := "#ffffff"
	darkText := "#111111"

	if got := chooseTextColor(bg, lightText, darkText); got != darkText {
		t.Fatalf("chooseTextColor(%q, %q, %q) = %q, want %q", bg, lightText, darkText, got, darkText)
	}
}

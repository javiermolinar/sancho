// Package tui provides the terminal user interface for sancho.
package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/javiermolinar/sancho/internal/tui/theme"
)

func TestStylesBackgroundCoverage(t *testing.T) {
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

	assertBg := func(t *testing.T, name string, style lipgloss.Style, want string) {
		t.Helper()
		bg, ok := style.GetBackground().(lipgloss.Color)
		if !ok {
			t.Fatalf("%s background type = %T, want lipgloss.Color", name, style.GetBackground())
		}
		if bg != lipgloss.Color(want) {
			t.Fatalf("%s background = %q, want %q", name, bg, want)
		}
	}

	assertBg(t, "EmptyCellStyle", styles.EmptyCellStyle, palette.Bg)
	assertBg(t, "SeparatorStyle", styles.SeparatorStyle, palette.Bg)
	assertBg(t, "TimeColumnStyle", styles.TimeColumnStyle, palette.Bg)
	assertBg(t, "TableStyle", styles.TableStyle, palette.Bg)
	assertBg(t, "ViewportStyle", styles.ViewportStyle, palette.Bg)
	assertBg(t, "StatsDeepStyle", styles.StatsDeepStyle, palette.Bg)
	assertBg(t, "StatsShallowStyle", styles.StatsShallowStyle, palette.Bg)
}

func TestTaskCurrentStyleWidthContrast(t *testing.T) {
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
	derived := theme.NewPalette(palette)

	style := styles.TaskCurrentStyleWidth(12, true)
	bg, ok := style.GetBackground().(lipgloss.Color)
	if !ok {
		t.Fatalf("TaskCurrentStyleWidth background type = %T, want lipgloss.Color", style.GetBackground())
	}
	if bg != lipgloss.Color(palette.Current) {
		t.Fatalf("TaskCurrentStyleWidth background = %q, want %q", bg, palette.Current)
	}

	fg, ok := style.GetForeground().(lipgloss.Color)
	if !ok {
		t.Fatalf("TaskCurrentStyleWidth foreground type = %T, want lipgloss.Color", style.GetForeground())
	}
	if fg != derived.TextOnCurrent {
		t.Fatalf("TaskCurrentStyleWidth foreground = %q, want %q", fg, derived.TextOnCurrent)
	}
}

func TestCursorStyleContrast(t *testing.T) {
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

	bg, ok := styles.CursorStyle.GetBackground().(lipgloss.Color)
	if !ok {
		t.Fatalf("CursorStyle background type = %T, want lipgloss.Color", styles.CursorStyle.GetBackground())
	}
	if bg != lipgloss.Color(palette.BgSelection) {
		t.Fatalf("CursorStyle background = %q, want %q", bg, palette.BgSelection)
	}

	fg, ok := styles.CursorStyle.GetForeground().(lipgloss.Color)
	if !ok {
		t.Fatalf("CursorStyle foreground type = %T, want lipgloss.Color", styles.CursorStyle.GetForeground())
	}
	if fg != lipgloss.Color(palette.Accent) {
		t.Fatalf("CursorStyle foreground = %q, want %q", fg, palette.Accent)
	}
}

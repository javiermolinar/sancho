package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func TestOverlayToggle(t *testing.T) {
	overlay := NewOverlayModel()
	if overlay.Active() {
		t.Fatalf("expected overlay to start inactive")
	}

	overlay.Toggle()
	if !overlay.Active() {
		t.Fatalf("expected overlay to be active after toggle")
	}

	overlay.Toggle()
	if overlay.Active() {
		t.Fatalf("expected overlay to be inactive after second toggle")
	}
}

func TestOverlayRenderInactiveReturnsBase(t *testing.T) {
	overlay := NewOverlayModel()
	base := "alpha\nbeta"
	got := overlay.Render(base, 10, 2, "content")
	if got != base {
		t.Fatalf("expected base content unchanged when inactive")
	}
}

func TestOverlayRenderAddsOverlay(t *testing.T) {
	overlay := NewOverlayModel()
	overlay.SetBackground(lipgloss.Color("#0c0c0c"))
	overlay.Toggle()

	width := 30
	height := 12
	row := strings.Repeat(".", width)
	base := strings.Repeat(row+"\n", height-1) + row
	content := "TASK FORM"
	got := overlay.Render(base, width, height, content)

	lines := strings.Split(got, "\n")
	if len(lines) != height {
		t.Fatalf("expected %d lines, got %d", height, len(lines))
	}

	boxW, boxH := overlay.boxSize(width, height)
	if boxW <= 0 || boxH <= 0 {
		t.Fatalf("expected non-zero box size")
	}
	top := (height - boxH) / 2
	bgSeq := ansi.Style{}.BackgroundColor(ansi.HexColor(string(overlay.bgColor))).String()
	stripped := ansi.Strip(got)
	if !strings.Contains(stripped, content) {
		t.Fatalf("expected rendered content to include task form text")
	}

	for i, line := range lines {
		lineWidth := lipgloss.Width(line)
		if lineWidth != width {
			t.Fatalf("expected line width %d, got %d", width, lineWidth)
		}

		hasBg := strings.Contains(line, bgSeq)
		if i >= top && i < top+boxH {
			if !hasBg {
				t.Fatalf("expected overlay background on line %d", i)
			}
		} else if hasBg {
			t.Fatalf("expected no overlay background on line %d", i)
		}
	}
}

func TestOverlayRenderUsesBackgroundColor(t *testing.T) {
	overlay := NewOverlayModel()
	overlay.SetBackground(lipgloss.Color("#123456"))
	overlay.Toggle()

	width := 20
	height := 6
	row := strings.Repeat(".", width)
	base := strings.Repeat(row+"\n", height-1) + row
	got := overlay.Render(base, width, height, "x")

	bgSeq := ansi.Style{}.BackgroundColor(ansi.HexColor(string(overlay.bgColor))).String()
	if !strings.Contains(got, bgSeq) {
		t.Fatalf("expected overlay background sequence in output")
	}
}

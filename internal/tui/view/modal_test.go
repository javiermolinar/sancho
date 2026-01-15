// Package view provides rendering helpers for the TUI.
package view

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderModalButtons_UsesModalBodySeparator(t *testing.T) {
	styles := ModalStyles{
		ModalBodyStyle:         lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		ModalButtonStyle:       lipgloss.NewStyle(),
		ModalButtonActiveStyle: lipgloss.NewStyle(),
	}

	view := RenderModalButtons(styles, "[Enter] Save", "[Esc] Cancel")
	sep := styles.ModalBodyStyle.Render(" ")
	if !strings.Contains(view, sep) {
		t.Fatalf("expected modal button separator to use modal body style")
	}
}

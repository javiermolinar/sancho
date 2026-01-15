// Package view provides rendering helpers for the TUI.
package view

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ModalStyles groups the styles needed to render modal frames and buttons.
type ModalStyles struct {
	ModalHeaderStyle       lipgloss.Style
	ModalTitleStyle        lipgloss.Style
	ModalFooterStyle       lipgloss.Style
	ModalStyle             lipgloss.Style
	ModalButtonStyle       lipgloss.Style
	ModalButtonActiveStyle lipgloss.Style
	ModalBodyStyle         lipgloss.Style
}

// RenderModalFrame renders a modal with the provided title, body, and footer.
func RenderModalFrame(title, body, footer string, styles ModalStyles) string {
	var b strings.Builder

	header := styles.ModalHeaderStyle.Render(styles.ModalTitleStyle.Render(title))
	b.WriteString(header)
	if body != "" {
		b.WriteString("\n\n")
		b.WriteString(body)
	}
	if footer != "" {
		b.WriteString("\n\n")
		b.WriteString(styles.ModalFooterStyle.Render(footer))
	}

	return styles.ModalStyle.Render(b.String())
}

// RenderModalButtons renders a row of modal buttons with the first one active.
func RenderModalButtons(styles ModalStyles, labels ...string) string {
	parts := make([]string, 0, len(labels))
	for i, label := range labels {
		style := styles.ModalButtonStyle
		if i == 0 {
			style = styles.ModalButtonActiveStyle
		}
		parts = append(parts, style.Render(label))
	}
	sep := styles.ModalBodyStyle.Render(" ")
	return strings.Join(parts, sep)
}

// RenderModalButtonsCompact renders a compact row of modal buttons.
func RenderModalButtonsCompact(styles ModalStyles, labels ...string) string {
	parts := make([]string, 0, len(labels))
	buttonStyle := styles.ModalButtonStyle.Padding(0, 1)
	activeStyle := styles.ModalButtonActiveStyle.Padding(0, 1)
	for i, label := range labels {
		style := buttonStyle
		if i == 0 {
			style = activeStyle
		}
		parts = append(parts, style.Render(label))
	}
	sep := styles.ModalBodyStyle.Render(" ")
	return strings.Join(parts, sep)
}

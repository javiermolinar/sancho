// Package view provides rendering helpers for the TUI.
package view

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// WeekSummaryLineStyle indicates how a week summary line should be styled.
type WeekSummaryLineStyle int

const (
	WeekSummaryLineBody WeekSummaryLineStyle = iota
	WeekSummaryLineMeta
	WeekSummaryLineSection
)

// WeekSummaryLine is a display-ready line for the week summary modal.
type WeekSummaryLine struct {
	Text  string
	Style WeekSummaryLineStyle
}

// WeekSummaryStyles groups styles for week summary rendering.
type WeekSummaryStyles struct {
	BodyStyle         stringRenderer
	MetaStyle         stringRenderer
	SectionTitleStyle stringRenderer
}

// RenderWeekSummaryBody renders week summary lines into a wrapped modal body.
func RenderWeekSummaryBody(lines []WeekSummaryLine, styles WeekSummaryStyles, contentWidth int) string {
	if len(lines) == 0 {
		return ""
	}

	rendered := make([]string, 0, len(lines))
	for _, line := range lines {
		rendered = append(rendered, wrapWeekSummaryLine(line, styles, contentWidth)...)
	}
	return strings.Join(rendered, "\n")
}

// ModalContentWidth returns the content width for a modal body.
func ModalContentWidth(style lipgloss.Style, fallback int) int {
	width := style.GetWidth()
	if width <= 0 {
		return fallback
	}
	contentWidth := width - 4
	if contentWidth < 10 {
		return 10
	}
	return contentWidth
}

func wrapWeekSummaryLine(line WeekSummaryLine, styles WeekSummaryStyles, width int) []string {
	switch line.Style {
	case WeekSummaryLineSection:
		return wrapModalText(styles.SectionTitleStyle, line.Text, width)
	case WeekSummaryLineMeta:
		return wrapModalText(styles.MetaStyle, line.Text, width)
	default:
		return wrapModalText(styles.BodyStyle, line.Text, width)
	}
}

func wrapModalText(style stringRenderer, text string, width int) []string {
	if width <= 0 {
		return []string{style.Render("")}
	}
	lines := WrapTextToWidths(text, width, width)
	if len(lines) == 0 {
		return []string{style.Render("")}
	}

	wrapped := make([]string, 0, len(lines))
	for _, line := range lines {
		wrapped = append(wrapped, style.Render(line))
	}
	return wrapped
}

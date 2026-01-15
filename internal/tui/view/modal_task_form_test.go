// Package view provides rendering helpers for the TUI.
package view

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderTaskFormBody_IncludesSections(t *testing.T) {
	styles := TaskFormStyles{
		TagStyle:          lipgloss.NewStyle(),
		BodyStyle:         lipgloss.NewStyle(),
		SectionTitleStyle: lipgloss.NewStyle().Bold(true),
		DurationActive:    lipgloss.NewStyle(),
		DurationInactive:  lipgloss.NewStyle(),
		HintStyle:         lipgloss.NewStyle(),
	}
	model := TaskFormModel{
		MetaDate:        "Mon Jan 2",
		TimeRange:       "09:00-10:00",
		DurationLabel:   "1h",
		NameValue:       "Test",
		DescStyle:       lipgloss.NewStyle(),
		DurationOptions: []string{"30m", "1h"},
		ActiveDuration:  1,
	}

	body := RenderTaskFormBody(model, styles)
	if !strings.Contains(body, styles.SectionTitleStyle.Render("TASK NAME")) {
		t.Fatalf("expected task name section title")
	}
	if !strings.Contains(body, styles.SectionTitleStyle.Render("DURATION")) {
		t.Fatalf("expected duration section title")
	}
}

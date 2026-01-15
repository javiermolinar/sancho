// Package view provides rendering helpers for the TUI.
package view

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderTaskDetailBody_UsesBodyStyleForDescription(t *testing.T) {
	styles := TaskDetailStyles{
		BodyStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("6")),
		LabelStyle: lipgloss.NewStyle(),
	}
	model := TaskDetailModel{
		Description:   "Write report",
		CategoryIcon:  "D",
		CategoryLabel: "Deep work",
		TimeRange:     "10:00 - 10:30 (30m)",
		DateLabel:     "Monday, Jan 2, 2006",
		OutcomeLabel:  "On time",
	}

	body := RenderTaskDetailBody(model, styles)
	expected := styles.BodyStyle.Render(model.Description)
	if !strings.Contains(body, expected) {
		t.Fatalf("expected description to use body style")
	}
}

func TestRenderConfirmDeleteBody_UsesBodyStyleForMessage(t *testing.T) {
	styles := ConfirmDeleteStyles{
		BodyStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
	}
	model := ConfirmDeleteModel{HasTask: false}

	body := RenderConfirmDeleteBody(model, styles)
	expected := styles.BodyStyle.Render("This will mark the task as cancelled.\nAre you sure?")
	if !strings.Contains(body, expected) {
		t.Fatalf("expected confirm delete message to use body style")
	}
}

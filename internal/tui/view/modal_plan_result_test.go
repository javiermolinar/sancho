// Package view provides rendering helpers for the TUI.
package view

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderPlanResultBody_UsesIssuesSection(t *testing.T) {
	styles := PlanResultStyles{
		MetaStyle:         lipgloss.NewStyle(),
		SectionTitleStyle: lipgloss.NewStyle().Bold(true),
		BodyStyle:         lipgloss.NewStyle(),
	}
	model := PlanResultModel{
		IntroMessage:   "Intro",
		Issues:         []string{"Missing time"},
		NoTasks:        true,
		NoTasksMessage: "No tasks proposed.",
		Summary:        "Total: 0 tasks",
		AmendHint:      "Press m to amend.",
	}

	body := RenderPlanResultBody(model, styles)
	if !strings.Contains(body, styles.SectionTitleStyle.Render("ISSUES")) {
		t.Fatalf("expected issues section title")
	}
}

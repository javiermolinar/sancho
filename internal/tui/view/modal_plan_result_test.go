// Package view provides rendering helpers for the TUI.
package view

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/javiermolinar/sancho/internal/dwplanner"
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

func TestRenderPlanResultBody_UsesWarningsSection(t *testing.T) {
	styles := PlanResultStyles{
		MetaStyle:         lipgloss.NewStyle(),
		SectionTitleStyle: lipgloss.NewStyle().Bold(true),
		BodyStyle:         lipgloss.NewStyle(),
	}
	model := PlanResultModel{
		IntroMessage:   "Intro",
		Warnings:       []string{"Low capacity"},
		NoTasks:        true,
		NoTasksMessage: "No tasks proposed.",
		Summary:        "Total: 0 tasks",
		AmendHint:      "Press m to amend.",
	}

	body := RenderPlanResultBody(model, styles)
	if !strings.Contains(body, styles.SectionTitleStyle.Render("WARNINGS")) {
		t.Fatalf("expected warnings section title")
	}
	if !strings.Contains(body, "- Low capacity") {
		t.Fatalf("expected warning content")
	}
}

func TestNewPlanResultModelBuildsIssuesAndWarnings(t *testing.T) {
	result := &dwplanner.PlanResult{
		ValidationErrors: []dwplanner.ValidationError{
			{Message: "Invalid time"},
		},
		Warnings: []string{"Short day"},
		TasksByDate: map[string][]dwplanner.PlannedTask{
			"2026-01-12": {
				{
					Description:    "Focus block",
					Category:       "deep",
					ScheduledStart: "09:00",
					ScheduledEnd:   "10:00",
				},
			},
		},
		SortedDates: []string{"2026-01-12"},
	}

	model := NewPlanResultModel(result)
	if len(model.Issues) != 1 || model.Issues[0] != "Invalid time" {
		t.Fatalf("expected validation issues to be mapped")
	}
	if len(model.Warnings) != 1 || model.Warnings[0] != "Short day" {
		t.Fatalf("expected warnings to be mapped")
	}
	if len(model.Days) != 1 || model.Days[0].DateLabel == "" {
		t.Fatalf("expected days to be mapped with date labels")
	}
	if len(model.Days[0].Lines) != 1 || !strings.Contains(model.Days[0].Lines[0], "Focus block") {
		t.Fatalf("expected planned task lines to be formatted")
	}
}

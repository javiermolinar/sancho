// Package view provides rendering helpers for the TUI.
package view

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TaskFormModel contains the fields needed to render the task form body.
type TaskFormModel struct {
	MetaDate         string
	TimeRange        string
	DurationLabel    string
	NameValue        string
	NameLocked       bool
	DescStyle        lipgloss.Style
	DurationOptions  []string
	ActiveDuration   int
	ShowDurationHint bool
}

// TaskFormStyles groups styles for the task form body.
type TaskFormStyles struct {
	TagStyle          lipgloss.Style
	BodyStyle         lipgloss.Style
	SectionTitleStyle lipgloss.Style
	DurationActive    lipgloss.Style
	DurationInactive  lipgloss.Style
	HintStyle         lipgloss.Style
}

// RenderTaskFormBody renders the modal body for the task form.
func RenderTaskFormBody(model TaskFormModel, styles TaskFormStyles) string {
	var body strings.Builder
	sep := styles.BodyStyle.Render(" ")

	meta := styles.TagStyle.Render(model.MetaDate)
	timeTag := styles.TagStyle.Render(model.TimeRange)
	durationTag := styles.TagStyle.Render(model.DurationLabel)
	body.WriteString(meta + sep + timeTag + sep + durationTag + "\n\n")

	body.WriteString(styles.SectionTitleStyle.Render("TASK NAME") + "\n")
	body.WriteString(model.DescStyle.Render(model.NameValue) + "\n")
	if model.NameLocked {
		body.WriteString(styles.HintStyle.Render("Name locked for past tasks.") + "\n")
	}
	body.WriteString("\n")

	body.WriteString(styles.SectionTitleStyle.Render("DURATION") + "\n")
	parts := make([]string, 0, len(model.DurationOptions))
	for i, label := range model.DurationOptions {
		if i == model.ActiveDuration {
			parts = append(parts, styles.DurationActive.Render(label))
		} else {
			parts = append(parts, styles.DurationInactive.Render(label))
		}
	}
	body.WriteString(strings.Join(parts, sep))
	if model.ShowDurationHint {
		body.WriteString(sep + styles.HintStyle.Render("Use left/right"))
	}
	body.WriteString("\n")

	return body.String()
}

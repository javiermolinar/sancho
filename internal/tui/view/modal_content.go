// Package view provides rendering helpers for the TUI.
package view

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TaskDetailModel contains the fields needed to render the task detail body.
type TaskDetailModel struct {
	Description   string
	CategoryIcon  string
	CategoryLabel string
	TimeRange     string
	DateLabel     string
	OutcomeLabel  string
}

// TaskDetailStyles groups styles for the task detail body.
type TaskDetailStyles struct {
	BodyStyle  lipgloss.Style
	LabelStyle lipgloss.Style
}

// RenderTaskDetailBody renders the modal body for task details.
func RenderTaskDetailBody(model TaskDetailModel, styles TaskDetailStyles) string {
	var body strings.Builder

	body.WriteString(" " + styles.BodyStyle.Render(model.Description) + "\n\n")
	body.WriteString(styles.BodyStyle.Render(fmt.Sprintf(" [%s] %s", model.CategoryIcon, model.CategoryLabel)) + "\n")
	body.WriteString(styles.BodyStyle.Render(" "+model.TimeRange) + "\n")
	body.WriteString(styles.BodyStyle.Render(" "+model.DateLabel) + "\n\n")
	body.WriteString(styles.LabelStyle.Render(" Outcome:") + styles.BodyStyle.Render(model.OutcomeLabel))

	return body.String()
}

// ConfirmDeleteModel contains the fields needed to render the confirm delete body.
type ConfirmDeleteModel struct {
	Description string
	TimeRange   string
	DateLabel   string
	HasTask     bool
}

// ConfirmDeleteStyles groups styles for the confirm delete body.
type ConfirmDeleteStyles struct {
	BodyStyle lipgloss.Style
}

// RenderConfirmDeleteBody renders the modal body for the delete confirmation.
func RenderConfirmDeleteBody(model ConfirmDeleteModel, styles ConfirmDeleteStyles) string {
	var body strings.Builder

	if model.HasTask {
		body.WriteString(styles.BodyStyle.Render(fmt.Sprintf("\"%s\"", model.Description)) + "\n")
		body.WriteString(styles.BodyStyle.Render(model.TimeRange) + "\n")
		body.WriteString(styles.BodyStyle.Render(model.DateLabel) + "\n\n")
	}
	body.WriteString(styles.BodyStyle.Render("This will mark the task as cancelled.\nAre you sure?"))

	return body.String()
}

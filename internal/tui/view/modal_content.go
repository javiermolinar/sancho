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

// InitModalModel contains fields for the startup initialization modal.
type InitModalModel struct {
	ConfigPath    string
	DBPath        string
	ConfigMissing bool
	DBMissing     bool
	ErrorMessage  string
}

// InitModalStyles groups styles for the initialization modal body.
type InitModalStyles struct {
	BodyStyle  lipgloss.Style
	LabelStyle lipgloss.Style
	HintStyle  lipgloss.Style
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

// RenderInitBody renders the modal body for startup initialization.
func RenderInitBody(model InitModalModel, styles InitModalStyles) string {
	var body strings.Builder

	body.WriteString(styles.BodyStyle.Render("Sancho needs permission to create its configuration and database files.") + "\n\n")

	if model.ConfigMissing {
		body.WriteString(styles.LabelStyle.Render("Config:") + "\n")
		body.WriteString(styles.BodyStyle.Render(" "+model.ConfigPath) + "\n")
	}
	if model.DBMissing {
		body.WriteString(styles.LabelStyle.Render("Database:") + "\n")
		body.WriteString(styles.BodyStyle.Render(" "+model.DBPath) + "\n")
	}

	body.WriteString("\n")
	body.WriteString(styles.HintStyle.Render("These files are required to continue."))

	if model.ErrorMessage != "" {
		body.WriteString("\n\n")
		body.WriteString(styles.LabelStyle.Render("Error:") + " " + styles.BodyStyle.Render(model.ErrorMessage))
	}

	return body.String()
}

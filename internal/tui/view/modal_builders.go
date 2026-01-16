// Package view provides rendering helpers for the TUI.
package view

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/javiermolinar/sancho/internal/dwplanner"
	"github.com/javiermolinar/sancho/internal/task"
)

// TaskFormInput contains the data needed to build a task form model.
type TaskFormInput struct {
	Date             time.Time
	StartTime        string
	EndTime          string
	DurationMinutes  int
	NameValue        string
	NameLocked       bool
	DescStyle        lipgloss.Style
	DurationOptions  []int
	ActiveDuration   int
	ShowDurationHint bool
}

// NewTaskFormModel builds a task form model from input data.
func NewTaskFormModel(input TaskFormInput) TaskFormModel {
	durationLabels := make([]string, 0, len(input.DurationOptions))
	for _, d := range input.DurationOptions {
		durationLabels = append(durationLabels, FormatDuration(d))
	}

	return TaskFormModel{
		MetaDate:         input.Date.Format("Mon Jan 2"),
		TimeRange:        fmt.Sprintf("%s-%s", input.StartTime, input.EndTime),
		DurationLabel:    FormatDuration(input.DurationMinutes),
		NameValue:        input.NameValue,
		NameLocked:       input.NameLocked,
		DescStyle:        input.DescStyle,
		DurationOptions:  durationLabels,
		ActiveDuration:   input.ActiveDuration,
		ShowDurationHint: input.ShowDurationHint,
	}
}

// NewTaskDetailModel builds a task detail model from a task.
func NewTaskDetailModel(t *task.Task) TaskDetailModel {
	categoryIcon := "D"
	categoryLabel := "Deep work"
	if t.IsShallow() {
		categoryIcon = "S"
		categoryLabel = "Shallow work"
	}
	outcomeStr := "Not set"
	if t.Outcome != nil {
		switch *t.Outcome {
		case task.OutcomeOnTime:
			outcomeStr = "On time"
		case task.OutcomeOver:
			outcomeStr = "Over time"
		case task.OutcomeUnder:
			outcomeStr = "Under time"
		}
	}

	return TaskDetailModel{
		Description:   t.Description,
		CategoryIcon:  categoryIcon,
		CategoryLabel: categoryLabel,
		TimeRange:     fmt.Sprintf("%s - %s (%s)", t.ScheduledStart, t.ScheduledEnd, FormatDuration(t.Duration())),
		DateLabel:     t.ScheduledDate.Format("Monday, Jan 2, 2006"),
		OutcomeLabel:  outcomeStr,
	}
}

// NewConfirmDeleteModel builds a delete confirmation model from a task.
func NewConfirmDeleteModel(t *task.Task) ConfirmDeleteModel {
	if t == nil {
		return ConfirmDeleteModel{HasTask: false}
	}
	return ConfirmDeleteModel{
		Description: t.Description,
		TimeRange:   fmt.Sprintf("%s - %s", t.ScheduledStart, t.ScheduledEnd),
		DateLabel:   t.ScheduledDate.Format("Mon Jan 2"),
		HasTask:     true,
	}
}

// NewPlanResultModel builds a plan result model from a plan result.
func NewPlanResultModel(result *dwplanner.PlanResult) PlanResultModel {
	if result == nil {
		return PlanResultModel{}
	}

	issues := make([]string, 0, len(result.ValidationErrors))
	for _, ve := range result.ValidationErrors {
		issues = append(issues, ve.Message)
	}
	days := make([]PlanResultDay, 0, len(result.SortedDates))
	for _, dateStr := range result.SortedDates {
		tasks := result.TasksByDate[dateStr]
		if len(tasks) == 0 {
			continue
		}
		label := dateStr + ":"
		if date, err := time.Parse("2006-01-02", dateStr); err == nil {
			label = date.Format("Mon Jan 2") + ":"
		}
		lines := make([]string, 0, len(tasks))
		for _, t := range tasks {
			icon := "D"
			if t.Category == "shallow" {
				icon = "S"
			}
			lines = append(lines, fmt.Sprintf("  [%s] %s-%s %s", icon, t.ScheduledStart, t.ScheduledEnd, t.Description))
		}
		days = append(days, PlanResultDay{
			DateLabel: label,
			Lines:     lines,
		})
	}

	summary := fmt.Sprintf("Total: %d tasks", result.TotalTasks())
	if len(result.SortedDates) > 1 {
		summary += fmt.Sprintf(" across %d days", len(result.SortedDates))
	}

	return PlanResultModel{
		IntroMessage:   "Review the draft and amend it before applying.",
		Issues:         issues,
		Warnings:       result.Warnings,
		Days:           days,
		NoTasks:        result.TotalTasks() == 0,
		NoTasksMessage: "No tasks proposed.",
		Summary:        summary,
		AmendHint:      "Press m to amend the plan text before applying.",
	}
}

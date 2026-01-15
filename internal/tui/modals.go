// Package tui provides the terminal user interface for sancho.
package tui

import (
	"fmt"
	"time"

	"github.com/javiermolinar/sancho/internal/task"
	"github.com/javiermolinar/sancho/internal/tui/view"
)

// renderModal renders the current modal.
func (m Model) renderModal() string {
	switch m.modalType {
	case ModalTaskForm:
		return m.renderTaskFormModal()
	case ModalTaskDetail:
		return m.renderTaskDetailModal()
	case ModalConfirmDelete:
		return m.renderConfirmDeleteModal()
	case ModalPlanResult:
		return m.renderPlanResultModal()
	case ModalWeekSummary:
		return m.renderWeekSummaryModal()
	case ModalInit:
		return m.renderInitModal()
	default:
		return ""
	}
}

func (m Model) modalStyles() view.ModalStyles {
	return view.ModalStyles{
		ModalHeaderStyle:       m.styles.ModalHeaderStyle,
		ModalTitleStyle:        m.styles.ModalTitleStyle,
		ModalFooterStyle:       m.styles.ModalFooterStyle,
		ModalStyle:             m.styles.ModalStyle,
		ModalButtonStyle:       m.styles.ModalButtonStyle,
		ModalButtonActiveStyle: m.styles.ModalButtonActiveStyle,
		ModalBodyStyle:         m.styles.ModalBodyStyle,
	}
}

// renderTaskFormModal renders the task creation form.
func (m Model) renderTaskFormModal() string {
	taskDate := m.weekStart.AddDate(0, 0, m.cursor.Day)
	startTime := m.slotToTime(m.cursor.Slot)
	duration := durationOptions[m.formDuration]
	endTime := addMinutesToTime(startTime, duration)
	title := "New Task"
	nameLocked := false
	nameValue := m.formDesc.View()
	if m.modalTask != nil {
		title = "Edit Task"
		taskDate = m.modalTask.ScheduledDate
		startTime = m.modalTask.ScheduledStart
		endTime = m.modalTask.ScheduledEnd
		duration = m.modalTask.Duration()
		if m.modalTask.IsPast() {
			nameLocked = true
			nameValue = m.modalTask.Description
		}
	}

	descStyle := m.styles.ModalInputStyle
	if nameLocked {
		descStyle = m.styles.ModalInputLockedStyle
	} else {
		input := m.formDesc
		textStyle := m.styles.ModalInputTextStyle
		cursorStyle := textStyle
		if m.formFocus == 0 {
			focusedBg := m.styles.ModalInputFocusedStyle.GetBackground()
			textStyle = textStyle.Background(focusedBg)
			input.PlaceholderStyle = m.styles.ModalPlaceholderStyle.Background(focusedBg)
			cursorStyle = m.styles.ModalInputCursorStyle
			descStyle = m.styles.ModalInputFocusedStyle
		}
		input.TextStyle = textStyle
		input.PromptStyle = textStyle
		input.Cursor.TextStyle = textStyle
		input.Cursor.Style = cursorStyle
		nameValue = input.View()
	}
	durationLabels := make([]string, 0, len(durationOptions))
	for _, d := range durationOptions {
		durationLabels = append(durationLabels, formatDuration(d))
	}
	body := view.RenderTaskFormBody(
		view.TaskFormModel{
			MetaDate:         taskDate.Format("Mon Jan 2"),
			TimeRange:        fmt.Sprintf("%s-%s", startTime, endTime),
			DurationLabel:    formatDuration(duration),
			NameValue:        nameValue,
			NameLocked:       nameLocked,
			DescStyle:        descStyle,
			DurationOptions:  durationLabels,
			ActiveDuration:   m.formDuration,
			ShowDurationHint: m.formFocus == 1,
		},
		view.TaskFormStyles{
			TagStyle:          m.styles.ModalTagStyle,
			BodyStyle:         m.styles.ModalBodyStyle,
			SectionTitleStyle: m.styles.ModalSectionTitleStyle,
			DurationActive:    m.styles.DurationActiveStyle,
			DurationInactive:  m.styles.DurationInactiveStyle,
			HintStyle:         m.styles.ModalHintStyle,
		},
	)

	footer := view.RenderModalButtons(m.modalStyles(), "[Enter] Save", "[Esc] Cancel")
	return view.RenderModalFrame(title, body, footer, m.modalStyles())
}

// renderTaskDetailModal renders the task detail popup.
func (m Model) renderTaskDetailModal() string {
	if m.modalTask == nil {
		return ""
	}
	t := m.modalTask
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
	body := view.RenderTaskDetailBody(
		view.TaskDetailModel{
			Description:   t.Description,
			CategoryIcon:  categoryIcon,
			CategoryLabel: categoryLabel,
			TimeRange:     fmt.Sprintf("%s - %s (%s)", t.ScheduledStart, t.ScheduledEnd, formatDuration(t.Duration())),
			DateLabel:     t.ScheduledDate.Format("Monday, Jan 2, 2006"),
			OutcomeLabel:  outcomeStr,
		},
		view.TaskDetailStyles{
			BodyStyle:  m.styles.ModalBodyStyle,
			LabelStyle: m.styles.ModalLabelStyle,
		},
	)

	var footer string
	if t.IsPast() {
		footer = view.RenderModalButtons(m.modalStyles(), "[o] Outcome", "[Esc] Close")
	} else {
		footer = view.RenderModalButtonsCompact(m.modalStyles(), "[o] Outcome", "[e] Edit", "[x] Cancel", "[Esc] Close")
	}
	return view.RenderModalFrame("Task Details", body, footer, m.modalStyles())
}

// renderConfirmDeleteModal renders the delete confirmation modal.
func (m Model) renderConfirmDeleteModal() string {
	model := view.ConfirmDeleteModel{HasTask: false}
	if m.modalTask != nil {
		model = view.ConfirmDeleteModel{
			Description: m.modalTask.Description,
			TimeRange:   fmt.Sprintf("%s - %s", m.modalTask.ScheduledStart, m.modalTask.ScheduledEnd),
			DateLabel:   m.modalTask.ScheduledDate.Format("Mon Jan 2"),
			HasTask:     true,
		}
	}
	body := view.RenderConfirmDeleteBody(model, view.ConfirmDeleteStyles{
		BodyStyle: m.styles.ModalBodyStyle,
	})
	footer := view.RenderModalButtons(m.modalStyles(), "[y/Enter] Confirm", "[n/Esc] Cancel")
	return view.RenderModalFrame("Confirm Cancel", body, footer, m.modalStyles())
}

// renderPlanResultModal renders the LLM planning result modal.
func (m Model) renderPlanResultModal() string {
	if m.planResult == nil {
		return ""
	}
	issues := make([]string, 0, len(m.planResult.ValidationErrors))
	for _, ve := range m.planResult.ValidationErrors {
		issues = append(issues, ve.Message)
	}
	days := make([]view.PlanResultDay, 0, len(m.planResult.SortedDates))
	for _, dateStr := range m.planResult.SortedDates {
		tasks := m.planResult.TasksByDate[dateStr]
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
		days = append(days, view.PlanResultDay{
			DateLabel: label,
			Lines:     lines,
		})
	}

	summary := fmt.Sprintf("Total: %d tasks", m.planResult.TotalTasks())
	if len(m.planResult.SortedDates) > 1 {
		summary += fmt.Sprintf(" across %d days", len(m.planResult.SortedDates))
	}
	body := view.RenderPlanResultBody(
		view.PlanResultModel{
			IntroMessage:   "Review the draft and amend it before applying.",
			Issues:         issues,
			Warnings:       m.planResult.Warnings,
			Days:           days,
			NoTasks:        m.planResult.TotalTasks() == 0,
			NoTasksMessage: "No tasks proposed.",
			Summary:        summary,
			AmendHint:      "Press m to amend the plan text before applying.",
		},
		view.PlanResultStyles{
			MetaStyle:         m.styles.ModalMetaStyle,
			SectionTitleStyle: m.styles.ModalSectionTitleStyle,
			BodyStyle:         m.styles.ModalBodyStyle,
		},
	)

	var footer string
	if m.planResult.HasValidationErrors() {
		footer = view.RenderModalButtons(m.modalStyles(), "[m] Amend", "[Esc/c] Cancel")
	} else {
		footer = view.RenderModalButtons(m.modalStyles(), "[Enter/a] Apply", "[m] Amend", "[Esc/c] Cancel")
	}
	return view.RenderModalFrame("LLM Draft", body, footer, m.modalStyles())
}

// renderInitModal renders the startup initialization modal.
func (m Model) renderInitModal() string {
	body := view.RenderInitBody(
		view.InitModalModel{
			ConfigPath:    m.initState.ConfigPath,
			DBPath:        m.initState.DBPath,
			ConfigMissing: m.initState.ConfigMissing,
			DBMissing:     m.initState.DBMissing,
			ErrorMessage:  m.initError,
		},
		view.InitModalStyles{
			BodyStyle:  m.styles.ModalBodyStyle,
			LabelStyle: m.styles.ModalLabelStyle,
			HintStyle:  m.styles.ModalHintStyle,
		},
	)
	footer := view.RenderModalButtons(m.modalStyles(), "[Enter] Allow", "[Esc] Quit")
	return view.RenderModalFrame("Initialize Sancho", body, footer, m.modalStyles())
}

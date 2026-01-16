// Package tui provides the terminal user interface for sancho.
package tui

import (
	"fmt"
	"time"

	"github.com/javiermolinar/sancho/internal/task"
	"github.com/javiermolinar/sancho/internal/tui/view"
)

type taskFormModalViewModel struct {
	Title  string
	Model  view.TaskFormModel
	Styles view.TaskFormStyles
}

func (m Model) taskFormModalViewModel() taskFormModalViewModel {
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
		durationLabels = append(durationLabels, view.FormatDuration(d))
	}

	return taskFormModalViewModel{
		Title: title,
		Model: view.TaskFormModel{
			MetaDate:         taskDate.Format("Mon Jan 2"),
			TimeRange:        fmt.Sprintf("%s-%s", startTime, endTime),
			DurationLabel:    view.FormatDuration(duration),
			NameValue:        nameValue,
			NameLocked:       nameLocked,
			DescStyle:        descStyle,
			DurationOptions:  durationLabels,
			ActiveDuration:   m.formDuration,
			ShowDurationHint: m.formFocus == 1,
		},
		Styles: view.TaskFormStyles{
			TagStyle:          m.styles.ModalTagStyle,
			BodyStyle:         m.styles.ModalBodyStyle,
			SectionTitleStyle: m.styles.ModalSectionTitleStyle,
			DurationActive:    m.styles.DurationActiveStyle,
			DurationInactive:  m.styles.DurationInactiveStyle,
			HintStyle:         m.styles.ModalHintStyle,
		},
	}
}

type taskDetailModalViewModel struct {
	Model  view.TaskDetailModel
	Styles view.TaskDetailStyles
	IsPast bool
}

func (m Model) taskDetailModalViewModel() (taskDetailModalViewModel, bool) {
	if m.modalTask == nil {
		return taskDetailModalViewModel{}, false
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

	return taskDetailModalViewModel{
		Model: view.TaskDetailModel{
			Description:   t.Description,
			CategoryIcon:  categoryIcon,
			CategoryLabel: categoryLabel,
			TimeRange:     fmt.Sprintf("%s - %s (%s)", t.ScheduledStart, t.ScheduledEnd, view.FormatDuration(t.Duration())),
			DateLabel:     t.ScheduledDate.Format("Monday, Jan 2, 2006"),
			OutcomeLabel:  outcomeStr,
		},
		Styles: view.TaskDetailStyles{
			BodyStyle:  m.styles.ModalBodyStyle,
			LabelStyle: m.styles.ModalLabelStyle,
		},
		IsPast: t.IsPast(),
	}, true
}

type confirmDeleteModalViewModel struct {
	Model  view.ConfirmDeleteModel
	Styles view.ConfirmDeleteStyles
}

func (m Model) confirmDeleteModalViewModel() confirmDeleteModalViewModel {
	model := view.ConfirmDeleteModel{HasTask: false}
	if m.modalTask != nil {
		model = view.ConfirmDeleteModel{
			Description: m.modalTask.Description,
			TimeRange:   fmt.Sprintf("%s - %s", m.modalTask.ScheduledStart, m.modalTask.ScheduledEnd),
			DateLabel:   m.modalTask.ScheduledDate.Format("Mon Jan 2"),
			HasTask:     true,
		}
	}

	return confirmDeleteModalViewModel{
		Model: model,
		Styles: view.ConfirmDeleteStyles{
			BodyStyle: m.styles.ModalBodyStyle,
		},
	}
}

type planResultModalViewModel struct {
	Model               view.PlanResultModel
	Styles              view.PlanResultStyles
	HasValidationErrors bool
}

func (m Model) planResultModalViewModel() (planResultModalViewModel, bool) {
	if m.planResult == nil {
		return planResultModalViewModel{}, false
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

	return planResultModalViewModel{
		Model: view.PlanResultModel{
			IntroMessage:   "Review the draft and amend it before applying.",
			Issues:         issues,
			Warnings:       m.planResult.Warnings,
			Days:           days,
			NoTasks:        m.planResult.TotalTasks() == 0,
			NoTasksMessage: "No tasks proposed.",
			Summary:        summary,
			AmendHint:      "Press m to amend the plan text before applying.",
		},
		Styles: view.PlanResultStyles{
			MetaStyle:         m.styles.ModalMetaStyle,
			SectionTitleStyle: m.styles.ModalSectionTitleStyle,
			BodyStyle:         m.styles.ModalBodyStyle,
		},
		HasValidationErrors: m.planResult.HasValidationErrors(),
	}, true
}

type initModalViewModel struct {
	Model  view.InitModalModel
	Styles view.InitModalStyles
}

const weekSummaryFallbackWidth = 60

func (m Model) initModalViewModel() initModalViewModel {
	return initModalViewModel{
		Model: view.InitModalModel{
			ConfigPath:    m.initState.ConfigPath,
			DBPath:        m.initState.DBPath,
			ConfigMissing: m.initState.ConfigMissing,
			DBMissing:     m.initState.DBMissing,
			ErrorMessage:  m.initError,
		},
		Styles: view.InitModalStyles{
			BodyStyle:  m.styles.ModalBodyStyle,
			LabelStyle: m.styles.ModalLabelStyle,
			HintStyle:  m.styles.ModalHintStyle,
		},
	}
}

type weekSummaryBodyViewModel struct {
	Lines  []view.WeekSummaryLine
	Styles view.WeekSummaryStyles
	Width  int
}

func (m Model) weekSummaryBodyViewModel() weekSummaryBodyViewModel {
	lines := m.weekSummarySummaryText
	if m.weekSummaryView == weekSummaryViewTasks {
		lines = m.weekSummaryTasksText
	}
	return weekSummaryBodyViewModel{
		Lines: lines,
		Styles: view.WeekSummaryStyles{
			BodyStyle:         m.styles.ModalBodyStyle,
			MetaStyle:         m.styles.ModalMetaStyle,
			SectionTitleStyle: m.styles.ModalSectionTitleStyle,
		},
		Width: view.ModalContentWidth(m.styles.ModalStyle, weekSummaryFallbackWidth),
	}
}

// renderWeekSummaryModal renders the week summary modal.
func (m Model) renderWeekSummaryModal() string {
	if m.weekSummary == nil {
		return ""
	}
	body := m.weekSummaryBody()
	footer := m.weekSummaryFooter()
	return view.RenderModalFrame("Week Summary", body, footer, m.modalStyles())
}

func (m Model) weekSummaryFooter() string {
	return view.WeekSummaryFooter(m.weekSummaryView == weekSummaryViewTasks, m.modalStyles())
}

func (m Model) weekSummaryBody() string {
	vm := m.weekSummaryBodyViewModel()
	return view.RenderWeekSummaryBody(vm.Lines, vm.Styles, vm.Width)
}

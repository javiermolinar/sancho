// Package tui provides the terminal user interface for sancho.
package tui

import "github.com/javiermolinar/sancho/internal/tui/view"

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

	styleSet := m.modalStyleSet()
	return taskFormModalViewModel{
		Title: title,
		Model: view.NewTaskFormModel(view.TaskFormInput{
			Date:             taskDate,
			StartTime:        startTime,
			EndTime:          endTime,
			DurationMinutes:  duration,
			NameValue:        nameValue,
			NameLocked:       nameLocked,
			DescStyle:        descStyle,
			DurationOptions:  durationOptions,
			ActiveDuration:   m.formDuration,
			ShowDurationHint: m.formFocus == 1,
		}),
		Styles: styleSet.TaskFormStyles(),
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
	styleSet := m.modalStyleSet()
	return taskDetailModalViewModel{
		Model:  view.NewTaskDetailModel(m.modalTask),
		Styles: styleSet.TaskDetailStyles(),
		IsPast: m.modalTask.IsPast(),
	}, true
}

type confirmDeleteModalViewModel struct {
	Model  view.ConfirmDeleteModel
	Styles view.ConfirmDeleteStyles
}

func (m Model) confirmDeleteModalViewModel() confirmDeleteModalViewModel {
	styleSet := m.modalStyleSet()
	return confirmDeleteModalViewModel{
		Model:  view.NewConfirmDeleteModel(m.modalTask),
		Styles: styleSet.ConfirmDeleteStyles(),
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

	styleSet := m.modalStyleSet()
	return planResultModalViewModel{
		Model:               view.NewPlanResultModel(m.planResult),
		Styles:              styleSet.PlanResultStyles(),
		HasValidationErrors: m.planResult.HasValidationErrors(),
	}, true
}

type initModalViewModel struct {
	Model  view.InitModalModel
	Styles view.InitModalStyles
}

const weekSummaryFallbackWidth = 60

func (m Model) initModalViewModel() initModalViewModel {
	styleSet := m.modalStyleSet()
	return initModalViewModel{
		Model: view.InitModalModel{
			ConfigPath:    m.initState.ConfigPath,
			DBPath:        m.initState.DBPath,
			ConfigMissing: m.initState.ConfigMissing,
			DBMissing:     m.initState.DBMissing,
			ErrorMessage:  m.initError,
		},
		Styles: styleSet.InitModalStyles(),
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
	styleSet := m.modalStyleSet()
	return weekSummaryBodyViewModel{
		Lines:  lines,
		Styles: styleSet.WeekSummaryStyles(),
		Width:  view.ModalContentWidth(m.styles.ModalStyle, weekSummaryFallbackWidth),
	}
}

func (m Model) modalStyleSet() view.ModalStyleSet {
	return view.ModalStyleSet{
		BodyStyle:             m.styles.ModalBodyStyle,
		MetaStyle:             m.styles.ModalMetaStyle,
		SectionTitleStyle:     m.styles.ModalSectionTitleStyle,
		TagStyle:              m.styles.ModalTagStyle,
		LabelStyle:            m.styles.ModalLabelStyle,
		HintStyle:             m.styles.ModalHintStyle,
		DurationActiveStyle:   m.styles.DurationActiveStyle,
		DurationInactiveStyle: m.styles.DurationInactiveStyle,
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

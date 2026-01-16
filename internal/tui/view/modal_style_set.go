// Package view provides rendering helpers for the TUI.
package view

import "github.com/charmbracelet/lipgloss"

// ModalStyleSet groups modal styles to reduce call-site verbosity.
type ModalStyleSet struct {
	BodyStyle             lipgloss.Style
	MetaStyle             lipgloss.Style
	SectionTitleStyle     lipgloss.Style
	TagStyle              lipgloss.Style
	LabelStyle            lipgloss.Style
	HintStyle             lipgloss.Style
	DurationActiveStyle   lipgloss.Style
	DurationInactiveStyle lipgloss.Style
}

// TaskFormStyles returns the modal styles needed for the task form.
func (s ModalStyleSet) TaskFormStyles() TaskFormStyles {
	return TaskFormStyles{
		TagStyle:          s.TagStyle,
		BodyStyle:         s.BodyStyle,
		SectionTitleStyle: s.SectionTitleStyle,
		DurationActive:    s.DurationActiveStyle,
		DurationInactive:  s.DurationInactiveStyle,
		HintStyle:         s.HintStyle,
	}
}

// TaskDetailStyles returns the modal styles needed for task details.
func (s ModalStyleSet) TaskDetailStyles() TaskDetailStyles {
	return TaskDetailStyles{
		BodyStyle:  s.BodyStyle,
		LabelStyle: s.LabelStyle,
	}
}

// ConfirmDeleteStyles returns the modal styles needed for delete confirmation.
func (s ModalStyleSet) ConfirmDeleteStyles() ConfirmDeleteStyles {
	return ConfirmDeleteStyles{
		BodyStyle: s.BodyStyle,
	}
}

// PlanResultStyles returns the modal styles needed for plan results.
func (s ModalStyleSet) PlanResultStyles() PlanResultStyles {
	return PlanResultStyles{
		MetaStyle:         s.MetaStyle,
		SectionTitleStyle: s.SectionTitleStyle,
		BodyStyle:         s.BodyStyle,
	}
}

// InitModalStyles returns the modal styles needed for initialization.
func (s ModalStyleSet) InitModalStyles() InitModalStyles {
	return InitModalStyles{
		BodyStyle:  s.BodyStyle,
		LabelStyle: s.LabelStyle,
		HintStyle:  s.HintStyle,
	}
}

// WeekSummaryStyles returns the modal styles needed for week summaries.
func (s ModalStyleSet) WeekSummaryStyles() WeekSummaryStyles {
	return WeekSummaryStyles{
		BodyStyle:         s.BodyStyle,
		MetaStyle:         s.MetaStyle,
		SectionTitleStyle: s.SectionTitleStyle,
	}
}

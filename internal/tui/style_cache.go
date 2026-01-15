// Package tui provides the terminal user interface for sancho.
package tui

import "github.com/charmbracelet/lipgloss"

// StyleCache stores width-specific styles to avoid per-cell mutations.
type StyleCache struct {
	TitleBoxStyle          lipgloss.Style
	DayHeader              lipgloss.Style
	DayHeaderToday         lipgloss.Style
	EmptyCell              lipgloss.Style
	Cursor                 lipgloss.Style
	TaskDeep               lipgloss.Style
	TaskShallow            lipgloss.Style
	TaskDeepAlt            lipgloss.Style
	TaskShallowAlt         lipgloss.Style
	TaskPastDeep           lipgloss.Style
	TaskPastShallow        lipgloss.Style
	TaskPastDeepAlt        lipgloss.Style
	TaskPastShallowAlt     lipgloss.Style
	TaskSelected           lipgloss.Style
	TaskMovePreview        lipgloss.Style
	TaskShifted            lipgloss.Style
	TaskCurrentDeep        lipgloss.Style
	TaskCurrentShallow     lipgloss.Style
	TaskCurrentDeepBody    lipgloss.Style
	TaskCurrentShallowBody lipgloss.Style
}

// NewStyleCache precomputes all width-dependent styles for the grid.
func NewStyleCache(styles *Styles, width int) StyleCache {
	contentWidth := max(1, width-1)
	return StyleCache{
		TitleBoxStyle:          styles.TitleStyle.Border(lipgloss.RoundedBorder()).Padding(0, 2),
		DayHeader:              styles.DayHeaderStyleWidth(width),
		DayHeaderToday:         styles.DayHeaderTodayStyleWidth(width),
		EmptyCell:              styles.EmptyCellStyleWidth(width),
		Cursor:                 styles.CursorStyleWidth(width),
		TaskDeep:               styles.TaskDeepStyleWidth(width),
		TaskShallow:            styles.TaskShallowStyleWidth(width),
		TaskDeepAlt:            styles.TaskDeepAltStyleWidth(width),
		TaskShallowAlt:         styles.TaskShallowAltStyleWidth(width),
		TaskPastDeep:           styles.TaskPastDeepStyleWidth(width),
		TaskPastShallow:        styles.TaskPastShallowStyleWidth(width),
		TaskPastDeepAlt:        styles.TaskPastDeepAltStyleWidth(width),
		TaskPastShallowAlt:     styles.TaskPastShallowAltStyleWidth(width),
		TaskSelected:           styles.TaskSelectedStyleWidth(width),
		TaskMovePreview:        styles.TaskMovePreviewStyleWidth(width),
		TaskShifted:            styles.TaskShiftedStyleWidth(width),
		TaskCurrentDeep:        styles.TaskCurrentStyleWidth(width, true),
		TaskCurrentShallow:     styles.TaskCurrentStyleWidth(width, false),
		TaskCurrentDeepBody:    styles.TaskCurrentStyleWidth(contentWidth, true),
		TaskCurrentShallowBody: styles.TaskCurrentStyleWidth(contentWidth, false),
	}
}

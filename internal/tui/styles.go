// Package tui provides the terminal user interface for sancho.
package tui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/javiermolinar/sancho/internal/tui/theme"
)

// Default column width - will be recalculated dynamically.
const defaultColWidth = 18

// Styles holds all lipgloss styles for the TUI, derived from a theme.
type Styles struct {
	// Theme colors as lipgloss colors
	colorBg          lipgloss.Color
	colorBgHighlight lipgloss.Color
	colorBgSelection lipgloss.Color
	colorFg          lipgloss.Color
	colorFgMuted     lipgloss.Color
	colorAccent      lipgloss.Color
	colorDeep        lipgloss.Color
	colorShallow     lipgloss.Color
	colorCurrent     lipgloss.Color
	colorWarning     lipgloss.Color

	colorTextOnAccent  lipgloss.Color
	colorTextOnWarning lipgloss.Color
	colorTextOnCurrent lipgloss.Color
	colorTextOnDeep    lipgloss.Color
	colorTextOnShallow lipgloss.Color

	// Derived darker colors for task backgrounds
	colorDeepBg    lipgloss.Color
	colorShallowBg lipgloss.Color

	// Alternate shade colors for adjacent same-category tasks
	colorDeepBgAlt    lipgloss.Color
	colorShallowBgAlt lipgloss.Color

	// Even more muted colors for past tasks (still shows deep vs shallow)
	colorDeepPastBg    lipgloss.Color
	colorShallowPastBg lipgloss.Color

	// Title style
	TitleStyle lipgloss.Style

	// Header styles
	HeaderStyle         lipgloss.Style
	DayHeaderStyle      lipgloss.Style
	DayHeaderTodayStyle lipgloss.Style

	// Time column
	TimeColumnStyle lipgloss.Style

	// Task cell styles
	TaskCellStyle           lipgloss.Style
	TaskDeepStyle           lipgloss.Style
	TaskShallowStyle        lipgloss.Style
	TaskDeepAltStyle        lipgloss.Style // Alternate shade for adjacent deep tasks
	TaskShallowAltStyle     lipgloss.Style // Alternate shade for adjacent shallow tasks
	TaskPastDeepStyle       lipgloss.Style // Past deep work (muted but visible)
	TaskPastShallowStyle    lipgloss.Style // Past shallow work (muted but visible)
	TaskPastDeepAltStyle    lipgloss.Style // Past deep work alternate shade
	TaskPastShallowAltStyle lipgloss.Style // Past shallow work alternate shade
	TaskSelectedStyle       lipgloss.Style
	TaskMovePreviewStyle    lipgloss.Style
	TaskShiftedStyle        lipgloss.Style // Tasks shifted to make room during move
	TaskCurrentStyle        lipgloss.Style // Current task (time-based)

	// Current task accent (left border indicator)
	CurrentAccentStyle lipgloss.Style

	// Empty cell
	EmptyCellStyle lipgloss.Style

	// Cursor style
	CursorStyle lipgloss.Style

	// Stats bar
	StatsBarStyle     lipgloss.Style
	StatsDeepStyle    lipgloss.Style
	StatsShallowStyle lipgloss.Style

	// Prompt box
	PromptStyle        lipgloss.Style
	PromptFocusedStyle lipgloss.Style

	// Status message
	StatusStyle lipgloss.Style

	// Help text
	HelpStyle lipgloss.Style

	// Modal styles
	ModalStyle             lipgloss.Style
	ModalBgColor           lipgloss.Color
	ModalBackdropColor     lipgloss.Color
	ModalHeaderStyle       lipgloss.Style
	ModalFooterStyle       lipgloss.Style
	ModalTitleStyle        lipgloss.Style
	ModalBodyStyle         lipgloss.Style
	ModalMetaStyle         lipgloss.Style
	ModalSectionTitleStyle lipgloss.Style
	ModalTagStyle          lipgloss.Style
	ModalLabelStyle        lipgloss.Style
	ModalInputStyle        lipgloss.Style
	ModalInputFocusedStyle lipgloss.Style
	ModalInputLockedStyle  lipgloss.Style
	ModalInputTextStyle    lipgloss.Style
	ModalInputCursorStyle  lipgloss.Style
	ModalPlaceholderStyle  lipgloss.Style
	ModalButtonStyle       lipgloss.Style
	ModalButtonActiveStyle lipgloss.Style
	ModalHintStyle         lipgloss.Style

	// Category toggle styles
	CategoryActiveStyle   lipgloss.Style
	CategoryInactiveStyle lipgloss.Style

	// Duration option styles
	DurationActiveStyle   lipgloss.Style
	DurationInactiveStyle lipgloss.Style

	// Table container
	TableStyle lipgloss.Style

	// App container
	AppStyle lipgloss.Style

	// Viewport background
	ViewportStyle lipgloss.Style

	// Separator style
	SeparatorStyle lipgloss.Style
}

// NewStyles creates a new Styles instance from a theme.
func NewStyles(t *theme.Theme) *Styles {
	s := &Styles{}
	palette := theme.NewPalette(t)

	// Convert theme colors to lipgloss colors
	s.colorBg = palette.Bg
	s.colorBgHighlight = palette.BgHighlight
	s.colorBgSelection = palette.BgSelection
	s.colorFg = palette.Fg
	s.colorFgMuted = palette.FgMuted
	s.colorAccent = palette.Accent
	s.colorDeep = palette.Deep
	s.colorShallow = palette.Shallow
	s.colorCurrent = palette.Current
	s.colorWarning = palette.Warning

	s.colorTextOnAccent = palette.TextOnAccent
	s.colorTextOnWarning = palette.TextOnWarning
	s.colorTextOnCurrent = palette.TextOnCurrent
	s.colorTextOnDeep = palette.TextOnDeep
	s.colorTextOnShallow = palette.TextOnShallow

	// Create darker versions of task colors for backgrounds
	s.colorDeepBg = palette.DeepBg
	s.colorShallowBg = palette.ShallowBg

	// Create alternate shade for adjacent same-category tasks (lighter version)
	s.colorDeepBgAlt = palette.DeepBgAlt
	s.colorShallowBgAlt = palette.ShallowBgAlt

	// Create even more muted versions for past tasks (30% brightness)
	// Still distinguishes deep vs shallow but clearly shows task is past
	s.colorDeepPastBg = palette.DeepPastBg
	s.colorShallowPastBg = palette.ShallowPastBg

	// Build styles from colors

	// Title style
	s.TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(s.colorAccent).
		Background(s.colorBg)

	// Header styles
	s.HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(s.colorAccent).
		Background(s.colorBg)

	// Day column header
	s.DayHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Align(lipgloss.Center).
		Foreground(s.colorFg).
		Background(s.colorBg).
		Width(defaultColWidth)

	s.DayHeaderTodayStyle = s.DayHeaderStyle.
		Foreground(s.colorAccent).
		Bold(true)

	// Time column
	s.TimeColumnStyle = lipgloss.NewStyle().
		Foreground(s.colorAccent).
		Background(s.colorBg).
		Width(6)

	// Task cell styles
	s.TaskCellStyle = lipgloss.NewStyle().
		Width(defaultColWidth).
		Align(lipgloss.Left)

	// Deep work: darker background with white text for readability
	s.TaskDeepStyle = s.TaskCellStyle.
		Background(s.colorDeepBg).
		Foreground(s.colorFg).
		Bold(true)

	// Shallow work: darker background with white text for readability
	s.TaskShallowStyle = s.TaskCellStyle.
		Background(s.colorShallowBg).
		Foreground(s.colorFg).
		Bold(true)

	// Alternate deep work: lighter shade for adjacent same-category tasks
	s.TaskDeepAltStyle = s.TaskCellStyle.
		Background(s.colorDeepBgAlt).
		Foreground(s.colorFg).
		Bold(true)

	// Alternate shallow work: lighter shade for adjacent same-category tasks
	s.TaskShallowAltStyle = s.TaskCellStyle.
		Background(s.colorShallowBgAlt).
		Foreground(s.colorFg).
		Bold(true)

	// Past deep work: muted background but readable text
	// Use bright foreground for proper contrast
	s.TaskPastDeepStyle = s.TaskCellStyle.
		Background(s.colorDeepPastBg).
		Foreground(s.colorFg)

	// Past shallow work: muted background but readable text
	// Use bright foreground for proper contrast
	s.TaskPastShallowStyle = s.TaskCellStyle.
		Background(s.colorShallowPastBg).
		Foreground(s.colorFg)

	// Past deep work alternate: lighter muted shade
	s.TaskPastDeepAltStyle = s.TaskCellStyle.
		Background(palette.DeepPastBgAlt).
		Foreground(s.colorFg)

	// Past shallow work alternate: lighter muted shade
	s.TaskPastShallowAltStyle = s.TaskCellStyle.
		Background(palette.ShallowPastBgAlt).
		Foreground(s.colorFg)

	s.TaskSelectedStyle = s.TaskCellStyle.
		Background(s.colorWarning).
		Foreground(s.colorTextOnWarning).
		Bold(true)

	s.TaskMovePreviewStyle = s.TaskCellStyle.
		Background(s.colorAccent).
		Foreground(s.colorTextOnAccent).
		Bold(true)

	// Shifted task style - tasks that have been moved to make room
	// Use a dimmed version of the background selection color
	s.TaskShiftedStyle = s.TaskCellStyle.
		Background(s.colorBgSelection).
		Foreground(s.colorFg).
		Italic(true)

	// Current task style - bright background to stand out
	// Note: Avoid borders as they break grid layout
	s.TaskCurrentStyle = s.TaskCellStyle.
		Background(s.colorCurrent).
		Foreground(s.colorTextOnCurrent).
		Bold(true)

	// Current task accent (left border indicator in gold/yellow)
	s.CurrentAccentStyle = lipgloss.NewStyle().
		Foreground(s.colorCurrent)

	// Empty cell
	s.EmptyCellStyle = lipgloss.NewStyle().
		Width(defaultColWidth).
		Foreground(s.colorFgMuted).
		Background(s.colorBg)

	// Cursor style - use warning color for high visibility
	s.CursorStyle = lipgloss.NewStyle().
		Width(defaultColWidth).
		Background(s.colorBgSelection).
		Foreground(s.colorAccent).
		Bold(true)

	// Stats bar - no margins, use explicit newlines in View() for spacing
	s.StatsBarStyle = lipgloss.NewStyle().
		Foreground(s.colorFg).
		Background(s.colorBg).
		Padding(0, 0)

	s.StatsDeepStyle = lipgloss.NewStyle().
		Foreground(s.colorDeep).
		Background(s.colorBg).
		Bold(true)

	s.StatsShallowStyle = lipgloss.NewStyle().
		Foreground(s.colorShallow).
		Background(s.colorBg).
		Bold(true)

	// Prompt box
	s.PromptStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.colorFgMuted).
		BorderBackground(s.colorBg).
		Background(s.colorBgHighlight).
		Foreground(s.colorFg).
		Padding(0, 1)

	s.PromptFocusedStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.colorAccent).
		BorderBackground(s.colorBg).
		Background(s.colorBgSelection).
		Foreground(s.colorFg).
		Bold(true).
		Padding(0, 1)

	// Status message
	s.StatusStyle = lipgloss.NewStyle().
		Foreground(s.colorWarning).
		Background(s.colorBg).
		Bold(true)

	// Help text
	s.HelpStyle = lipgloss.NewStyle().
		Foreground(s.colorFg).
		Background(s.colorBg)

	// Modal styles - use high-contrast theme colors
	modal := palette.Modal
	modalBg := modal.Bg
	modalBorder := modal.Border
	modalText := modal.Text
	modalMuted := modal.Muted
	modalHighlight := modal.Highlight
	modalPanel := modal.Panel
	modalReverseText := modal.ReverseText
	placeholderColor := modalMuted
	s.ModalBackdropColor = modal.Backdrop
	s.ModalBgColor = modalBg

	s.ModalStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(modalBorder).
		Background(modalBg).
		Foreground(modalText).
		Padding(1, 1).
		Width(72).
		Align(lipgloss.Left)

	s.ModalHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(modalText).
		Background(modalBg).
		Padding(0, 1).
		Align(lipgloss.Center)

	s.ModalFooterStyle = lipgloss.NewStyle().
		Padding(0, 1).
		Background(modalBg)

	s.ModalTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(modalText).
		Background(modalBg)

	s.ModalBodyStyle = lipgloss.NewStyle().
		Foreground(modalText).
		Background(modalBg)

	s.ModalMetaStyle = lipgloss.NewStyle().
		Foreground(modalMuted).
		Background(modalBg)

	s.ModalSectionTitleStyle = lipgloss.NewStyle().
		Foreground(modalText).
		Bold(true).
		PaddingLeft(1).
		Background(modalBg)

	s.ModalTagStyle = lipgloss.NewStyle().
		Foreground(modalText).
		Background(modalPanel).
		Bold(true).
		Padding(0, 1)

	s.ModalLabelStyle = lipgloss.NewStyle().
		Foreground(modalText).
		Bold(true).
		Width(12).
		Background(modalBg)

	s.ModalInputStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(modalBorder).
		Background(modalBg).
		Foreground(modalText).
		Padding(0, 1).
		Width(54)

	s.ModalInputFocusedStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(modalHighlight).
		Background(modalPanel).
		Foreground(modalText).
		Padding(0, 1).
		Width(54)

	s.ModalInputLockedStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(modalMuted).
		Background(modalBg).
		Foreground(modalMuted).
		Padding(0, 1).
		Width(54)

	s.ModalInputTextStyle = lipgloss.NewStyle().
		Foreground(modalText).
		Background(modalBg)

	s.ModalInputCursorStyle = lipgloss.NewStyle().
		Foreground(modalReverseText).
		Background(modalHighlight)

	s.ModalPlaceholderStyle = lipgloss.NewStyle().
		Foreground(placeholderColor).
		Background(modalBg)

	s.ModalButtonStyle = lipgloss.NewStyle().
		Background(modalPanel).
		Foreground(modalText).
		Padding(0, 3)

	s.ModalButtonActiveStyle = lipgloss.NewStyle().
		Background(modalHighlight).
		Foreground(modalReverseText).
		Padding(0, 3).
		MarginRight(0).
		Underline(true)

	s.ModalHintStyle = lipgloss.NewStyle().
		Foreground(modalMuted).
		Background(modalBg)

	// Category toggle styles
	s.CategoryActiveStyle = lipgloss.NewStyle().
		Background(s.colorDeep).
		Foreground(s.colorTextOnDeep).
		Bold(true).
		Padding(0, 1)

	s.CategoryInactiveStyle = lipgloss.NewStyle().
		Background(s.colorBgHighlight).
		Foreground(s.colorFgMuted).
		Padding(0, 1)

	// Duration option styles
	s.DurationActiveStyle = lipgloss.NewStyle().
		Background(modalHighlight).
		Foreground(modalReverseText).
		Bold(true).
		Padding(0, 1)

	s.DurationInactiveStyle = lipgloss.NewStyle().
		Background(modalBg).
		Foreground(modalMuted).
		Padding(0, 1)

	// Table container - border and internal padding only
	s.TableStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.colorAccent).
		Background(s.colorBg).
		Padding(0, 1)

	// App container - padding provides consistent indentation for all content
	s.AppStyle = lipgloss.NewStyle().
		Background(s.colorBg).
		PaddingTop(1).
		PaddingLeft(2).
		PaddingRight(2).
		PaddingBottom(1)

	// Viewport background - fill entire terminal with base background.
	s.ViewportStyle = lipgloss.NewStyle().
		Background(s.colorBg)

	// Separator style
	s.SeparatorStyle = lipgloss.NewStyle().
		Foreground(s.colorBgSelection).
		Background(s.colorBg)

	return s
}

// WithWidth returns a copy of the style with the specified width.
func (s *Styles) TaskDeepStyleWidth(width int) lipgloss.Style {
	return s.TaskDeepStyle.Width(width)
}

// TaskShallowStyleWidth returns the shallow task style with specified width.
func (s *Styles) TaskShallowStyleWidth(width int) lipgloss.Style {
	return s.TaskShallowStyle.Width(width)
}

// TaskDeepAltStyleWidth returns the alternate deep task style with specified width.
func (s *Styles) TaskDeepAltStyleWidth(width int) lipgloss.Style {
	return s.TaskDeepAltStyle.Width(width)
}

// TaskShallowAltStyleWidth returns the alternate shallow task style with specified width.
func (s *Styles) TaskShallowAltStyleWidth(width int) lipgloss.Style {
	return s.TaskShallowAltStyle.Width(width)
}

// TaskPastDeepStyleWidth returns the past deep task style with specified width.
func (s *Styles) TaskPastDeepStyleWidth(width int) lipgloss.Style {
	return s.TaskPastDeepStyle.Width(width)
}

// TaskPastShallowStyleWidth returns the past shallow task style with specified width.
func (s *Styles) TaskPastShallowStyleWidth(width int) lipgloss.Style {
	return s.TaskPastShallowStyle.Width(width)
}

// TaskPastDeepAltStyleWidth returns the alternate past deep task style with specified width.
func (s *Styles) TaskPastDeepAltStyleWidth(width int) lipgloss.Style {
	return s.TaskPastDeepAltStyle.Width(width)
}

// TaskPastShallowAltStyleWidth returns the alternate past shallow task style with specified width.
func (s *Styles) TaskPastShallowAltStyleWidth(width int) lipgloss.Style {
	return s.TaskPastShallowAltStyle.Width(width)
}

// TaskSelectedStyleWidth returns the selected task style with specified width.
func (s *Styles) TaskSelectedStyleWidth(width int) lipgloss.Style {
	return s.TaskSelectedStyle.Width(width)
}

// TaskMovePreviewStyleWidth returns the move preview style with specified width.
func (s *Styles) TaskMovePreviewStyleWidth(width int) lipgloss.Style {
	return s.TaskMovePreviewStyle.Width(width)
}

// TaskShiftedStyleWidth returns the shifted task style with specified width.
func (s *Styles) TaskShiftedStyleWidth(width int) lipgloss.Style {
	return s.TaskShiftedStyle.Width(width)
}

// EmptyCellStyleWidth returns the empty cell style with specified width.
func (s *Styles) EmptyCellStyleWidth(width int) lipgloss.Style {
	return s.EmptyCellStyle.Width(width)
}

// CursorStyleWidth returns the cursor style with specified width.
func (s *Styles) CursorStyleWidth(width int) lipgloss.Style {
	return s.CursorStyle.Width(width)
}

// DayHeaderStyleWidth returns the day header style with specified width.
func (s *Styles) DayHeaderStyleWidth(width int) lipgloss.Style {
	return s.DayHeaderStyle.Width(width)
}

// DayHeaderTodayStyleWidth returns the today header style with specified width.
func (s *Styles) DayHeaderTodayStyleWidth(width int) lipgloss.Style {
	return s.DayHeaderTodayStyle.Width(width)
}

// TaskCurrentStyleWidth returns the current task style with specified width.
// Uses brighter category backgrounds to mark the current task across all lines.
// Note: We avoid borders because they add extra lines that break the grid layout.
func (s *Styles) TaskCurrentStyleWidth(width int, _ bool) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Background(s.colorCurrent).
		Foreground(s.colorTextOnCurrent).
		Bold(true)
}

// TaskSeparatorColor returns the color used for task separators.
func (s *Styles) TaskSeparatorColor() lipgloss.Color {
	return s.colorFgMuted
}

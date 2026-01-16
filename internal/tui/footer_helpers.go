package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/javiermolinar/sancho/internal/tui/view"
)

// statusMsgOrDefault returns the status message or a space to preserve layout.
func (m Model) statusMsgOrDefault() string {
	if m.statusMsg == "" {
		return " "
	}
	return m.statusMsg
}

// renderStatsBar renders the statistics bar.
func (m Model) renderStatsBar(width int) string {
	ww := m.slotState.WeekWindow()
	if ww == nil || ww.Current() == nil {
		return ""
	}

	week := ww.Current()
	stats := week.Stats()
	dayStats := stats.DayStats[m.cursor.Day]
	dayDeep := view.FormatDuration(dayStats.DeepMinutes)
	dayShallow := view.FormatDuration(dayStats.ShallowMinutes)
	weekTotal := view.FormatDuration(stats.TotalMinutes())
	weekDeep := stats.DeepPercent()

	pending := 0
	done := 0
	for _, d := range week.Days {
		for _, t := range d.Tasks() {
			if t.IsScheduled() {
				if t.IsPast() {
					done++
				} else {
					pending++
				}
			}
		}
	}

	loadingIndicator := ""
	if m.loading {
		loadingIndicator = " [Loading...]"
	}

	editIndicator := ""
	if m.slotState.IsEditing() {
		if m.slotState.HasChanges() {
			editIndicator = " [EDIT*]"
		} else {
			editIndicator = " [EDIT]"
		}
	}

	barStyle := lipgloss.NewStyle().
		Foreground(m.styles.colorFg).
		Background(m.styles.colorBg)
	deepStyle := lipgloss.NewStyle().
		Foreground(m.styles.colorDeep).
		Background(m.styles.colorBg).
		Bold(true)
	shallowStyle := lipgloss.NewStyle().
		Foreground(m.styles.colorShallow).
		Background(m.styles.colorBg).
		Bold(true)

	var bar strings.Builder
	bar.WriteString(barStyle.Render("Day: "))
	bar.WriteString(deepStyle.Render(dayDeep))
	bar.WriteString(barStyle.Render(" deep, "))
	bar.WriteString(shallowStyle.Render(dayShallow))
	bar.WriteString(barStyle.Render(" shallow | Week: "))
	bar.WriteString(barStyle.Render(weekTotal))
	bar.WriteString(barStyle.Render(" total, "))
	bar.WriteString(barStyle.Render(fmt.Sprintf("%d%% deep | ", weekDeep)))
	bar.WriteString(barStyle.Render(fmt.Sprintf("%d pending, %d done", pending, done)))
	if loadingIndicator != "" {
		bar.WriteString(barStyle.Render(loadingIndicator))
	}
	if editIndicator != "" {
		bar.WriteString(barStyle.Render(editIndicator))
	}

	statsStyle := m.layoutCache.StatsBarStyle
	frameW, _ := statsStyle.GetFrameSize()
	contentWidth := max(0, width-frameW)
	statsStyle = statsStyle.Width(contentWidth)
	content := bar.String()
	if contentWidth > 0 {
		content = ansi.Truncate(content, contentWidth, "")
	}
	return statsStyle.Render(content)
}

// renderLegend renders the legend for task categories.
func (m Model) renderLegend() string {
	baseStyle := lipgloss.NewStyle().
		Foreground(m.styles.colorFg).
		Background(m.styles.colorBg)
	deepLabelStyle := baseStyle.
		Foreground(m.styles.colorDeep).
		Bold(true)
	shallowLabelStyle := baseStyle.
		Foreground(m.styles.colorShallow).
		Bold(true)

	var legend strings.Builder
	legend.WriteString(baseStyle.Render("Legend: "))
	legend.WriteString(deepLabelStyle.Render("[D] Deep"))
	legend.WriteString(baseStyle.Render("  "))
	legend.WriteString(shallowLabelStyle.Render("[S] Shallow"))
	return legend.String()
}

// promptCursor returns the cursor character if in prompt mode.
func (m Model) promptCursor() string {
	if m.mode == ModePrompt {
		return "_"
	}
	return ""
}

// renderHelp renders the help bar.
func (m Model) renderHelp() string {
	var help string
	switch m.mode {
	case ModeEdit:
		undoInfo := ""
		if m.slotState.CanUndo() {
			undoInfo = fmt.Sprintf(" (%d)", m.slotState.UndoCount())
		}
		help = fmt.Sprintf("EDIT: g/s/Space/x: modify | y: move | u: undo%s | Enter: save | Esc: discard", undoInfo)
	case ModeMove:
		help = "h/j/k/l: navigate | Enter: confirm | Esc: cancel"
	case ModePrompt:
		help = "Enter: submit | Esc: cancel"
	case ModeModal:
		switch m.modalType {
		case ModalTaskForm:
			help = "Tab: next field | Enter: save | Esc: cancel"
		case ModalTaskDetail:
			if m.modalTask != nil && m.modalTask.IsPast() {
				help = "o: outcome | Enter/Esc: close"
			} else {
				help = "o: outcome | e: edit task | x: cancel task | Enter/Esc: close"
			}
		case ModalConfirmDelete:
			help = "y/Enter: confirm | n/Esc: cancel"
		case ModalPlanResult:
			help = "a/Enter: apply | m: amend | c/Esc: cancel"
		default:
			help = "Esc: close"
		}
	default:
		help = "h/j/k/l: navigate | i: edit mode | d: defer | /: commands | q: quit"
	}
	return m.styles.HelpStyle.Render(help)
}

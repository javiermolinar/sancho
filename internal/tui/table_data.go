package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/javiermolinar/sancho/internal/task"
)

func (m Model) visibleSlotsForTable(height int) int {
	if height <= 0 || m.rowLines <= 0 {
		return 0
	}

	tableChrome := 4 // top border + header + header separator + bottom border
	if height <= tableChrome {
		return 0
	}

	rows := (height - tableChrome) / m.rowLines
	if rows < 0 {
		return 0
	}
	return rows
}

func (m Model) buildGridTableRows(visibleSlots int) ([][]string, [][]lipgloss.Style) {
	if visibleSlots <= 0 {
		return nil, nil
	}

	rows := make([][]string, 0, visibleSlots)
	cellStyles := make([][]lipgloss.Style, 0, visibleSlots)

	shadeByDay := m.cachedShadeMap
	cursorTask := m.cachedCursorTask()

	for i := 0; i < visibleSlots; i++ {
		slot := m.scrollOffset + i
		row := make([]string, 0, 8)
		rowStyles := make([]lipgloss.Style, 0, 8)

		timeLabel := ""
		if slot >= 0 && slot < m.maxSlots() {
			timeLabel = minutesToTime(m.dayStartMinutes() + (slot * m.rowHeight))
		}

		row = append(row, m.timeColumnContent(timeLabel))
		rowStyles = append(rowStyles, m.timeColumnStyle())

		for day := 0; day < 7; day++ {
			dayTasks := m.gridCache[day]
			var t *task.Task
			if slot >= 0 && slot < len(dayTasks) {
				t = dayTasks[slot]
			}

			style, lines := m.cellStyleAndLines(day, slot, t, dayTasks, cursorTask, shadeByDay)
			row = append(row, strings.Join(lines, "\n"))
			rowStyles = append(rowStyles, style)
		}

		rows = append(rows, row)
		cellStyles = append(cellStyles, rowStyles)
	}

	return rows, cellStyles
}

func (m Model) timeColumnStyle() lipgloss.Style {
	return m.styles.TimeColumnStyle.Width(6).Height(m.rowLines)
}

func (m Model) timeColumnContent(label string) string {
	lines := make([]string, m.rowLines)
	if len(lines) > 0 {
		lines[0] = padRight(label, 6)
	}
	return strings.Join(lines, "\n")
}

func (m Model) cellStyleAndLines(
	day, slot int,
	t *task.Task,
	dayTasks []*task.Task,
	cursorTask *task.Task,
	shadeByDay map[int]map[int64]bool,
) (lipgloss.Style, []string) {
	style, isCursor, isPartOfCursorTask := m.cellStyleForSlot(day, slot, t, cursorTask, shadeByDay)

	lines := m.cellContentLines(slot, t, dayTasks)
	if t == nil && isCursor && !isPartOfCursorTask && len(lines) > 0 {
		lines[0] = ">"
	}

	for i := range lines {
		if lines[i] != "" {
			lines[i] = " " + lines[i]
		}
	}

	style = style.Width(m.colWidth).Height(m.rowLines)
	return style, lines
}

func (m Model) cellContentLines(slot int, t *task.Task, dayTasks []*task.Task) []string {
	lines := make([]string, m.rowLines)
	if t == nil || m.rowLines <= 0 {
		return lines
	}

	indicator := "S"
	if t.IsDeep() {
		indicator = "D"
	}
	descLines := m.cachedTaskLines[t.ID]

	startSlot := slot
	for startSlot > 0 {
		prevTask := dayTasks[startSlot-1]
		if prevTask == nil || prevTask.ID != t.ID {
			break
		}
		startSlot--
	}
	endSlot := slot
	for endSlot+1 < len(dayTasks) {
		nextTask := dayTasks[endSlot+1]
		if nextTask == nil || nextTask.ID != t.ID {
			break
		}
		endSlot++
	}
	slotIndex := slot - startSlot
	totalSlots := endSlot - startSlot + 1
	if totalSlots < 1 || slotIndex >= totalSlots {
		return lines
	}
	totalLines := totalSlots * m.rowLines
	if totalLines <= 0 {
		return lines
	}

	if m.rowLines == 1 {
		if totalSlots == 1 {
			lines[0] = m.singleLineTaskContent(indicator, t)
			return lines
		}
		lineIndex := slotIndex
		timeIndex := len(descLines)
		switch lineIndex {
		case 0:
			if len(descLines) > 0 {
				lines[0] = "[" + indicator + "] " + descLines[0]
			}
		case timeIndex:
			lines[0] = t.ScheduledStart + "-" + t.ScheduledEnd
		default:
			if lineIndex > 0 && lineIndex < timeIndex {
				lines[0] = descLines[lineIndex]
			}
		}
		return lines
	}

	timeIndex := len(descLines)
	startLineIndex := slotIndex * m.rowLines
	for i := 0; i < m.rowLines; i++ {
		lineIndex := startLineIndex + i
		switch lineIndex {
		case 0:
			if len(descLines) > 0 {
				lines[i] = "[" + indicator + "] " + descLines[0]
			}
		case timeIndex:
			lines[i] = t.ScheduledStart + "-" + t.ScheduledEnd
		default:
			if lineIndex > 0 && lineIndex < timeIndex {
				lines[i] = descLines[lineIndex]
			}
		}
	}

	return lines
}

func (m Model) cellStyleForSlot(
	day, slot int,
	t *task.Task,
	cursorTask *task.Task,
	shadeByDay map[int]map[int64]bool,
) (lipgloss.Style, bool, bool) {
	isCursor := m.cursor.Day == day && m.cursor.Slot == slot
	isPartOfCursorTask := cursorTask != nil && t != nil && t.ID == cursorTask.ID

	style := m.styleCache.EmptyCell
	if t != nil {
		isCurrent := m.isCurrentTask(t)
		useAltShade := false
		if dayShade := shadeByDay[day]; dayShade != nil {
			useAltShade = dayShade[t.ID]
		}

		switch {
		case t.IsPast():
			if t.IsDeep() {
				if useAltShade {
					style = m.styleCache.TaskPastDeepAlt
				} else {
					style = m.styleCache.TaskPastDeep
				}
			} else {
				if useAltShade {
					style = m.styleCache.TaskPastShallowAlt
				} else {
					style = m.styleCache.TaskPastShallow
				}
			}
		case isCurrent:
			if t.IsDeep() {
				style = m.styleCache.TaskCurrentDeep
			} else {
				style = m.styleCache.TaskCurrentShallow
			}
		case t.IsDeep():
			if useAltShade {
				style = m.styleCache.TaskDeepAlt
			} else {
				style = m.styleCache.TaskDeep
			}
		default:
			if useAltShade {
				style = m.styleCache.TaskShallowAlt
			} else {
				style = m.styleCache.TaskShallow
			}
		}
	}

	if isCursor || isPartOfCursorTask {
		if m.mode == ModeMove {
			style = m.styleCache.TaskMovePreview
		} else {
			style = m.styleCache.Cursor
		}
	}

	movingTask := m.slotState.MovingTask()
	if m.mode == ModeMove && t != nil && movingTask != nil {
		if t.ID == movingTask.ID {
			style = m.styleCache.TaskSelected
		} else if m.isTaskShifted(t) {
			style = m.styleCache.TaskShifted
		}
	}

	return style, isCursor, isPartOfCursorTask
}

func (m Model) taskShadeMap() map[int]map[int64]bool {
	if m.maxSlots() <= 0 {
		return nil
	}

	shadeByDay := make(map[int]map[int64]bool, 7)
	for day := 0; day < 7; day++ {
		dayTasks := m.gridCache[day]
		if len(dayTasks) == 0 {
			continue
		}
		dayShade := make(map[int64]bool)
		var lastTask *task.Task
		lastAlt := false

		maxSlots := min(m.maxSlots(), len(dayTasks))
		for slot := 0; slot < maxSlots; slot++ {
			t := dayTasks[slot]
			if t == nil {
				lastTask = nil
				lastAlt = false
				continue
			}

			if lastTask != nil && t.ID == lastTask.ID {
				continue
			}

			useAlt := false
			if lastTask != nil {
				sameCategory := t.IsDeep() == lastTask.IsDeep()
				if sameCategory {
					useAlt = !lastAlt
				}
			}

			dayShade[t.ID] = useAlt
			lastTask = t
			lastAlt = useAlt
		}

		if len(dayShade) > 0 {
			shadeByDay[day] = dayShade
		}
	}

	return shadeByDay
}

// renderCell renders a single cell line for tests and legacy callers.
func (m Model) renderCell(
	slot, line int,
	t *task.Task,
	dayTasks []*task.Task,
	cursorTask *task.Task,
	shadeByDay map[int]map[int64]bool,
) string {
	style, lines := m.cellStyleAndLines(0, slot, t, dayTasks, cursorTask, shadeByDay)
	if line < 0 || line >= len(lines) {
		return style.Render(" ")
	}
	content := lines[line]
	if content == "" {
		return style.Render(" ")
	}
	return style.Render(content)
}

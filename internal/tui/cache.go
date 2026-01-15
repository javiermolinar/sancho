package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/javiermolinar/sancho/internal/task"
)

func (m *Model) markCacheDirty() {
	m.cacheNeedsUpdate = true
}

func (m *Model) refreshViewCaches() {
	m.refreshGridCache()
	m.cachedShadeMap = m.taskShadeMap()
	m.cachedTaskLines = m.buildTaskLines()
	m.refreshRenderCache()
	m.cacheNeedsUpdate = false
}

func (m *Model) refreshCachesIfNeeded() {
	if m.cacheNeedsUpdate {
		m.refreshViewCaches()
	}
}

func (m *Model) buildTaskLines() map[int64][]string {
	lines := make(map[int64][]string)
	if m.colWidth <= 0 {
		return lines
	}

	spanByID := make(map[int64]int)
	for day := 0; day < 7; day++ {
		dayTasks := m.gridCache[day]
		if len(dayTasks) == 0 {
			continue
		}
		for slot := 0; slot < len(dayTasks); {
			t := dayTasks[slot]
			if t == nil {
				slot++
				continue
			}
			runID := t.ID
			start := slot
			for slot < len(dayTasks) && dayTasks[slot] != nil && dayTasks[slot].ID == runID {
				slot++
			}
			span := slot - start
			if span > spanByID[runID] {
				spanByID[runID] = span
			}
		}
	}

	firstWidth := max(1, m.colWidth-5)
	otherWidth := max(1, m.colWidth-1)
	for _, t := range m.slotState.AllTasks() {
		if t == nil {
			continue
		}
		if _, exists := lines[t.ID]; exists {
			continue
		}
		span := spanByID[t.ID]
		if span <= 0 {
			span = m.taskSlotSpan(t)
		}
		maxLines := (span * m.rowLines) - 1
		if maxLines < 1 {
			continue
		}
		lines[t.ID] = wrapTextWithWidths(t.Description, firstWidth, otherWidth, maxLines)
	}
	return lines
}

func (m *Model) refreshGridCache() {
	maxSlots := m.maxSlots()
	if maxSlots <= 0 {
		for day := range m.gridCache {
			m.gridCache[day] = nil
		}
		return
	}

	dayStart := m.dayStartMinutes()
	for day := 0; day < 7; day++ {
		dayCache := m.gridCache[day]
		if cap(dayCache) < maxSlots {
			dayCache = make([]*task.Task, maxSlots)
		} else {
			dayCache = dayCache[:maxSlots]
			for i := range dayCache {
				dayCache[i] = nil
			}
		}

		for slot := 0; slot < maxSlots; slot++ {
			timeLabel := minutesToTime(dayStart + (slot * m.rowHeight))
			var t *task.Task
			if m.mode == ModeMove {
				t = m.taskAtPreview(day, timeLabel)
			} else {
				t = m.taskAt(day, timeLabel)
			}
			dayCache[slot] = t
		}
		m.gridCache[day] = dayCache
	}
}

func (m *Model) refreshRenderCache() {
	rc := RenderCache{}
	rc.VerticalSep = m.styles.SeparatorStyle.Render("│")
	rc.EmptyCell = m.styleCache.EmptyCell.Render(" ")
	gap := m.styles.SeparatorStyle.Render(" ")
	rc.TimeBlankPrefix = m.styles.TimeColumnStyle.Render("      ") + gap + rc.VerticalSep

	maxSlots := m.maxSlots()
	if maxSlots > 0 {
		rc.TimeLabelPrefix = make([]string, maxSlots)
		dayStart := m.dayStartMinutes()
		for slot := 0; slot < maxSlots; slot++ {
			timeLabel := minutesToTime(dayStart + (slot * m.rowHeight))
			rc.TimeLabelPrefix[slot] = m.styles.TimeColumnStyle.Render(padRight(timeLabel, 6)) + gap + rc.VerticalSep
		}
	}

	extra := m.extraDayPadding()
	if extra > 0 {
		if extra == 1 {
			rc.ExtraLinePadding = rc.VerticalSep
		} else {
			extraPadding := lipgloss.NewStyle().
				Background(m.styles.colorBg).
				Render(strings.Repeat(" ", extra-1))
			rc.ExtraLinePadding = rc.VerticalSep + extraPadding
		}
	}

	rc.HorizontalSep = m.buildHorizontalSeparator()
	m.renderCache = rc
}

func (m *Model) cachedCursorTask() *task.Task {
	if m.cursor.Day < 0 || m.cursor.Day > 6 {
		return nil
	}
	dayTasks := m.gridCache[m.cursor.Day]
	if m.cursor.Slot < 0 || m.cursor.Slot >= len(dayTasks) {
		return nil
	}
	return dayTasks[m.cursor.Slot]
}

func (m *Model) buildHorizontalSeparator() string {
	var line strings.Builder
	line.WriteString(m.styles.SeparatorStyle.Render("───────┼"))
	extra := m.extraDayPadding()

	for i := 0; i < 7; i++ {
		line.WriteString(m.styles.SeparatorStyle.Render(strings.Repeat("─", m.colWidth)))
		if i < 6 {
			line.WriteString(m.styles.SeparatorStyle.Render("┼"))
		}
	}

	if extra > 0 {
		line.WriteString(m.styles.SeparatorStyle.Render("┼"))
		if extra > 1 {
			line.WriteString(m.styles.SeparatorStyle.Render(strings.Repeat("─", extra-1)))
		}
	}

	return line.String()
}

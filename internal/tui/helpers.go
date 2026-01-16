package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/javiermolinar/sancho/internal/task"
)

// Layout constants for boxed rendering.
const (
	footerCompact = 2

	footerBaseLines       = 4 // Stats(1) + Legend(1) + Status(1) + Help(1)
	promptBorderLines     = 2
	promptMinContentLines = 1

	footerMinHeight     = footerBaseLines + promptBorderLines + promptMinContentLines
	footerFullMinHeight = 15
)

// getFooterHeight returns the total lines consumed by the footer.
// This logic must match View()'s allocation.
func (m *Model) getFooterHeight() int {
	appH, appV := m.styles.AppStyle.GetFrameSize()
	innerW := m.width - appH
	innerH := m.height - appV
	_ = appH // Width not needed here

	if innerH <= 0 {
		return 0
	}

	if innerH < footerFullMinHeight {
		return footerCompact
	}

	if innerW < 0 {
		innerW = 0
	}

	promptWidth := promptContentWidth(m.styles, innerW)
	return m.fullFooterHeight(innerH, promptWidth)
}

// calculateLayout determines row height (minutes) and row lines based on terminal height.
// The grid always uses 15-minute slots; only row line height adapts to available space.
func (m *Model) calculateLayout() {
	if m.height == 0 {
		m.rowHeight = 15
		m.rowLines = 1
		return
	}

	appH, appV := m.styles.AppStyle.GetFrameSize()
	_ = appH
	innerH := m.height - appV
	footer := m.getFooterHeight()
	availableLines := innerH - footer

	if availableLines < 4 {
		availableLines = 4
	}

	dayStart := m.dayStartMinutes()
	dayEnd := m.dayEndMinutes()
	totalMinutes := dayEnd - dayStart

	m.rowHeight = 15

	slots := totalMinutes / m.rowHeight
	if slots <= 0 {
		slots = 1
	}

	linesPerRow := availableLines / slots
	if availableLines%slots != 0 {
		linesPerRow++
	}
	if linesPerRow < 1 {
		linesPerRow = 1
	}

	m.rowLines = linesPerRow

	LogChromeBreakdown(map[string]int{
		"terminal":  m.height,
		"inner":     innerH,
		"header":    0,
		"footer":    footer,
		"available": availableLines,
		"utilized":  slots * m.rowLines,
		"rowHeight": m.rowHeight,
		"rowLines":  m.rowLines,
	})
}

const layoutChrome = 22

// calculateColWidth determines the column width based on terminal width.
func (m *Model) calculateColWidth() int {
	if m.width == 0 {
		return defaultColWidth
	}

	// Layout chrome that reduces available width:
	// - AppStyle padding: 2 left + 2 right = 4
	// - TableStyle: border (2) + padding (2) = 4
	// - Time column: 6 chars + 1 space + separator (1) = 8
	// - Column separators: 6 separators between 7 days = 6
	// Total chrome: 4 + 4 + 8 + 6 = 22
	available := m.width - layoutChrome

	// Divide by 7 days
	colWidth := available / 7

	// Clamp to a minimum for readability.
	if colWidth < 10 {
		return 10
	}

	return colWidth
}

func (m *Model) extraDayPadding() int {
	return 0
}

// maxSlots returns the number of time slots in the grid.
func (m *Model) maxSlots() int {
	dayStart := m.dayStartMinutes()
	dayEnd := m.dayEndMinutes()
	if m.rowHeight <= 0 {
		return 0
	}
	return (dayEnd - dayStart) / m.rowHeight
}

func (m *Model) taskSlotSpan(t *task.Task) int {
	if t == nil {
		return 0
	}
	if m.rowHeight <= 0 {
		return 1
	}
	start := task.TimeToMinutes(t.ScheduledStart)
	end := task.TimeToMinutes(t.ScheduledEnd)
	duration := end - start
	if duration <= 0 {
		return 1
	}
	slots := duration / m.rowHeight
	if duration%m.rowHeight != 0 {
		slots++
	}
	if slots < 1 {
		return 1
	}
	return slots
}

// visibleRows returns the number of time SLOTS that fit in the terminal.
func (m *Model) visibleRows() int {
	visible := m.visibleSlotsForTable(m.layoutCache.GridH)
	if visible < 1 {
		visible = 1
	}

	totalSlots := m.maxSlots()
	if visible > totalSlots {
		return totalSlots
	}

	return visible
}

// ensureCursorVisible adjusts scroll offset to keep cursor visible.
func (m *Model) ensureCursorVisible() {
	visible := m.visibleRows()

	// If cursor is above visible area, scroll up
	if m.cursor.Slot < m.scrollOffset {
		m.scrollOffset = m.cursor.Slot
	}

	// If cursor is below visible area, scroll down
	if m.cursor.Slot >= m.scrollOffset+visible {
		m.scrollOffset = m.cursor.Slot - visible + 1
	}

	// Clamp scroll offset to valid range
	maxScroll := m.maxSlots() - visible
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.scrollOffset > maxScroll {
		m.scrollOffset = maxScroll
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

// dayStartMinutes returns the day start time in minutes.
func (m *Model) dayStartMinutes() int {
	start, _ := m.displayRangeMinutes()
	return start
}

// dayEndMinutes returns the day end time in minutes.
func (m *Model) dayEndMinutes() int {
	_, end := m.displayRangeMinutes()
	return end
}

func (m *Model) displayRangeMinutes() (int, int) {
	start := task.TimeToMinutes(m.config.Schedule.DayStart)
	end := task.TimeToMinutes(m.config.Schedule.DayEnd)

	ww := m.slotState.WeekWindow()
	if ww == nil || ww.Current() == nil {
		return start, end
	}

	for day := 0; day < 7; day++ {
		d := ww.Current().Day(day)
		if d == nil {
			continue
		}
		for _, t := range d.ScheduledTasks() {
			taskStart := task.TimeToMinutes(t.ScheduledStart)
			taskEnd := task.TimeToMinutes(t.ScheduledEnd)
			if taskStart < start {
				start = taskStart
			}
			if taskEnd > end {
				end = taskEnd
			}
		}
	}

	if start < 0 {
		start = 0
	}
	if end > MinutesPerDay {
		end = MinutesPerDay
	}
	if end-start < m.rowHeight {
		end = start + m.rowHeight
		if end > MinutesPerDay {
			end = MinutesPerDay
		}
	}

	return start, end
}

// slotToTime converts a slot index to a time string.
func (m *Model) slotToTime(slot int) string {
	mins := m.dayStartMinutes() + (slot * m.rowHeight)
	return minutesToTime(mins)
}

// taskAtCursor returns the task at the current cursor position.
func (m *Model) taskAtCursor() *task.Task {
	timeLabel := m.slotToTime(m.cursor.Slot)
	return m.taskAt(m.cursor.Day, timeLabel)
}

// taskAt returns the task at a specific day and time.
// The time represents the start of a display slot. We treat a task as occupying a slot
// if the slot index falls within the task's display coverage. If multiple tasks overlap
// the slot, we prefer the task that starts within the slot range. Otherwise, we show the
// task that started most recently before the slot and continues into it.
func (m *Model) taskAt(day int, timeLabel string) *task.Task {
	ww := m.slotState.WeekWindow()
	if ww == nil || ww.Current() == nil {
		return nil
	}

	d := ww.Current().Day(day)
	if d == nil {
		return nil
	}

	dayStart := m.dayStartMinutes()
	slotStart := task.TimeToMinutes(timeLabel)
	slotEnd := slotStart + m.rowHeight
	slotIndex := (slotStart - dayStart) / m.rowHeight
	if slotIndex < 0 {
		slotIndex = 0
	}

	var inSlot *task.Task
	inSlotStart := 0
	var bestLong *task.Task
	bestOverlap := -1
	minOverlap := m.rowHeight / 2

	for _, t := range d.ScheduledTasks() {
		taskStart := task.TimeToMinutes(t.ScheduledStart)
		taskEnd := task.TimeToMinutes(t.ScheduledEnd)
		taskDuration := taskEnd - taskStart
		// Check if task overlaps the display slot's time range
		// Overlap: task starts before slot ends AND task ends after slot starts
		if taskStart < slotEnd && taskEnd > slotStart {
			startSlot := (taskStart - dayStart) / m.rowHeight
			if taskStart < dayStart {
				startSlot = 0
			}
			endSlot := (taskEnd - dayStart) / m.rowHeight
			if endSlot <= startSlot {
				endSlot = startSlot + 1
			}
			if slotIndex < startSlot || slotIndex >= endSlot {
				continue
			}
			overlap := min(taskEnd, slotEnd) - max(taskStart, slotStart)
			if taskDuration <= m.rowHeight {
				if taskStart >= slotStart && taskStart < slotEnd {
					if inSlot == nil || taskStart < inSlotStart {
						inSlot = t
						inSlotStart = taskStart
					}
				}
				continue
			}
			if overlap > minOverlap {
				if bestLong == nil || overlap > bestOverlap {
					bestLong = t
					bestOverlap = overlap
				}
			}
		}
	}

	if inSlot != nil {
		// Log only when in move mode and this is the cursor day
		if m.mode == ModeMove && day == m.cursor.Day {
			LogTaskLookup(day, timeLabel, true, inSlot.ID, inSlot.Description)
		}
		return inSlot
	}
	if bestLong != nil {
		if m.mode == ModeMove && day == m.cursor.Day {
			LogTaskLookup(day, timeLabel, true, bestLong.ID, bestLong.Description)
		}
		return bestLong
	}

	// Log failed lookups in move mode for cursor day
	if m.mode == ModeMove && day == m.cursor.Day {
		LogTaskLookup(day, timeLabel, false, 0, "")
	}

	return nil
}

func (m *Model) now() time.Time {
	return m.nowFunc()()
}

func (m *Model) nowFunc() func() time.Time {
	if m.slotState != nil {
		cfg := m.slotState.Config()
		if cfg.Now != nil {
			return cfg.Now
		}
	}
	return time.Now
}

func (m *Model) currentWeekDayIndex(now time.Time) int {
	if m.slotState != nil {
		cfg := m.slotState.Config()
		gridDay := cfg.DateToDayIndex(now)
		if gridDay >= 0 {
			weekIndex, dayOfWeek := DayIndexToWeekAndDay(gridDay)
			if weekIndex == 1 {
				return dayOfWeek
			}
		}
	}
	return weekdayIndex(now)
}

func (m *Model) timeToDisplaySlot(now time.Time) int {
	if m.rowHeight <= 0 {
		return 0
	}

	mins := now.Hour()*60 + now.Minute()
	dayStart := m.dayStartMinutes()
	dayEnd := m.dayEndMinutes()
	if mins < dayStart {
		mins = dayStart
	}
	if mins >= dayEnd {
		mins = dayEnd - 1
	}
	if mins < dayStart {
		mins = dayStart
	}

	displaySlot := (mins - dayStart) / m.rowHeight
	maxSlot := m.maxSlots() - 1
	if maxSlot < 0 {
		return 0
	}
	if displaySlot > maxSlot {
		return maxSlot
	}
	if displaySlot < 0 {
		return 0
	}
	return displaySlot
}

func (m *Model) focusCursorOnCurrentTaskOrTime() {
	now := m.now()
	dayIndex := m.currentWeekDayIndex(now)
	if dayIndex < 0 || dayIndex > 6 {
		dayIndex = 0
	}

	if ww := m.slotState.WeekWindow(); ww != nil && ww.Current() != nil {
		if day := ww.Current().Day(dayIndex); day != nil {
			for _, t := range day.ScheduledTasks() {
				if m.isCurrentTask(t) {
					m.cursor.Day = dayIndex
					m.cursor.Slot = m.timeToDisplaySlot(now)
					m.ensureCursorVisible()
					return
				}
			}
		}
	}

	m.cursor.Day = dayIndex
	m.cursor.Slot = m.timeToDisplaySlot(now)
	m.ensureCursorVisible()
}

// isWorkday returns true if the given date is a workday.
func (m *Model) isWorkday(date time.Time) bool {
	weekdayName := date.Weekday().String()
	return m.config.IsWorkday(weekdayName)
}

// isCurrentTask returns true if the given task is happening right now.
// A task is "current" if today matches its scheduled date and the current time
// falls within its scheduled start and end times.
func (m *Model) isCurrentTask(t *task.Task) bool {
	if t == nil {
		return false
	}

	now := m.now()

	// Check if task is scheduled for today
	if !sameDay(t.ScheduledDate, now) {
		return false
	}

	// Get current time in minutes
	currentMins := now.Hour()*60 + now.Minute()
	startMins := task.TimeToMinutes(t.ScheduledStart)
	endMins := task.TimeToMinutes(t.ScheduledEnd)

	// Check if current time is within task's time range
	return currentMins >= startMins && currentMins < endMins
}

// Utility functions

// startOfWeek returns the Monday of the week containing the given date.
func startOfWeek(t time.Time) time.Time {
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return t.AddDate(0, 0, -(weekday - 1))
}

// weekdayIndex returns the weekday index (0=Monday, 6=Sunday).
func weekdayIndex(t time.Time) int {
	weekday := int(t.Weekday())
	if weekday == 0 {
		return 6 // Sunday
	}
	return weekday - 1
}

// sameDay returns true if two dates are the same day.
func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

// minutesToTime converts minutes since midnight to HH:MM format.
func minutesToTime(mins int) string {
	h := mins / 60
	m := mins % 60
	return fmt.Sprintf("%02d:%02d", h, m)
}

// addMinutesToTime adds minutes to a time string.
func addMinutesToTime(timeStr string, minutes int) string {
	mins := task.TimeToMinutes(timeStr) + minutes
	return minutesToTime(mins)
}

// truncate truncates a string to the given length.
func wrapTextWithWidths(s string, firstWidth, otherWidth, maxLines int) []string {
	if firstWidth <= 0 || otherWidth <= 0 || maxLines <= 0 {
		return nil
	}

	words := strings.Fields(s)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	var line string

	for _, word := range words {
		width := otherWidth
		if len(lines) == 0 {
			width = firstWidth
		}
		if line == "" {
			if len(word) > width {
				line = word[:width]
			} else {
				line = word
			}
			continue
		}

		if len(line)+1+len(word) <= width {
			line += " " + word
		} else {
			lines = append(lines, line)
			line = word
		}
	}
	if line != "" {
		lines = append(lines, line)
	}

	if len(lines) <= maxLines {
		return lines
	}

	lines = lines[:maxLines]
	last := lines[maxLines-1]
	width := otherWidth
	if maxLines == 1 {
		width = firstWidth
	}
	if width == 1 {
		lines[maxLines-1] = "…"
		return lines
	}
	if len(last) >= width {
		lines[maxLines-1] = last[:width-1] + "…"
		return lines
	}
	lines[maxLines-1] = last + "…"
	return lines
}

func truncateWithEllipsis(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if len(s) <= width {
		return s
	}
	if width == 1 {
		return "…"
	}
	return s[:width-1] + "…"
}

func (m Model) singleLineTaskContent(indicator string, t *task.Task) string {
	contentWidth := max(0, m.colWidth-1)
	if contentWidth == 0 {
		return ""
	}

	prefix := "[" + indicator + "] "
	if contentWidth <= len(prefix) {
		return prefix[:contentWidth]
	}

	available := contentWidth - len(prefix)
	timeRange := t.ScheduledStart + "-" + t.ScheduledEnd
	if available > len(timeRange)+1 {
		descWidth := available - len(timeRange) - 1
		desc := truncateWithEllipsis(t.Description, descWidth)
		gap := descWidth - len(desc)
		if gap < 1 {
			gap = 1
		}
		return prefix + desc + strings.Repeat(" ", gap) + timeRange
	}

	desc := truncateWithEllipsis(t.Description, available)
	return prefix + desc
}

// nextSlotDown returns the next slot when moving down.
// If on a task, jumps to the first slot after that task ends.
// If on empty space, moves one slot down.
func (m *Model) nextSlotDown() int {
	maxSlot := m.maxSlots() - 1
	currentTask := m.taskAtCursor()

	if currentTask == nil {
		// Empty slot - move one slot down
		return min(maxSlot, m.cursor.Slot+1)
	}

	// On a task - find the first slot that starts AT OR AFTER the task ends
	endMins := task.TimeToMinutes(currentTask.ScheduledEnd)
	dayStart := m.dayStartMinutes()

	// Calculate which slot contains the end time
	// We need the first slot where slotStart >= endMins
	// slotStart = dayStart + slot * rowHeight
	// We want: dayStart + slot * rowHeight >= endMins
	// So: slot >= (endMins - dayStart) / rowHeight
	// Use ceiling division to get the first slot at or after the end
	offsetMins := endMins - dayStart
	endSlot := offsetMins / m.rowHeight
	// If there's a remainder, we need the next slot
	if offsetMins%m.rowHeight != 0 {
		endSlot++
	}

	// If the end slot is at or beyond max, we're at the bottom
	if endSlot > maxSlot {
		return maxSlot
	}

	return endSlot
}

// isTaskShifted returns true if the given task has been shifted from its original position.
// This is used during move mode to style shifted tasks differently.
func (m *Model) isTaskShifted(t *task.Task) bool {
	if m.mode != ModeMove || !m.slotState.IsMoving() || t == nil {
		return false
	}

	moveState := m.slotState.MoveState()
	if moveState == nil {
		return false
	}

	for _, shiftedTask := range moveState.ShiftedTasks {
		if shiftedTask.ID == t.ID {
			return true
		}
	}
	return false
}

// taskAtPreview returns the task at a specific position using the preview state.
// In move mode, the SlotStateManager's workingGrid already contains the correct preview positions
// (computed fresh from beforeMoveGrid on each MoveUp/MoveDown/MoveRight call).
func (m *Model) taskAtPreview(day int, timeLabel string) *task.Task {
	// workingGrid always has the correct state - no special handling needed
	return m.taskAt(day, timeLabel)
}

// movingTaskSlotCount returns the number of slots the moving task occupies.
// Returns 0 if not in move mode or task not found.
func (m *Model) movingTaskSlotCount() int {
	if !m.slotState.IsMoving() {
		return 0
	}
	movingTask := m.slotState.MovingTask()
	if movingTask == nil {
		return 0
	}
	// Duration in minutes, divided by slot size (rowHeight)
	duration := movingTask.Duration()
	slots := duration / m.rowHeight
	if slots < 1 {
		slots = 1
	}
	return slots
}

// nextSlotUp returns the next slot when moving up.
// If on a task, jumps to the slot just before that task starts.
// If on empty space, moves one slot up.
func (m *Model) nextSlotUp() int {
	currentTask := m.taskAtCursor()

	if currentTask == nil {
		// Empty slot - move one slot up
		return max(0, m.cursor.Slot-1)
	}

	// On a task - find the slot just before the task starts
	startMins := task.TimeToMinutes(currentTask.ScheduledStart)
	dayStart := m.dayStartMinutes()

	// Calculate which slot contains the start time
	// The task starts in slot: (startMins - dayStart) / rowHeight
	// We want the slot BEFORE that
	startSlot := (startMins - dayStart) / m.rowHeight

	// Move to the slot before the task starts
	// If task starts exactly on a slot boundary, go to previous slot
	// If task starts mid-slot, the slot before is still startSlot - 1
	return max(0, startSlot-1)
}

// updateCursorToMovingTask updates the cursor position to follow the moving task.
// This is called after direction-based moves (MoveUp/MoveDown/MoveRight) to keep
// the cursor synchronized with the task's new position in the SlotGrid.
func (m *Model) updateCursorToMovingTask() {
	if !m.slotState.IsMoving() {
		return
	}

	moveState := m.slotState.MoveState()
	if moveState == nil {
		return
	}

	// Get the task's new position from the slot state
	targetDay := moveState.TargetDay
	targetSlot := moveState.TargetSlot

	// Convert slot grid day index (0-20 for 3 weeks) to current week day (0-6)
	// The current week is days 7-13 in the grid
	_, dayOfWeek := DayIndexToWeekAndDay(targetDay)

	// Update cursor position
	m.cursor.Day = dayOfWeek
	// Convert 15-min slots to display slots based on rowHeight
	// SlotGrid uses 15-min slots, but display might use 30 or 60 min
	m.cursor.Slot = m.slotToDisplaySlot(targetSlot)
	m.ensureCursorVisible()
}

func (m *Model) focusCursorOnTaskEnd(t *task.Task) {
	if t == nil || m.slotState == nil {
		return
	}

	grid := m.slotState.Grid()
	if grid == nil {
		return
	}

	day, _, endSlot, found := grid.FindTask(t)
	if !found || endSlot <= 0 {
		return
	}

	lastSlot := endSlot - 1
	_, dayOfWeek := DayIndexToWeekAndDay(day)
	m.cursor.Day = dayOfWeek
	m.cursor.Slot = m.slotToDisplaySlot(lastSlot)
	m.ensureCursorVisible()
}

// slotToDisplaySlot converts a 15-min slot index to display slot index.
// SlotGrid always uses 15-min slots internally, but the TUI display may use
// different slot sizes (15, 30, or 60 minutes) based on terminal height.
func (m *Model) slotToDisplaySlot(slot15min int) int {
	// Convert slot to minutes (15 min per slot)
	mins := slot15min * 15

	// Adjust for day start offset
	dayStart := m.dayStartMinutes()
	offsetMins := mins - dayStart

	if offsetMins < 0 {
		return 0
	}

	// Convert to display slot
	displaySlot := offsetMins / m.rowHeight
	maxSlot := m.maxSlots() - 1
	if displaySlot > maxSlot {
		return maxSlot
	}
	return displaySlot
}

// displaySlotToSlot converts a display slot index to a 15-min slot index.
func (m *Model) displaySlotToSlot(slot int) int {
	mins := m.dayStartMinutes() + (slot * m.rowHeight)
	return mins / 15
}

package tui

import (
	"time"

	"github.com/javiermolinar/sancho/internal/task"
)

// TasksToSlotGrid converts a slice of tasks to a SlotGrid.
// Tasks are placed at their scheduled time positions.
func TasksToSlotGrid(tasks []*task.Task, cfg SlotConfig) *SlotGrid {
	grid := NewSlotGrid(cfg)

	for _, t := range tasks {
		if t == nil {
			continue
		}

		// Only include scheduled tasks (not cancelled/postponed)
		if t.Status != task.StatusScheduled {
			continue
		}

		// Find which day this task belongs to
		dayIndex := cfg.DateToDayIndex(t.ScheduledDate)
		if dayIndex < 0 || dayIndex >= cfg.NumDays {
			continue // Task is outside the grid's date range
		}

		// Convert time to slot
		startSlot := cfg.MinutesToSlot(task.TimeToMinutes(t.ScheduledStart))
		endSlot := cfg.MinutesToSlot(task.TimeToMinutes(t.ScheduledEnd))

		if endSlot <= startSlot {
			continue
		}

		// Place task in grid (directly modify slots, bypassing Place() validation)
		// This is safe during initial load
		for s := startSlot; s < endSlot && s < SlotsPerDay; s++ {
			idx := grid.slotIndex(dayIndex, s)
			if idx >= 0 && idx < len(grid.slots) {
				grid.slots[idx] = t
			}
		}
	}

	return grid
}

// WeekWindowToSlotGrid converts a WeekWindow to a SlotGrid.
// The SlotGrid will contain 21 days (3 weeks) starting from prev week's Monday.
func WeekWindowToSlotGrid(ww *task.WeekWindow, cfg SlotConfig) *SlotGrid {
	if ww == nil {
		return NewSlotGrid(cfg)
	}

	var allTasks []*task.Task

	// Collect tasks from all weeks
	if ww.Previous() != nil {
		allTasks = append(allTasks, ww.Previous().AllTasks()...)
	}
	if ww.Current() != nil {
		allTasks = append(allTasks, ww.Current().AllTasks()...)
	}
	if ww.Next() != nil {
		allTasks = append(allTasks, ww.Next().AllTasks()...)
	}

	return TasksToSlotGrid(allTasks, cfg)
}

// SlotGridChanges represents changes to be persisted.
type SlotGridChanges struct {
	// UpdatedTasks contains tasks with modified times
	UpdatedTasks []*task.Task
}

// GetChangedTasks compares two grids and returns tasks that have changed position.
// Returns tasks from 'after' grid with their new times.
func GetChangedTasks(before, after *SlotGrid) SlotGridChanges {
	changes := SlotGridChanges{}

	if before == nil || after == nil {
		return changes
	}

	// Track which tasks we've processed
	processed := make(map[int64]bool)

	// Find all tasks in the after grid
	for _, t := range after.AllTasks() {
		if t == nil || processed[t.ID] {
			continue
		}
		processed[t.ID] = true

		// Find position in both grids
		beforeDay, beforeSlot, beforeEnd, foundBefore := before.FindTask(t)
		afterDay, afterSlot, afterEnd, foundAfter := after.FindTask(t)

		if !foundAfter {
			continue // Task was deleted
		}

		// Check if position changed
		positionChanged := !foundBefore ||
			beforeDay != afterDay ||
			beforeSlot != afterSlot ||
			beforeEnd != afterEnd

		if positionChanged {
			// Create updated task with new times
			updatedTask := *t // Copy
			updatedTask.ScheduledDate = after.config.DayIndexToDate(afterDay)
			updatedTask.ScheduledStart = after.config.SlotToTime(afterSlot)
			updatedTask.ScheduledEnd = after.config.SlotToTime(afterEnd)
			changes.UpdatedTasks = append(changes.UpdatedTasks, &updatedTask)
		}
	}

	return changes
}

// SlotGridConfigFromWeekWindow creates a SlotConfig based on a WeekWindow.
// The grid will start from the previous week's Monday and span 3 weeks.
// rowHeight is the display row size in minutes (currently fixed at 15) - used for visual block movement.
func SlotGridConfigFromWeekWindow(ww *task.WeekWindow, workStart, workEnd string, now func() time.Time, rowHeight int) SlotConfig {
	var firstDate time.Time

	switch {
	case ww != nil && ww.Previous() != nil:
		firstDate = ww.Previous().StartDate
	case ww != nil && ww.Current() != nil:
		// Calculate previous Monday
		firstDate = ww.Current().StartDate.AddDate(0, 0, -7)
	default:
		// Default to 1 week before today's Monday
		today := time.Now()
		weekday := int(today.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday
		}
		monday := today.AddDate(0, 0, -(weekday - 1))
		firstDate = monday.AddDate(0, 0, -7)
	}

	// Truncate to start of day
	firstDate = time.Date(firstDate.Year(), firstDate.Month(), firstDate.Day(),
		0, 0, 0, 0, firstDate.Location())

	// Calculate DisplaySlotSize from rowHeight
	// DisplaySlotSize = rowHeight / 15 (since internal slots are 15 min each)
	displaySlotSize := rowHeight / DefaultSlotDuration
	if displaySlotSize < 1 {
		displaySlotSize = 1
	}

	cfg := SlotConfig{
		SlotDuration:      DefaultSlotDuration,
		NumDays:           DefaultNumWeeks * DaysPerWeek, // 21 days
		FirstDate:         firstDate,
		Now:               now,
		WorkingHoursStart: task.TimeToMinutes(workStart),
		WorkingHoursEnd:   task.TimeToMinutes(workEnd),
		DisplaySlotSize:   displaySlotSize,
	}

	if cfg.Now == nil {
		cfg.Now = time.Now
	}

	return cfg
}

// DayIndexToWeekAndDay converts a grid day index to week index (0-2) and day within week (0-6).
func DayIndexToWeekAndDay(dayIndex int) (weekIndex, dayOfWeek int) {
	weekIndex = dayIndex / DaysPerWeek
	dayOfWeek = dayIndex % DaysPerWeek
	return weekIndex, dayOfWeek
}

// WeekAndDayToDayIndex converts week index and day within week to a grid day index.
func WeekAndDayToDayIndex(weekIndex, dayOfWeek int) int {
	return weekIndex*DaysPerWeek + dayOfWeek
}

// SlotGridToWeekWindow converts a SlotGrid to a WeekWindow.
// The SlotGrid must contain 21 days (3 weeks):
//   - Days 0-6:   Previous week
//   - Days 7-13:  Current week
//   - Days 14-20: Next week
func SlotGridToWeekWindow(grid *SlotGrid) *task.WeekWindow {
	if grid == nil {
		return nil
	}

	// Create three weeks from the grid
	prevWeek := slotGridToWeek(grid, 0)  // Days 0-6
	currWeek := slotGridToWeek(grid, 7)  // Days 7-13
	nextWeek := slotGridToWeek(grid, 14) // Days 14-20

	return task.NewWeekWindow(prevWeek, currWeek, nextWeek)
}

// slotGridToWeek extracts a single Week from the SlotGrid starting at the given day index.
// startDay should be 0, 7, or 14 for prev/current/next week respectively.
func slotGridToWeek(grid *SlotGrid, startDay int) *task.Week {
	if grid == nil {
		return nil
	}

	// Get the Monday date for this week
	weekStart := grid.config.DayIndexToDate(startDay)
	week := task.NewWeek(weekStart)

	// Process each day of the week (7 days)
	for dayOffset := 0; dayOffset < DaysPerWeek; dayOffset++ {
		dayIndex := startDay + dayOffset

		// Get all tasks for this day
		tasks := grid.TasksOnDay(dayIndex)

		// Add each task to the day
		// Note: Tasks in SlotGrid already have their times stored in the Task struct,
		// but we need to update them based on their current slot position
		for _, t := range tasks {
			if t == nil {
				continue
			}

			// Find the task's position in the grid to get accurate times
			foundDay, startSlot, endSlot, found := grid.FindTask(t)
			if !found || foundDay != dayIndex {
				continue
			}

			// Create a copy of the task with updated times from the grid
			taskCopy := *t
			taskCopy.ScheduledDate = grid.config.DayIndexToDate(dayIndex)
			taskCopy.ScheduledStart = grid.config.SlotToTime(startSlot)
			taskCopy.ScheduledEnd = grid.config.SlotToTime(endSlot)

			// Add to the day (ignore overlap errors since SlotGrid doesn't allow overlaps)
			_ = week.Day(dayOffset).AddTask(&taskCopy)
		}
	}

	return week
}

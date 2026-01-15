package tui

import (
	"testing"
	"time"

	"github.com/javiermolinar/sancho/internal/task"
)

// translateTestConfig creates a SlotConfig for translation tests with a fixed first date.
// Now is set to one day BEFORE the first date so no tasks are considered "past".
func translateTestConfig(firstDate time.Time) SlotConfig {
	return SlotConfig{
		SlotDuration:      DefaultSlotDuration,
		NumDays:           21, // 3 weeks
		FirstDate:         firstDate,
		Now:               func() time.Time { return firstDate.Add(-24 * time.Hour) }, // Day before grid starts
		WorkingHoursStart: 8 * 60,                                                     // 08:00
		WorkingHoursEnd:   18 * 60,                                                    // 18:00
	}
}

// makeScheduledTask creates a task with given parameters for testing.
func makeScheduledTask(id int64, date time.Time, start, end string) *task.Task {
	return &task.Task{
		ID:             id,
		Description:    "Test task",
		Category:       task.CategoryDeep,
		ScheduledDate:  date,
		ScheduledStart: start,
		ScheduledEnd:   end,
		Status:         task.StatusScheduled,
	}
}

func TestTasksToSlotGrid(t *testing.T) {
	firstDate := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC) // Monday
	cfg := translateTestConfig(firstDate)

	tests := []struct {
		name      string
		tasks     []*task.Task
		wantCount int // expected number of unique tasks in grid
	}{
		{
			name:      "empty tasks",
			tasks:     nil,
			wantCount: 0,
		},
		{
			name:      "nil task in slice",
			tasks:     []*task.Task{nil},
			wantCount: 0,
		},
		{
			name: "single task on day 0",
			tasks: []*task.Task{
				makeScheduledTask(1, firstDate, "09:00", "10:00"),
			},
			wantCount: 1,
		},
		{
			name: "task on day 7 (second week)",
			tasks: []*task.Task{
				makeScheduledTask(1, firstDate.AddDate(0, 0, 7), "09:00", "10:00"),
			},
			wantCount: 1,
		},
		{
			name: "multiple tasks same day",
			tasks: []*task.Task{
				makeScheduledTask(1, firstDate, "09:00", "10:00"),
				makeScheduledTask(2, firstDate, "10:00", "11:00"),
			},
			wantCount: 2,
		},
		{
			name: "tasks on different days",
			tasks: []*task.Task{
				makeScheduledTask(1, firstDate, "09:00", "10:00"),
				makeScheduledTask(2, firstDate.AddDate(0, 0, 1), "09:00", "10:00"),
				makeScheduledTask(3, firstDate.AddDate(0, 0, 2), "09:00", "10:00"),
			},
			wantCount: 3,
		},
		{
			name: "cancelled task is excluded",
			tasks: []*task.Task{
				makeScheduledTask(1, firstDate, "09:00", "10:00"),
				{
					ID:             2,
					Description:    "Cancelled",
					Category:       task.CategoryDeep,
					ScheduledDate:  firstDate,
					ScheduledStart: "11:00",
					ScheduledEnd:   "12:00",
					Status:         task.StatusCancelled,
				},
			},
			wantCount: 1,
		},
		{
			name: "task outside grid date range is excluded",
			tasks: []*task.Task{
				makeScheduledTask(1, firstDate.AddDate(0, 0, -1), "09:00", "10:00"), // Day before grid
				makeScheduledTask(2, firstDate, "09:00", "10:00"),                   // Day 0
				makeScheduledTask(3, firstDate.AddDate(0, 0, 25), "09:00", "10:00"), // Day 25 (beyond 21)
			},
			wantCount: 1, // Only task 2
		},
		{
			name: "task with invalid end time (end <= start) excluded",
			tasks: []*task.Task{
				{
					ID:             1,
					Description:    "Invalid",
					Category:       task.CategoryDeep,
					ScheduledDate:  firstDate,
					ScheduledStart: "10:00",
					ScheduledEnd:   "09:00", // End before start
					Status:         task.StatusScheduled,
				},
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid := TasksToSlotGrid(tt.tasks, cfg)
			if grid == nil {
				t.Fatal("expected non-nil grid")
			}

			allTasks := grid.AllTasks()
			if len(allTasks) != tt.wantCount {
				t.Errorf("got %d tasks, want %d", len(allTasks), tt.wantCount)
			}
		})
	}
}

func TestTasksToSlotGrid_SlotPlacement(t *testing.T) {
	firstDate := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC) // Monday
	cfg := translateTestConfig(firstDate)

	// Task from 09:00 to 10:00 should occupy slots 36-39 (4 slots)
	// 09:00 = 540 mins / 15 = 36
	// 10:00 = 600 mins / 15 = 40 (exclusive)
	task1 := makeScheduledTask(1, firstDate, "09:00", "10:00")

	grid := TasksToSlotGrid([]*task.Task{task1}, cfg)

	// Check that slots 36-39 are occupied on day 0
	for slot := 36; slot < 40; slot++ {
		got := grid.TaskAt(0, slot)
		if got == nil {
			t.Errorf("slot %d should have task, got nil", slot)
		} else if got.ID != 1 {
			t.Errorf("slot %d should have task 1, got task %d", slot, got.ID)
		}
	}

	// Check adjacent slots are empty
	if grid.TaskAt(0, 35) != nil {
		t.Error("slot 35 should be empty")
	}
	if grid.TaskAt(0, 40) != nil {
		t.Error("slot 40 should be empty")
	}
}

func TestWeekWindowToSlotGrid(t *testing.T) {
	// Monday, Jan 6, 2025
	currentMonday := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	prevMonday := currentMonday.AddDate(0, 0, -7)
	nextMonday := currentMonday.AddDate(0, 0, 7)

	// Create weeks with tasks
	prevWeek := task.NewWeek(prevMonday)
	currentWeek := task.NewWeek(currentMonday)
	nextWeek := task.NewWeek(nextMonday)

	// Add tasks to each week
	prevTask := makeScheduledTask(1, prevMonday, "09:00", "10:00")
	_ = prevWeek.Day(0).AddTask(prevTask)

	currentTask := makeScheduledTask(2, currentMonday, "10:00", "11:00")
	_ = currentWeek.Day(0).AddTask(currentTask)

	nextTask := makeScheduledTask(3, nextMonday, "11:00", "12:00")
	_ = nextWeek.Day(0).AddTask(nextTask)

	ww := task.NewWeekWindow(prevWeek, currentWeek, nextWeek)

	// Config starting from prev week's Monday
	cfg := SlotConfig{
		SlotDuration:      DefaultSlotDuration,
		NumDays:           21,
		FirstDate:         prevMonday,
		Now:               func() time.Time { return currentMonday },
		WorkingHoursStart: 8 * 60,
		WorkingHoursEnd:   18 * 60,
	}

	grid := WeekWindowToSlotGrid(ww, cfg)

	// Should have 3 tasks
	allTasks := grid.AllTasks()
	if len(allTasks) != 3 {
		t.Errorf("got %d tasks, want 3", len(allTasks))
	}

	// prevTask should be on day 0 (prevMonday)
	day0, slot0, _, found := grid.FindTask(prevTask)
	if !found {
		t.Error("prevTask not found in grid")
	} else {
		if day0 != 0 {
			t.Errorf("prevTask on day %d, want 0", day0)
		}
		if slot0 != 36 { // 09:00 = 540 mins / 15 = 36
			t.Errorf("prevTask at slot %d, want 36", slot0)
		}
	}

	// currentTask should be on day 7 (currentMonday)
	day1, slot1, _, found := grid.FindTask(currentTask)
	if !found {
		t.Error("currentTask not found in grid")
	} else {
		if day1 != 7 {
			t.Errorf("currentTask on day %d, want 7", day1)
		}
		if slot1 != 40 { // 10:00 = 600 mins / 15 = 40
			t.Errorf("currentTask at slot %d, want 40", slot1)
		}
	}

	// nextTask should be on day 14 (nextMonday)
	day2, slot2, _, found := grid.FindTask(nextTask)
	if !found {
		t.Error("nextTask not found in grid")
	} else {
		if day2 != 14 {
			t.Errorf("nextTask on day %d, want 14", day2)
		}
		if slot2 != 44 { // 11:00 = 660 mins / 15 = 44
			t.Errorf("nextTask at slot %d, want 44", slot2)
		}
	}
}

func TestWeekWindowToSlotGrid_NilWindow(t *testing.T) {
	cfg := translateTestConfig(time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC))

	grid := WeekWindowToSlotGrid(nil, cfg)
	if grid == nil {
		t.Fatal("expected non-nil grid for nil WeekWindow")
	}
	if len(grid.AllTasks()) != 0 {
		t.Error("expected empty grid for nil WeekWindow")
	}
}

func TestGetChangedTasks(t *testing.T) {
	firstDate := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	cfg := translateTestConfig(firstDate)

	// Create task at 09:00-10:00 on day 0
	task1 := makeScheduledTask(1, firstDate, "09:00", "10:00")

	beforeGrid := TasksToSlotGrid([]*task.Task{task1}, cfg)

	tests := []struct {
		name        string
		makeAfter   func() *SlotGrid
		wantUpdated int
	}{
		{
			name: "no changes",
			makeAfter: func() *SlotGrid {
				return TasksToSlotGrid([]*task.Task{task1}, cfg)
			},
			wantUpdated: 0,
		},
		{
			name: "task moved down (later time)",
			makeAfter: func() *SlotGrid {
				grid := TasksToSlotGrid([]*task.Task{task1}, cfg)
				// Simulate move: remove from 36-39, place at 40-43
				moved, _ := grid.MoveDown(task1)
				return moved
			},
			wantUpdated: 1,
		},
		{
			name: "task moved to different day",
			makeAfter: func() *SlotGrid {
				grid := TasksToSlotGrid([]*task.Task{task1}, cfg)
				moved, _ := grid.MoveRight(task1)
				return moved
			},
			wantUpdated: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			afterGrid := tt.makeAfter()
			changes := GetChangedTasks(beforeGrid, afterGrid)

			if len(changes.UpdatedTasks) != tt.wantUpdated {
				t.Errorf("got %d updated tasks, want %d", len(changes.UpdatedTasks), tt.wantUpdated)
			}
		})
	}
}

func TestGetChangedTasks_VerifiesNewTimes(t *testing.T) {
	firstDate := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	cfg := translateTestConfig(firstDate)

	task1 := makeScheduledTask(1, firstDate, "09:00", "10:00")
	beforeGrid := TasksToSlotGrid([]*task.Task{task1}, cfg)

	// Move task to day 1 via MoveRight
	afterGrid, err := beforeGrid.MoveRight(task1)
	if err != nil {
		t.Fatalf("MoveRight failed: %v", err)
	}

	changes := GetChangedTasks(beforeGrid, afterGrid)

	if len(changes.UpdatedTasks) != 1 {
		t.Fatalf("expected 1 updated task, got %d", len(changes.UpdatedTasks))
	}

	updated := changes.UpdatedTasks[0]

	// Should have new date (day 1 = firstDate + 1 day)
	expectedDate := firstDate.AddDate(0, 0, 1)
	if !updated.ScheduledDate.Equal(expectedDate) {
		t.Errorf("expected date %v, got %v", expectedDate, updated.ScheduledDate)
	}

	// Time should remain the same (09:00-10:00)
	if updated.ScheduledStart != "09:00" {
		t.Errorf("expected start 09:00, got %s", updated.ScheduledStart)
	}
	if updated.ScheduledEnd != "10:00" {
		t.Errorf("expected end 10:00, got %s", updated.ScheduledEnd)
	}
}

func TestGetChangedTasks_NilGrids(t *testing.T) {
	firstDate := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	cfg := translateTestConfig(firstDate)

	task1 := makeScheduledTask(1, firstDate, "09:00", "10:00")
	grid := TasksToSlotGrid([]*task.Task{task1}, cfg)

	// nil before
	changes := GetChangedTasks(nil, grid)
	if len(changes.UpdatedTasks) != 0 {
		t.Error("expected no changes with nil before grid")
	}

	// nil after
	changes = GetChangedTasks(grid, nil)
	if len(changes.UpdatedTasks) != 0 {
		t.Error("expected no changes with nil after grid")
	}

	// both nil
	changes = GetChangedTasks(nil, nil)
	if len(changes.UpdatedTasks) != 0 {
		t.Error("expected no changes with both grids nil")
	}
}

func TestSlotGridConfigFromWeekWindow(t *testing.T) {
	currentMonday := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	prevMonday := currentMonday.AddDate(0, 0, -7) // Dec 30, 2024

	prevWeek := task.NewWeek(prevMonday)
	currentWeek := task.NewWeek(currentMonday)
	nextWeek := task.NewWeek(currentMonday.AddDate(0, 0, 7))

	ww := task.NewWeekWindow(prevWeek, currentWeek, nextWeek)

	nowFunc := func() time.Time { return currentMonday }
	cfg := SlotGridConfigFromWeekWindow(ww, "08:00", "18:00", nowFunc, 60)

	// FirstDate should be prev week's Monday
	if !cfg.FirstDate.Equal(prevMonday) {
		t.Errorf("FirstDate = %v, want %v", cfg.FirstDate, prevMonday)
	}

	// NumDays should be 21
	if cfg.NumDays != 21 {
		t.Errorf("NumDays = %d, want 21", cfg.NumDays)
	}

	// SlotDuration should be 15
	if cfg.SlotDuration != 15 {
		t.Errorf("SlotDuration = %d, want 15", cfg.SlotDuration)
	}

	// Working hours
	if cfg.WorkingHoursStart != 8*60 {
		t.Errorf("WorkingHoursStart = %d, want %d", cfg.WorkingHoursStart, 8*60)
	}
	if cfg.WorkingHoursEnd != 18*60 {
		t.Errorf("WorkingHoursEnd = %d, want %d", cfg.WorkingHoursEnd, 18*60)
	}
}

func TestSlotGridConfigFromWeekWindow_NilPrevWeek(t *testing.T) {
	currentMonday := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	prevMonday := currentMonday.AddDate(0, 0, -7)

	currentWeek := task.NewWeek(currentMonday)
	nextWeek := task.NewWeek(currentMonday.AddDate(0, 0, 7))

	// nil prev week - should calculate prev Monday from current
	ww := task.NewWeekWindow(nil, currentWeek, nextWeek)

	nowFunc := func() time.Time { return currentMonday }
	cfg := SlotGridConfigFromWeekWindow(ww, "08:00", "18:00", nowFunc, 60)

	// FirstDate should still be prev Monday (calculated)
	if !cfg.FirstDate.Equal(prevMonday) {
		t.Errorf("FirstDate = %v, want %v", cfg.FirstDate, prevMonday)
	}
}

func TestSlotGridConfigFromWeekWindow_NilWindow(t *testing.T) {
	nowFunc := func() time.Time { return time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC) }
	cfg := SlotGridConfigFromWeekWindow(nil, "08:00", "18:00", nowFunc, 60)

	// Should not panic, return reasonable defaults
	if cfg.NumDays != 21 {
		t.Errorf("NumDays = %d, want 21", cfg.NumDays)
	}
	if cfg.SlotDuration != 15 {
		t.Errorf("SlotDuration = %d, want 15", cfg.SlotDuration)
	}
}

func TestDayIndexToWeekAndDay(t *testing.T) {
	tests := []struct {
		dayIndex   int
		wantWeek   int
		wantDayOff int
	}{
		{0, 0, 0},
		{1, 0, 1},
		{6, 0, 6},
		{7, 1, 0},
		{8, 1, 1},
		{13, 1, 6},
		{14, 2, 0},
		{20, 2, 6},
	}

	for _, tt := range tests {
		week, day := DayIndexToWeekAndDay(tt.dayIndex)
		if week != tt.wantWeek || day != tt.wantDayOff {
			t.Errorf("DayIndexToWeekAndDay(%d) = (%d, %d), want (%d, %d)",
				tt.dayIndex, week, day, tt.wantWeek, tt.wantDayOff)
		}
	}
}

func TestWeekAndDayToDayIndex(t *testing.T) {
	tests := []struct {
		weekIndex int
		dayOfWeek int
		want      int
	}{
		{0, 0, 0},
		{0, 1, 1},
		{0, 6, 6},
		{1, 0, 7},
		{1, 1, 8},
		{1, 6, 13},
		{2, 0, 14},
		{2, 6, 20},
	}

	for _, tt := range tests {
		got := WeekAndDayToDayIndex(tt.weekIndex, tt.dayOfWeek)
		if got != tt.want {
			t.Errorf("WeekAndDayToDayIndex(%d, %d) = %d, want %d",
				tt.weekIndex, tt.dayOfWeek, got, tt.want)
		}
	}
}

func TestDayIndexConversionsRoundTrip(t *testing.T) {
	for dayIndex := 0; dayIndex < 21; dayIndex++ {
		week, day := DayIndexToWeekAndDay(dayIndex)
		got := WeekAndDayToDayIndex(week, day)
		if got != dayIndex {
			t.Errorf("round trip failed for dayIndex %d: got %d", dayIndex, got)
		}
	}
}

func TestSlotGridToWeekWindow(t *testing.T) {
	// Monday, Dec 30, 2024 (prev week start)
	prevMonday := time.Date(2024, 12, 30, 0, 0, 0, 0, time.UTC)
	currentMonday := prevMonday.AddDate(0, 0, 7) // Jan 6, 2025
	nextMonday := currentMonday.AddDate(0, 0, 7) // Jan 13, 2025

	cfg := translateTestConfig(prevMonday)

	// Create tasks on each week
	prevTask := makeScheduledTask(1, prevMonday, "09:00", "10:00")                         // Day 0
	currentTask := makeScheduledTask(2, currentMonday, "10:00", "11:00")                   // Day 7
	currentTask2 := makeScheduledTask(3, currentMonday.AddDate(0, 0, 2), "14:00", "15:00") // Day 9 (Wed)
	nextTask := makeScheduledTask(4, nextMonday, "11:00", "12:00")                         // Day 14

	tasks := []*task.Task{prevTask, currentTask, currentTask2, nextTask}
	grid := TasksToSlotGrid(tasks, cfg)

	// Convert back to WeekWindow
	ww := SlotGridToWeekWindow(grid)

	if ww == nil {
		t.Fatal("expected non-nil WeekWindow")
	}

	// Verify previous week
	if ww.Previous() == nil {
		t.Fatal("Previous() is nil")
	}
	prevWeekTasks := ww.Previous().AllTasks()
	if len(prevWeekTasks) != 1 {
		t.Errorf("Previous week has %d tasks, want 1", len(prevWeekTasks))
	} else if prevWeekTasks[0].ID != 1 {
		t.Errorf("Previous week task ID = %d, want 1", prevWeekTasks[0].ID)
	}

	// Verify current week
	if ww.Current() == nil {
		t.Fatal("Current() is nil")
	}
	currentWeekTasks := ww.Current().AllTasks()
	if len(currentWeekTasks) != 2 {
		t.Errorf("Current week has %d tasks, want 2", len(currentWeekTasks))
	}

	// Verify next week
	if ww.Next() == nil {
		t.Fatal("Next() is nil")
	}
	nextWeekTasks := ww.Next().AllTasks()
	if len(nextWeekTasks) != 1 {
		t.Errorf("Next week has %d tasks, want 1", len(nextWeekTasks))
	} else if nextWeekTasks[0].ID != 4 {
		t.Errorf("Next week task ID = %d, want 4", nextWeekTasks[0].ID)
	}
}

func TestSlotGridToWeekWindow_NilGrid(t *testing.T) {
	ww := SlotGridToWeekWindow(nil)
	if ww != nil {
		t.Error("expected nil WeekWindow for nil grid")
	}
}

func TestSlotGridToWeekWindow_EmptyGrid(t *testing.T) {
	firstDate := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	cfg := translateTestConfig(firstDate.AddDate(0, 0, -7)) // Prev week start

	grid := NewSlotGrid(cfg)
	ww := SlotGridToWeekWindow(grid)

	if ww == nil {
		t.Fatal("expected non-nil WeekWindow")
	}

	// All weeks should exist but be empty
	if ww.Previous() == nil || len(ww.Previous().AllTasks()) != 0 {
		t.Error("Previous week should be empty")
	}
	if ww.Current() == nil || len(ww.Current().AllTasks()) != 0 {
		t.Error("Current week should be empty")
	}
	if ww.Next() == nil || len(ww.Next().AllTasks()) != 0 {
		t.Error("Next week should be empty")
	}
}

func TestSlotGridToWeekWindow_TaskTimesUpdated(t *testing.T) {
	// Verify that task times are updated from grid position, not original Task fields
	prevMonday := time.Date(2024, 12, 30, 0, 0, 0, 0, time.UTC)
	cfg := translateTestConfig(prevMonday)

	// Create task at 09:00
	originalTask := makeScheduledTask(1, prevMonday.AddDate(0, 0, 7), "09:00", "10:00") // Day 7

	grid := TasksToSlotGrid([]*task.Task{originalTask}, cfg)

	// Move the task down (simulating an edit)
	movedGrid, err := grid.MoveDown(originalTask)
	if err != nil {
		t.Fatalf("MoveDown failed: %v", err)
	}

	// Convert to WeekWindow
	ww := SlotGridToWeekWindow(movedGrid)

	// The task in WeekWindow should have UPDATED times
	currentTasks := ww.Current().AllTasks()
	if len(currentTasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(currentTasks))
	}

	movedTask := currentTasks[0]

	// Should NOT be 09:00 anymore - MoveDown moves by one slot (15 min) into a gap
	if movedTask.ScheduledStart == "09:00" {
		t.Errorf("Task start time should have changed from 09:00, still got 09:00")
	}
}

func TestSlotGridToWeekWindow_RoundTrip(t *testing.T) {
	// Test WeekWindow -> SlotGrid -> WeekWindow preserves tasks
	currentMonday := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	prevMonday := currentMonday.AddDate(0, 0, -7)
	nextMonday := currentMonday.AddDate(0, 0, 7)

	// Create original WeekWindow with tasks
	prevWeek := task.NewWeek(prevMonday)
	currentWeek := task.NewWeek(currentMonday)
	nextWeek := task.NewWeek(nextMonday)

	task1 := makeScheduledTask(1, prevMonday.AddDate(0, 0, 2), "09:00", "10:00")
	task2 := makeScheduledTask(2, currentMonday.AddDate(0, 0, 3), "10:00", "11:00")
	task3 := makeScheduledTask(3, nextMonday.AddDate(0, 0, 4), "11:00", "12:00")

	_ = prevWeek.Day(2).AddTask(task1)
	_ = currentWeek.Day(3).AddTask(task2)
	_ = nextWeek.Day(4).AddTask(task3)

	originalWW := task.NewWeekWindow(prevWeek, currentWeek, nextWeek)

	// Convert to SlotGrid
	cfg := SlotGridConfigFromWeekWindow(originalWW, "08:00", "18:00", time.Now, 60)
	grid := WeekWindowToSlotGrid(originalWW, cfg)

	// Convert back to WeekWindow
	resultWW := SlotGridToWeekWindow(grid)

	// Verify task counts match
	if len(resultWW.Previous().AllTasks()) != 1 {
		t.Errorf("Previous week: got %d tasks, want 1", len(resultWW.Previous().AllTasks()))
	}
	if len(resultWW.Current().AllTasks()) != 1 {
		t.Errorf("Current week: got %d tasks, want 1", len(resultWW.Current().AllTasks()))
	}
	if len(resultWW.Next().AllTasks()) != 1 {
		t.Errorf("Next week: got %d tasks, want 1", len(resultWW.Next().AllTasks()))
	}

	// Verify task times are preserved
	resultTask := resultWW.Previous().AllTasks()[0]
	if resultTask.ScheduledStart != "09:00" || resultTask.ScheduledEnd != "10:00" {
		t.Errorf("Task times not preserved: got %s-%s, want 09:00-10:00",
			resultTask.ScheduledStart, resultTask.ScheduledEnd)
	}
}

func TestSlotGridToWeekWindow_AfterMoveLeft(t *testing.T) {
	// Test that after moving a task left (to previous day), the WeekWindow reflects the new position
	prevMonday := time.Date(2024, 12, 30, 0, 0, 0, 0, time.UTC)
	currentMonday := prevMonday.AddDate(0, 0, 7)

	cfg := translateTestConfig(prevMonday)

	// Create tasks on Tuesday and Wednesday
	tuesday := currentMonday.AddDate(0, 0, 1)
	wednesday := currentMonday.AddDate(0, 0, 2)

	existingTask := &task.Task{
		ID: 1, Description: "Email check", Category: task.CategoryShallow,
		ScheduledDate: tuesday, ScheduledStart: "09:00", ScheduledEnd: "09:30",
		Status: task.StatusScheduled,
	}
	movingTask := &task.Task{
		ID: 2, Description: "Send email", Category: task.CategoryShallow,
		ScheduledDate: wednesday, ScheduledStart: "09:00", ScheduledEnd: "09:15",
		Status: task.StatusScheduled,
	}

	tasks := []*task.Task{existingTask, movingTask}
	grid := TasksToSlotGrid(tasks, cfg)

	// Verify initial state - movingTask is on grid day 9 (Wednesday)
	day, _, _, found := grid.FindTask(movingTask)
	if !found || day != 9 {
		t.Fatalf("Initial: movingTask should be on day 9 (Wednesday), got day %d, found=%v", day, found)
	}

	// Move task left (from Wednesday to Tuesday)
	newGrid, err := grid.MoveLeft(movingTask)
	if err != nil {
		t.Fatalf("MoveLeft failed: %v", err)
	}

	// Verify task moved to day 8 (Tuesday)
	newDay, newSlot, _, found := newGrid.FindTask(movingTask)
	if !found {
		t.Fatal("After MoveLeft: movingTask not found in grid")
	}
	if newDay != 8 {
		t.Errorf("After MoveLeft: movingTask should be on day 8 (Tuesday), got day %d", newDay)
	}
	t.Logf("After MoveLeft: movingTask is at day=%d, slot=%d", newDay, newSlot)

	// Convert to WeekWindow
	ww := SlotGridToWeekWindow(newGrid)
	if ww == nil {
		t.Fatal("WeekWindow is nil")
	}

	// Check Tuesday (day 1 in current week)
	tuesdayTasks := ww.Current().Day(1).ScheduledTasks()
	t.Logf("Tuesday has %d tasks:", len(tuesdayTasks))
	for _, tsk := range tuesdayTasks {
		t.Logf("  - ID=%d %s (%s-%s)", tsk.ID, tsk.Description, tsk.ScheduledStart, tsk.ScheduledEnd)
	}

	if len(tuesdayTasks) != 2 {
		t.Errorf("Tuesday should have 2 tasks (existing + moved), got %d", len(tuesdayTasks))
	}

	// Verify movingTask is on Tuesday with updated times
	var foundMovingTask *task.Task
	for _, tsk := range tuesdayTasks {
		if tsk.ID == 2 {
			foundMovingTask = tsk
			break
		}
	}
	if foundMovingTask == nil {
		t.Error("movingTask (ID=2) not found on Tuesday after MoveLeft")
	} else {
		t.Logf("Found movingTask on Tuesday: %s-%s", foundMovingTask.ScheduledStart, foundMovingTask.ScheduledEnd)
	}

	// Check Wednesday (day 2 in current week) - should be empty now
	wednesdayTasks := ww.Current().Day(2).ScheduledTasks()
	if len(wednesdayTasks) != 0 {
		t.Errorf("Wednesday should be empty after MoveLeft, got %d tasks", len(wednesdayTasks))
	}
}

func TestSlotGridToWeekWindow_MultipleAdjacentTasks(t *testing.T) {
	// Test multiple adjacent tasks on the same day to verify they're all converted correctly
	// This is the scenario that was causing rendering issues
	prevMonday := time.Date(2024, 12, 30, 0, 0, 0, 0, time.UTC)
	currentMonday := prevMonday.AddDate(0, 0, 7) // Jan 6, 2025

	cfg := translateTestConfig(prevMonday)

	// Create 4 adjacent tasks on Tuesday (day 8 in grid, day 1 in current week)
	tuesday := currentMonday.AddDate(0, 0, 1)
	task1 := &task.Task{
		ID: 1, Description: "Email check", Category: task.CategoryShallow,
		ScheduledDate: tuesday, ScheduledStart: "09:00", ScheduledEnd: "09:30",
		Status: task.StatusScheduled,
	}
	task2 := &task.Task{
		ID: 2, Description: "Work on brag document", Category: task.CategoryDeep,
		ScheduledDate: tuesday, ScheduledStart: "09:30", ScheduledEnd: "11:30",
		Status: task.StatusScheduled,
	}
	task3 := &task.Task{
		ID: 3, Description: "Send email to clients", Category: task.CategoryShallow,
		ScheduledDate: tuesday, ScheduledStart: "11:30", ScheduledEnd: "11:45",
		Status: task.StatusScheduled,
	}
	task4 := &task.Task{
		ID: 4, Description: "Deep coding", Category: task.CategoryDeep,
		ScheduledDate: tuesday, ScheduledStart: "11:45", ScheduledEnd: "14:45",
		Status: task.StatusScheduled,
	}

	tasks := []*task.Task{task1, task2, task3, task4}
	grid := TasksToSlotGrid(tasks, cfg)

	// Verify all tasks are in the grid
	gridTasks := grid.AllTasks()
	if len(gridTasks) != 4 {
		t.Errorf("Grid has %d tasks, want 4", len(gridTasks))
	}

	// Convert back to WeekWindow
	ww := SlotGridToWeekWindow(grid)
	if ww == nil {
		t.Fatal("expected non-nil WeekWindow")
	}

	// Verify current week has all 4 tasks
	currentWeekTasks := ww.Current().AllTasks()
	if len(currentWeekTasks) != 4 {
		t.Errorf("Current week has %d tasks, want 4. Tasks: ", len(currentWeekTasks))
		for _, task := range currentWeekTasks {
			t.Logf("  - %s (%s-%s)", task.Description, task.ScheduledStart, task.ScheduledEnd)
		}
	}

	// Verify tasks on Tuesday specifically
	tuesdayTasks := ww.Current().Day(1).ScheduledTasks()
	if len(tuesdayTasks) != 4 {
		t.Errorf("Tuesday has %d scheduled tasks, want 4. Tasks:", len(tuesdayTasks))
		for _, task := range tuesdayTasks {
			t.Logf("  - %s (%s-%s)", task.Description, task.ScheduledStart, task.ScheduledEnd)
		}
	}

	// Verify times are correctly transferred
	taskByID := make(map[int64]*task.Task)
	for _, tsk := range tuesdayTasks {
		taskByID[tsk.ID] = tsk
	}

	if t1, ok := taskByID[1]; ok {
		if t1.ScheduledStart != "09:00" || t1.ScheduledEnd != "09:30" {
			t.Errorf("Task 1 times wrong: got %s-%s, want 09:00-09:30", t1.ScheduledStart, t1.ScheduledEnd)
		}
	} else {
		t.Error("Task 1 not found")
	}

	if t2, ok := taskByID[2]; ok {
		if t2.ScheduledStart != "09:30" || t2.ScheduledEnd != "11:30" {
			t.Errorf("Task 2 times wrong: got %s-%s, want 09:30-11:30", t2.ScheduledStart, t2.ScheduledEnd)
		}
	} else {
		t.Error("Task 2 not found")
	}

	if t3, ok := taskByID[3]; ok {
		if t3.ScheduledStart != "11:30" || t3.ScheduledEnd != "11:45" {
			t.Errorf("Task 3 times wrong: got %s-%s, want 11:30-11:45", t3.ScheduledStart, t3.ScheduledEnd)
		}
	} else {
		t.Error("Task 3 not found")
	}

	if t4, ok := taskByID[4]; ok {
		if t4.ScheduledStart != "11:45" || t4.ScheduledEnd != "14:45" {
			t.Errorf("Task 4 times wrong: got %s-%s, want 11:45-14:45", t4.ScheduledStart, t4.ScheduledEnd)
		}
	} else {
		t.Error("Task 4 not found")
	}
}

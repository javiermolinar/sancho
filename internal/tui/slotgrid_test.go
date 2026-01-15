package tui

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/javiermolinar/sancho/internal/task"
)

// testConfig creates a SlotConfig for testing.
// Uses a 24-hour day (96 slots) with working hours 09:00-17:00.
// DisplaySlotSize is set to 2 (30-min blocks) for easier test readability.
func testConfig() SlotConfig {
	// Use a fixed date in the future so nothing is "past"
	futureDate := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	return SlotConfig{
		SlotDuration:      15,
		NumDays:           7, // 1 week for simpler tests
		FirstDate:         futureDate,
		WorkingHoursStart: 540,  // 09:00
		WorkingHoursEnd:   1020, // 17:00
		DisplaySlotSize:   2,    // 30-min blocks for visual movement
		Now: func() time.Time {
			// Return a time before the grid starts, so nothing is "past"
			return futureDate.Add(-24 * time.Hour)
		},
	}
}

// makeTask creates a task with the given ID.
func makeTask(id int64) *task.Task {
	return &task.Task{
		ID:          id,
		Description: "Task " + string(rune('A'+id)),
		Category:    task.CategoryDeep,
		Status:      task.StatusScheduled,
	}
}

// gridFromString creates a SlotGrid from string notation.
// - Letters (A-Z) represent tasks (A=ID 0, B=ID 1, etc.)
// - "-" represents empty slot
// - "|" separates days
//
// Note: With 96 slots per day, we only populate the first N slots
// based on the string length. Remaining slots are empty.
//
// Example: "AABB----|--CC----" = 2 days
//
//	Day 0: Task A (2 slots), Task B (2 slots), 4 empty, rest empty
//	Day 1: 2 empty, Task C (2 slots), 4 empty, rest empty
func gridFromString(s string, cfg SlotConfig) *SlotGrid {
	grid := NewSlotGrid(cfg)
	days := splitDays(s)

	// Create tasks for each unique letter
	tasks := make(map[rune]*task.Task)

	for dayIdx, dayStr := range days {
		if dayIdx >= cfg.NumDays {
			break
		}
		for slot, ch := range dayStr {
			if slot >= SlotsPerDay {
				break
			}
			if ch == '-' {
				continue
			}
			if ch >= 'A' && ch <= 'Z' {
				t, ok := tasks[ch]
				if !ok {
					id := int64(ch - 'A')
					t = makeTask(id)
					tasks[ch] = t
				}
				idx := grid.slotIndex(dayIdx, slot)
				grid.slots[idx] = t
			}
		}
	}

	return grid
}

// splitDays splits a grid string by "|" separator.
func splitDays(s string) []string {
	var days []string
	current := ""
	for _, ch := range s {
		if ch == '|' {
			days = append(days, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		days = append(days, current)
	}
	return days
}

// printDayPrefix returns the first n slots of a day as a string.
// This is useful for comparing test results when we only care about
// the first few slots.
func printDayPrefix(grid *SlotGrid, day, n int) string {
	full := grid.PrintDay(day)
	if len(full) < n {
		return full
	}
	return full[:n]
}

func TestSlotConfig_SlotsPerDay(t *testing.T) {
	cfg := testConfig()
	got := cfg.SlotsPerDay()
	want := 96 // 24 hours * 4 slots/hour
	if got != want {
		t.Errorf("SlotsPerDay() = %d, want %d", got, want)
	}
}

func TestSlotConfig_TotalSlots(t *testing.T) {
	cfg := testConfig()
	got := cfg.TotalSlots()
	want := 672 // 96 slots * 7 days
	if got != want {
		t.Errorf("TotalSlots() = %d, want %d", got, want)
	}
}

func TestSlotConfig_SlotToTime(t *testing.T) {
	cfg := testConfig()
	tests := []struct {
		slot int
		want string
	}{
		{0, "00:00"},
		{4, "01:00"},
		{36, "09:00"}, // 9 * 4 = 36
		{48, "12:00"}, // 12 * 4 = 48
		{95, "23:45"},
	}
	for _, tt := range tests {
		got := cfg.SlotToTime(tt.slot)
		if got != tt.want {
			t.Errorf("SlotToTime(%d) = %q, want %q", tt.slot, got, tt.want)
		}
	}
}

func TestSlotConfig_TimeToSlot(t *testing.T) {
	cfg := testConfig()
	tests := []struct {
		time string
		want int
	}{
		{"00:00", 0},
		{"01:00", 4},
		{"09:00", 36},
		{"12:00", 48},
		{"23:45", 95},
	}
	for _, tt := range tests {
		got := cfg.TimeToSlot(tt.time)
		if got != tt.want {
			t.Errorf("TimeToSlot(%q) = %d, want %d", tt.time, got, tt.want)
		}
	}
}

func TestSlotConfig_IsWorkingHours(t *testing.T) {
	cfg := testConfig()
	tests := []struct {
		slot int
		want bool
	}{
		{35, false}, // 08:45
		{36, true},  // 09:00 (start of working hours)
		{48, true},  // 12:00
		{67, true},  // 16:45
		{68, false}, // 17:00 (end of working hours)
	}
	for _, tt := range tests {
		got := cfg.IsWorkingHours(tt.slot)
		if got != tt.want {
			t.Errorf("IsWorkingHours(%d) = %v, want %v", tt.slot, got, tt.want)
		}
	}
}

func TestNewSlotGrid(t *testing.T) {
	cfg := testConfig()
	grid := NewSlotGrid(cfg)

	if len(grid.slots) != cfg.TotalSlots() {
		t.Errorf("len(slots) = %d, want %d", len(grid.slots), cfg.TotalSlots())
	}

	// All slots should be nil (empty)
	for i, slot := range grid.slots {
		if slot != nil {
			t.Errorf("slot[%d] should be nil, got %v", i, slot)
		}
	}
}

func TestGridFromString(t *testing.T) {
	cfg := testConfig()

	tests := []struct {
		name     string
		input    string
		wantDay0 string // prefix to compare
		wantDay1 string // prefix to compare
	}{
		{
			name:     "single task",
			input:    "AA------",
			wantDay0: "AA------",
			wantDay1: "--------",
		},
		{
			name:     "two tasks same day",
			input:    "AABB----",
			wantDay0: "AABB----",
			wantDay1: "--------",
		},
		{
			name:     "two days",
			input:    "AAAABB--|--CC----",
			wantDay0: "AAAABB--",
			wantDay1: "--CC----",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid := gridFromString(tt.input, cfg)

			gotDay0 := printDayPrefix(grid, 0, len(tt.wantDay0))
			if gotDay0 != tt.wantDay0 {
				t.Errorf("Day 0 = %q, want %q", gotDay0, tt.wantDay0)
			}

			gotDay1 := printDayPrefix(grid, 1, len(tt.wantDay1))
			if gotDay1 != tt.wantDay1 {
				t.Errorf("Day 1 = %q, want %q", gotDay1, tt.wantDay1)
			}
		})
	}
}

func TestSlotGrid_TaskAt(t *testing.T) {
	cfg := testConfig()
	grid := gridFromString("AABB----", cfg)

	tests := []struct {
		day     int
		slot    int
		wantID  int64
		wantNil bool
	}{
		{day: 0, slot: 0, wantID: 0, wantNil: false}, // A
		{day: 0, slot: 1, wantID: 0, wantNil: false}, // A
		{day: 0, slot: 2, wantID: 1, wantNil: false}, // B
		{day: 0, slot: 3, wantID: 1, wantNil: false}, // B
		{day: 0, slot: 4, wantNil: true},             // empty
		{day: 0, slot: 7, wantNil: true},             // empty
		{day: 1, slot: 0, wantNil: true},             // day 1 is empty
	}

	for _, tt := range tests {
		got := grid.TaskAt(tt.day, tt.slot)
		if tt.wantNil {
			if got != nil {
				t.Errorf("TaskAt(%d, %d) = %v, want nil", tt.day, tt.slot, got)
			}
		} else {
			if got == nil {
				t.Errorf("TaskAt(%d, %d) = nil, want task with ID %d", tt.day, tt.slot, tt.wantID)
			} else if got.ID != tt.wantID {
				t.Errorf("TaskAt(%d, %d).ID = %d, want %d", tt.day, tt.slot, got.ID, tt.wantID)
			}
		}
	}
}

func TestSlotGrid_IsEmpty(t *testing.T) {
	cfg := testConfig()
	grid := gridFromString("AA------", cfg)

	if grid.IsEmpty(0, 0) {
		t.Error("IsEmpty(0, 0) = true, want false (has task A)")
	}
	if !grid.IsEmpty(0, 4) {
		t.Error("IsEmpty(0, 4) = false, want true")
	}
}

func TestSlotGrid_FindTask(t *testing.T) {
	cfg := testConfig()
	grid := gridFromString("AAAABB--|--CC----", cfg)

	tests := []struct {
		name      string
		taskID    int64
		wantDay   int
		wantStart int
		wantEnd   int
		wantFound bool
	}{
		{
			name:      "find task A",
			taskID:    0,
			wantDay:   0,
			wantStart: 0,
			wantEnd:   4,
			wantFound: true,
		},
		{
			name:      "find task B",
			taskID:    1,
			wantDay:   0,
			wantStart: 4,
			wantEnd:   6,
			wantFound: true,
		},
		{
			name:      "find task C",
			taskID:    2,
			wantDay:   1,
			wantStart: 2,
			wantEnd:   4,
			wantFound: true,
		},
		{
			name:      "task not found",
			taskID:    99,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			day, start, end, found := grid.FindTaskByID(tt.taskID)

			if found != tt.wantFound {
				t.Errorf("found = %v, want %v", found, tt.wantFound)
				return
			}

			if !found {
				return
			}

			if day != tt.wantDay {
				t.Errorf("day = %d, want %d", day, tt.wantDay)
			}
			if start != tt.wantStart {
				t.Errorf("start = %d, want %d", start, tt.wantStart)
			}
			if end != tt.wantEnd {
				t.Errorf("end = %d, want %d", end, tt.wantEnd)
			}
		})
	}
}

func TestSlotGrid_AllTasks(t *testing.T) {
	cfg := testConfig()
	grid := gridFromString("AAAABB--|--CC----", cfg)

	tasks := grid.AllTasks()

	if len(tasks) != 3 {
		t.Errorf("len(AllTasks()) = %d, want 3", len(tasks))
	}

	// Check that we have tasks A, B, C (IDs 0, 1, 2)
	ids := make(map[int64]bool)
	for _, tsk := range tasks {
		ids[tsk.ID] = true
	}

	for _, id := range []int64{0, 1, 2} {
		if !ids[id] {
			t.Errorf("missing task with ID %d", id)
		}
	}
}

func TestSlotGrid_TasksOnDay(t *testing.T) {
	cfg := testConfig()
	grid := gridFromString("AAAABB--|--CC----|DD------", cfg)

	tests := []struct {
		day     int
		wantIDs []int64
	}{
		{day: 0, wantIDs: []int64{0, 1}}, // A, B
		{day: 1, wantIDs: []int64{2}},    // C
		{day: 2, wantIDs: []int64{3}},    // D
		{day: 3, wantIDs: []int64{}},     // empty
	}

	for _, tt := range tests {
		tasks := grid.TasksOnDay(tt.day)

		if len(tasks) != len(tt.wantIDs) {
			t.Errorf("TasksOnDay(%d): got %d tasks, want %d", tt.day, len(tasks), len(tt.wantIDs))
			continue
		}

		ids := make(map[int64]bool)
		for _, tsk := range tasks {
			ids[tsk.ID] = true
		}

		for _, id := range tt.wantIDs {
			if !ids[id] {
				t.Errorf("TasksOnDay(%d): missing task with ID %d", tt.day, id)
			}
		}
	}
}

func TestSlotGrid_Place(t *testing.T) {
	cfg := testConfig()

	tests := []struct {
		name      string
		initial   string
		task      *task.Task
		day       int
		startSlot int
		numSlots  int
		wantErr   error
		wantDay0  string
	}{
		{
			name:      "place in empty grid",
			initial:   "--------",
			task:      makeTask(0), // A
			day:       0,
			startSlot: 0,
			numSlots:  2,
			wantErr:   nil,
			wantDay0:  "AA------",
		},
		{
			name:      "place after existing",
			initial:   "AA------",
			task:      makeTask(1), // B
			day:       0,
			startSlot: 2,
			numSlots:  2,
			wantErr:   nil,
			wantDay0:  "AABB----",
		},
		{
			name:      "place at end of day",
			initial:   "--------",
			task:      makeTask(0),
			day:       0,
			startSlot: 6,
			numSlots:  2,
			wantErr:   nil,
			wantDay0:  "------AA",
		},
		{
			name:      "slot occupied error",
			initial:   "AA------",
			task:      makeTask(1), // B trying to overlap with A
			day:       0,
			startSlot: 1,
			numSlots:  2,
			wantErr:   ErrSlotOccupied,
		},
		{
			name:      "invalid position error",
			initial:   "--------",
			task:      makeTask(0),
			day:       -1,
			startSlot: 0,
			numSlots:  2,
			wantErr:   ErrInvalidSlotPosition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid := gridFromString(tt.initial, cfg)
			newGrid, err := grid.Place(tt.task, tt.day, tt.startSlot, tt.numSlots)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Original grid should be unchanged
			gotOrig := printDayPrefix(grid, 0, len(tt.initial))
			if gotOrig == tt.wantDay0 && tt.initial != tt.wantDay0 {
				t.Error("original grid was mutated")
			}

			// New grid should have the expected state
			gotNew := printDayPrefix(newGrid, 0, len(tt.wantDay0))
			if gotNew != tt.wantDay0 {
				t.Errorf("newGrid Day 0 = %q, want %q", gotNew, tt.wantDay0)
			}
		})
	}
}

func TestSlotGrid_Clone(t *testing.T) {
	cfg := testConfig()
	grid := gridFromString("AABB----", cfg)

	cloned := grid.clone()

	// Should have same content
	got := printDayPrefix(cloned, 0, 8)
	want := printDayPrefix(grid, 0, 8)
	if got != want {
		t.Errorf("cloned content differs: got %q, want %q", got, want)
	}

	// Modifying cloned should not affect original
	cloned.slots[0] = nil
	if grid.TaskAt(0, 0) == nil {
		t.Error("modifying clone affected original")
	}
}

func TestSlotGrid_CurrentTimePosition(t *testing.T) {
	baseDate := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		now      time.Time
		wantDay  int
		wantSlot int
	}{
		{
			name:     "at midnight",
			now:      time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
			wantDay:  0,
			wantSlot: 0,
		},
		{
			name:     "mid morning",
			now:      time.Date(2030, 1, 1, 9, 30, 0, 0, time.UTC),
			wantDay:  0,
			wantSlot: 38, // 9*4 + 2 = 38
		},
		{
			name:     "second day",
			now:      time.Date(2030, 1, 2, 10, 0, 0, 0, time.UTC),
			wantDay:  1,
			wantSlot: 40, // 10*4 = 40
		},
		{
			name:     "before grid start",
			now:      time.Date(2029, 12, 31, 12, 0, 0, 0, time.UTC),
			wantDay:  -1,
			wantSlot: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := SlotConfig{
				SlotDuration:      15,
				NumDays:           7,
				FirstDate:         baseDate,
				WorkingHoursStart: 540,
				WorkingHoursEnd:   1020,
				Now:               func() time.Time { return tt.now },
			}

			grid := NewSlotGrid(cfg)
			day, slot := grid.currentTimePosition()

			if day != tt.wantDay {
				t.Errorf("day = %d, want %d", day, tt.wantDay)
			}
			if slot != tt.wantSlot {
				t.Errorf("slot = %d, want %d", slot, tt.wantSlot)
			}
		})
	}
}

func TestSlotGrid_CanModifyTask(t *testing.T) {
	// Now is 2030-01-01 09:30 (day 0, slot 38)
	now := time.Date(2030, 1, 1, 9, 30, 0, 0, time.UTC)
	cfg := SlotConfig{
		SlotDuration:      15,
		NumDays:           7,
		FirstDate:         now,
		WorkingHoursStart: 540,
		WorkingHoursEnd:   1020,
		Now:               func() time.Time { return now },
	}

	// Create a grid with tasks at different slots
	grid := NewSlotGrid(cfg)

	// Task A at slots 0-1 (past - before 09:30)
	taskA := makeTask(0)
	grid, _ = grid.Place(taskA, 0, 0, 2)

	// Task B at slots 38-39 (current time)
	taskB := makeTask(1)
	grid, _ = grid.Place(taskB, 0, 38, 2)

	// Task C at slots 50-51 (future)
	taskC := makeTask(2)
	grid, _ = grid.Place(taskC, 0, 50, 2)

	tests := []struct {
		name    string
		task    *task.Task
		wantErr error
	}{
		{
			name:    "task A in past",
			task:    taskA,
			wantErr: ErrTaskAlreadyStarted,
		},
		{
			name:    "task B at current time",
			task:    taskB,
			wantErr: ErrTaskAlreadyStarted,
		},
		{
			name:    "task C in future",
			task:    taskC,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := grid.canModifyTask(tt.task)
			if err != tt.wantErr {
				t.Errorf("canModifyTask() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestSlotGrid_IsPastPosition(t *testing.T) {
	// Now is 2030-01-01 09:30 (day 0, slot 38)
	now := time.Date(2030, 1, 1, 9, 30, 0, 0, time.UTC)
	cfg := SlotConfig{
		SlotDuration:      15,
		NumDays:           7,
		FirstDate:         now,
		WorkingHoursStart: 540,
		WorkingHoursEnd:   1020,
		Now:               func() time.Time { return now },
	}

	grid := NewSlotGrid(cfg)

	tests := []struct {
		day      int
		slot     int
		wantPast bool
	}{
		{day: 0, slot: 0, wantPast: true},   // before now
		{day: 0, slot: 37, wantPast: true},  // just before now
		{day: 0, slot: 38, wantPast: true},  // at now (considered past)
		{day: 0, slot: 39, wantPast: false}, // after now
		{day: 0, slot: 95, wantPast: false}, // end of day
		{day: 1, slot: 0, wantPast: false},  // next day
	}

	for _, tt := range tests {
		got := grid.isPastPosition(tt.day, tt.slot)
		if got != tt.wantPast {
			t.Errorf("isPastPosition(%d, %d) = %v, want %v", tt.day, tt.slot, got, tt.wantPast)
		}
	}
}

// =============================================================================
// MoveDown Tests
// =============================================================================

func TestSlotGrid_MoveDown(t *testing.T) {
	cfg := testConfig()

	tests := []struct {
		name       string
		initial    string
		taskLetter rune
		wantDay0   string
	}{
		{
			name:       "swap with adjacent task",
			initial:    "AABB----",
			taskLetter: 'A',
			wantDay0:   "BBAA----",
		},
		{
			name:       "move into small gap",
			initial:    "AA--BB--",
			taskLetter: 'A',
			wantDay0:   "--AABB--", // A moves 2 slots (its size) into gap
		},
		{
			name:       "swap with second task when multiple",
			initial:    "AABBCC--",
			taskLetter: 'A',
			wantDay0:   "BBAACC--", // A swaps with B (single swap, not cascade)
		},
		{
			name:       "move one step into large gap",
			initial:    "AA----BB",
			taskLetter: 'A',
			wantDay0:   "--AA--BB", // A moves 2 slots (its size), not to end of gap
		},
		{
			name:       "at end of tasks - move into empty space",
			initial:    "AABB----",
			taskLetter: 'B',
			wantDay0:   "AA--BB--", // B moves 2 slots (its size) into gap
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid := gridFromString(tt.initial, cfg)
			taskID := int64(tt.taskLetter - 'A')
			tsk := grid.TaskAt(0, findTaskStart(grid, taskID))

			newGrid, err := grid.MoveDown(tsk)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			got := printDayPrefix(newGrid, 0, len(tt.wantDay0))
			if got != tt.wantDay0 {
				t.Errorf("result = %q, want %q", got, tt.wantDay0)
			}

			// Verify immutability
			origGot := printDayPrefix(grid, 0, len(tt.initial))
			if origGot != tt.initial {
				t.Errorf("original grid mutated: got %q, want %q", origGot, tt.initial)
			}
		})
	}
}

func TestSlotGrid_MoveDown_PreservesLength(t *testing.T) {
	cfg := testConfig()
	grid := gridFromString("AABBCC--", cfg)
	taskB := grid.TaskAt(0, 2)
	_, startSlot, endSlot, found := grid.FindTask(taskB)
	if !found {
		t.Fatal("task B not found")
	}
	origLen := endSlot - startSlot

	newGrid, err := grid.MoveDown(taskB)
	if err != nil {
		t.Fatalf("MoveDown failed: %v", err)
	}

	_, newStart, newEnd, found := newGrid.FindTask(taskB)
	if !found {
		t.Fatal("task B not found after move")
	}
	newLen := newEnd - newStart
	if newLen != origLen {
		t.Errorf("task length changed: got %d, want %d", newLen, origLen)
	}
}

// =============================================================================
// MoveUp Tests
// =============================================================================

func TestSlotGrid_MoveUp(t *testing.T) {
	cfg := testConfig()

	tests := []struct {
		name       string
		initial    string
		taskLetter rune
		wantDay0   string
	}{
		{
			name:       "swap with adjacent task",
			initial:    "AABB----",
			taskLetter: 'B',
			wantDay0:   "BBAA----",
		},
		{
			name:       "move into gap",
			initial:    "--AABB--",
			taskLetter: 'A',
			wantDay0:   "AA--BB--", // A moves 2 slots (its size) up
		},
		{
			name:       "at start - no op",
			initial:    "AA------",
			taskLetter: 'A',
			wantDay0:   "AA------",
		},
		{
			name:       "move one step up in large gap",
			initial:    "----AABB",
			taskLetter: 'A',
			wantDay0:   "--AA--BB", // A moves 2 slots (its size), not to start of gap
		},
		{
			name:       "swap with previous when multiple tasks",
			initial:    "--AABBCC",
			taskLetter: 'C',
			wantDay0:   "--AACCBB", // C swaps with B (single swap, not cascade)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid := gridFromString(tt.initial, cfg)
			taskID := int64(tt.taskLetter - 'A')
			tsk := grid.TaskAt(0, findTaskStart(grid, taskID))

			newGrid, err := grid.MoveUp(tsk)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			got := printDayPrefix(newGrid, 0, len(tt.wantDay0))
			if got != tt.wantDay0 {
				t.Errorf("result = %q, want %q", got, tt.wantDay0)
			}
		})
	}
}

// =============================================================================
// MoveRight Tests
// =============================================================================

func TestSlotGrid_MoveRight(t *testing.T) {
	cfg := testConfig()

	tests := []struct {
		name     string
		initial  string
		taskDay  int
		taskSlot int
		wantDay0 string
		wantDay1 string
	}{
		{
			name:     "move to empty day",
			initial:  "AABB----|--------",
			taskDay:  0,
			taskSlot: 0, // A
			wantDay0: "BB------",
			wantDay1: "AA------",
		},
		{
			name:     "move with shift on target",
			initial:  "AABB----|CC------",
			taskDay:  0,
			taskSlot: 0, // A
			wantDay0: "BB------",
			wantDay1: "AACC----",
		},
		{
			name:     "source cascade",
			initial:  "AABBCC--|--------",
			taskDay:  0,
			taskSlot: 0, // A
			wantDay0: "BBCC----",
			wantDay1: "AA------",
		},
		{
			name:     "insert into middle of task inserts after it",
			initial:  "--AA----|BBBBBBBB",
			taskDay:  0,
			taskSlot: 2, // A starts at slot 2
			wantDay0: "--------",
			wantDay1: "BBBBBBBBAA", // A inserted after B (which ends at slot 8)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid := gridFromString(tt.initial, cfg)
			tsk := grid.TaskAt(tt.taskDay, tt.taskSlot)

			newGrid, err := grid.MoveRight(tsk)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			got0 := printDayPrefix(newGrid, 0, len(tt.wantDay0))
			if got0 != tt.wantDay0 {
				t.Errorf("Day 0 = %q, want %q", got0, tt.wantDay0)
			}

			got1 := printDayPrefix(newGrid, 1, len(tt.wantDay1))
			if got1 != tt.wantDay1 {
				t.Errorf("Day 1 = %q, want %q", got1, tt.wantDay1)
			}
		})
	}
}

func TestSlotGrid_MoveRight_Overflow(t *testing.T) {
	cfg := testConfig()

	// Create a grid where target day is packed
	// Use slots near the end of day (96 total)
	grid := NewSlotGrid(cfg)

	// Place task A (2 slots) at slot 0 on day 0
	taskA := makeTask(0)
	grid, _ = grid.Place(taskA, 0, 0, 2)

	// Place task B (4 slots) at slot 94-95 on day 1 (would overflow if shifted)
	taskB := makeTask(1)
	grid, _ = grid.Place(taskB, 1, 94, 2)

	// MoveRight should be no-op because shifting B would overflow
	newGrid, err := grid.MoveRight(taskA)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should be same grid (no-op)
	if newGrid != grid {
		// Check if A is still on day 0
		_, start, _, found := newGrid.FindTask(taskA)
		if !found || start != 0 {
			// Moved - that's wrong, it should be no-op
			t.Error("MoveRight should be no-op when target would overflow")
		}
	}
}

func TestSlotGrid_MoveRight_PastValidation(t *testing.T) {
	// Now is day 0, slot 10
	now := time.Date(2030, 1, 1, 2, 30, 0, 0, time.UTC) // 02:30 = slot 10
	cfg := SlotConfig{
		SlotDuration:      15,
		NumDays:           7,
		FirstDate:         now,
		WorkingHoursStart: 540,
		WorkingHoursEnd:   1020,
		Now:               func() time.Time { return now },
	}

	grid := NewSlotGrid(cfg)

	// Task A at slots 0-1 (past)
	taskA := makeTask(0)
	grid, _ = grid.Place(taskA, 0, 0, 2)

	// Task B at slots 20-21 (future)
	taskB := makeTask(1)
	grid, _ = grid.Place(taskB, 0, 20, 2)

	// MoveRight on past task should fail
	_, err := grid.MoveRight(taskA)
	if err != ErrTaskAlreadyStarted {
		t.Errorf("MoveRight past task: err = %v, want ErrTaskAlreadyStarted", err)
	}

	// MoveRight on future task should succeed
	newGrid, err := grid.MoveRight(taskB)
	if err != nil {
		t.Errorf("MoveRight future task: unexpected error: %v", err)
	}

	// B should now be on day 1
	day, _, _, found := newGrid.FindTask(taskB)
	if !found || day != 1 {
		t.Errorf("Task B should be on day 1, got day %d, found %v", day, found)
	}
}

// =============================================================================
// MoveLeft Tests
// =============================================================================

func TestSlotGrid_MoveLeft(t *testing.T) {
	cfg := testConfig()

	tests := []struct {
		name     string
		initial  string
		taskDay  int
		taskSlot int
		wantDay0 string
		wantDay1 string
	}{
		{
			name:     "move to empty day",
			initial:  "--------|AABB----",
			taskDay:  1,
			taskSlot: 0, // A
			wantDay0: "AA------",
			wantDay1: "BB------",
		},
		{
			name:     "move with shift on target",
			initial:  "CC------|AABB----",
			taskDay:  1,
			taskSlot: 0, // A
			wantDay0: "AACC----",
			wantDay1: "BB------",
		},
		{
			name:     "source cascade",
			initial:  "--------|AABBCC--",
			taskDay:  1,
			taskSlot: 0, // A
			wantDay0: "AA------",
			wantDay1: "BBCC----",
		},
		{
			name:     "insert into middle of task inserts after it",
			initial:  "BBBBBBBB|--AA----",
			taskDay:  1,
			taskSlot: 2,            // A starts at slot 2
			wantDay0: "BBBBBBBBAA", // A inserted after B (which ends at slot 8)
			wantDay1: "--------",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid := gridFromString(tt.initial, cfg)
			tsk := grid.TaskAt(tt.taskDay, tt.taskSlot)

			newGrid, err := grid.MoveLeft(tsk)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			got0 := printDayPrefix(newGrid, 0, len(tt.wantDay0))
			if got0 != tt.wantDay0 {
				t.Errorf("Day 0 = %q, want %q", got0, tt.wantDay0)
			}

			got1 := printDayPrefix(newGrid, 1, len(tt.wantDay1))
			if got1 != tt.wantDay1 {
				t.Errorf("Day 1 = %q, want %q", got1, tt.wantDay1)
			}
		})
	}
}

func TestSlotGrid_MoveLeft_NoOpAtDayZero(t *testing.T) {
	cfg := testConfig()

	grid := gridFromString("AABB----|--------", cfg)
	taskA := grid.TaskAt(0, 0)

	// MoveLeft should be no-op at day 0
	newGrid, err := grid.MoveLeft(taskA)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should be same grid (no-op)
	if newGrid != grid {
		t.Error("MoveLeft should be no-op when at day 0")
	}
}

func TestSlotGrid_MoveLeft_Overflow(t *testing.T) {
	cfg := testConfig()

	// Create a grid where target day has a task near the end
	grid := NewSlotGrid(cfg)

	// Place task A (2 slots) at slot 0 on day 1
	taskA := makeTask(0)
	grid, _ = grid.Place(taskA, 1, 0, 2)

	// Place task B (2 slots) at slot 94-95 on day 0 (would overflow if shifted)
	taskB := makeTask(1)
	grid, _ = grid.Place(taskB, 0, 94, 2)

	// MoveLeft should be no-op because shifting B would overflow
	newGrid, err := grid.MoveLeft(taskA)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should be same grid (no-op)
	if newGrid != grid {
		// Check if A is still on day 1
		day, _, _, found := newGrid.FindTask(taskA)
		if !found || day != 1 {
			t.Error("MoveLeft should be no-op when target would overflow")
		}
	}
}

func TestSlotGrid_MoveLeft_PastValidation(t *testing.T) {
	// Now is day 1, slot 40 (10:00)
	// - Task A at day 1, slot 20-21 (05:00-05:30) -> source is in past, should error
	// - Task B at day 1, slot 50-51 (12:30-13:00) -> source is future, target day 0 slot 50 is past (before day 1 slot 40)
	// - Task C at day 2, slot 50-51 (12:30-13:00) -> source is future, target day 1 slot 50 is future, should succeed
	now := time.Date(2030, 1, 2, 10, 0, 0, 0, time.UTC) // Day 1, 10:00 = slot 40
	cfg := SlotConfig{
		SlotDuration:      15,
		NumDays:           7,
		FirstDate:         time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
		WorkingHoursStart: 540,
		WorkingHoursEnd:   1020,
		Now:               func() time.Time { return now },
	}

	grid := NewSlotGrid(cfg)

	// Task A at day 1, slots 20-21 (source is in past - before slot 40)
	taskA := makeTask(0)
	grid, _ = grid.Place(taskA, 1, 20, 2)

	// Task B at day 1, slots 50-51 (source is future, but target day 0 slot 50 is in past)
	taskB := makeTask(1)
	grid, _ = grid.Place(taskB, 1, 50, 2)

	// Task C at day 2, slots 50-51 (source is future, target day 1 slot 50 is also future)
	taskC := makeTask(2)
	grid, _ = grid.Place(taskC, 2, 50, 2)

	// MoveLeft on task A (source in past) should return error
	_, err := grid.MoveLeft(taskA)
	if err != ErrTaskAlreadyStarted {
		t.Errorf("MoveLeft source in past: got err = %v, want ErrTaskAlreadyStarted", err)
	}

	// MoveLeft on task B (target in past) should be no-op
	newGridB, err := grid.MoveLeft(taskB)
	if err != nil {
		t.Errorf("MoveLeft target in past: unexpected error: %v", err)
	}
	// Should be no-op - B still on day 1
	dayB, _, _, foundB := newGridB.FindTask(taskB)
	if !foundB || dayB != 1 {
		t.Errorf("Task B should still be on day 1, got day %d", dayB)
	}

	// MoveLeft on task C (both source and target in future) should succeed
	newGridC, err := grid.MoveLeft(taskC)
	if err != nil {
		t.Errorf("MoveLeft future: unexpected error: %v", err)
	}

	// C should now be on day 1
	dayC, _, _, foundC := newGridC.FindTask(taskC)
	if !foundC || dayC != 1 {
		t.Errorf("Task C should be on day 1, got day %d, found %v", dayC, foundC)
	}
}

// =============================================================================
// Grow Tests
// =============================================================================

func TestSlotGrid_Grow(t *testing.T) {
	cfg := testConfig()

	tests := []struct {
		name       string
		initial    string
		taskLetter rune
		wantDay0   string
	}{
		{
			name:       "grow into empty space",
			initial:    "AA------",
			taskLetter: 'A',
			wantDay0:   "AAA-----",
		},
		{
			name:       "grow with shift",
			initial:    "AABB----",
			taskLetter: 'A',
			wantDay0:   "AAABB---",
		},
		{
			name:       "grow cascade shift",
			initial:    "AABBCC--",
			taskLetter: 'A',
			wantDay0:   "AAABBCC-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid := gridFromString(tt.initial, cfg)
			taskID := int64(tt.taskLetter - 'A')
			tsk := grid.TaskAt(0, findTaskStart(grid, taskID))

			newGrid, err := grid.Grow(tsk)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			got := printDayPrefix(newGrid, 0, len(tt.wantDay0))
			if got != tt.wantDay0 {
				t.Errorf("result = %q, want %q", got, tt.wantDay0)
			}
		})
	}
}

func TestSlotGrid_Grow_NoOp(t *testing.T) {
	cfg := testConfig()

	// Task at end of day
	grid := NewSlotGrid(cfg)
	taskA := makeTask(0)
	grid, _ = grid.Place(taskA, 0, 94, 2) // slots 94-95

	newGrid, err := grid.Grow(taskA)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should be same grid (no-op at day end)
	_, _, end, _ := newGrid.FindTask(taskA)
	if end != 96 {
		t.Errorf("expected task to still end at slot 96, got %d", end)
	}
}

// =============================================================================
// Shrink Tests
// =============================================================================

func TestSlotGrid_Shrink(t *testing.T) {
	cfg := testConfig()

	tests := []struct {
		name       string
		initial    string
		taskLetter rune
		wantDay0   string
		wantErr    error
	}{
		{
			name:       "shrink 3 slots to 2",
			initial:    "AAA-----",
			taskLetter: 'A',
			wantDay0:   "AA------",
		},
		{
			name:       "shrink 2 slots to 1",
			initial:    "AABB----",
			taskLetter: 'A',
			wantDay0:   "A-BB----",
		},
		{
			name:       "shrink minimum - error",
			initial:    "A-------",
			taskLetter: 'A',
			wantErr:    ErrMinimumSlotsDuration,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid := gridFromString(tt.initial, cfg)
			taskID := int64(tt.taskLetter - 'A')
			tsk := grid.TaskAt(0, findTaskStart(grid, taskID))

			newGrid, err := grid.Shrink(tsk)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			got := printDayPrefix(newGrid, 0, len(tt.wantDay0))
			if got != tt.wantDay0 {
				t.Errorf("result = %q, want %q", got, tt.wantDay0)
			}
		})
	}
}

// =============================================================================
// AddSpace Tests
// =============================================================================

func TestSlotGrid_AddSpace(t *testing.T) {
	cfg := testConfig()

	tests := []struct {
		name       string
		initial    string
		taskLetter rune
		wantDay0   string
	}{
		{
			name:       "add space shifts one task",
			initial:    "AABB----",
			taskLetter: 'A',
			wantDay0:   "AA-BB---",
		},
		{
			name:       "add space cascade",
			initial:    "AABBCC--",
			taskLetter: 'A',
			wantDay0:   "AA-BBCC-",
		},
		{
			name:       "add space at end - nothing to shift",
			initial:    "AA------",
			taskLetter: 'A',
			wantDay0:   "AA------", // no change, already space
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid := gridFromString(tt.initial, cfg)
			taskID := int64(tt.taskLetter - 'A')
			tsk := grid.TaskAt(0, findTaskStart(grid, taskID))

			newGrid, err := grid.AddSpace(tsk)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			got := printDayPrefix(newGrid, 0, len(tt.wantDay0))
			if got != tt.wantDay0 {
				t.Errorf("result = %q, want %q", got, tt.wantDay0)
			}
		})
	}
}

func TestSlotGrid_AddSpaceAt_EmptySlotShiftsRight(t *testing.T) {
	cfg := testConfig()
	grid := gridFromString("AA--BB--", cfg)

	newGrid, err := grid.AddSpaceAt(0, 2)
	if err != nil {
		t.Fatalf("AddSpaceAt failed: %v", err)
	}

	got := printDayPrefix(newGrid, 0, 8)
	want := "AA---BB-"
	if got != want {
		t.Errorf("day after AddSpaceAt = %q, want %q", got, want)
	}
}

func TestSlotGrid_AddSpace_AllowsOngoingTask(t *testing.T) {
	now := time.Date(2030, 1, 1, 9, 30, 0, 0, time.UTC)
	cfg := SlotConfig{
		SlotDuration:    15,
		NumDays:         7,
		FirstDate:       time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
		DisplaySlotSize: 1,
		Now:             func() time.Time { return now },
	}
	grid := NewSlotGrid(cfg)
	taskA := makeTask(0)
	taskB := makeTask(1)
	var err error
	grid, err = grid.Place(taskA, 0, 36, 4) // 09:00-10:00
	if err != nil {
		t.Fatalf("place task A: %v", err)
	}
	grid, err = grid.Place(taskB, 0, 40, 4) // 10:00-11:00
	if err != nil {
		t.Fatalf("place task B: %v", err)
	}

	newGrid, err := grid.AddSpace(taskA)
	if err != nil {
		t.Fatalf("AddSpace failed: %v", err)
	}

	got := printDayPrefix(newGrid, 0, 45)
	want := "------------------------------------AAAA-BBBB"
	if got != want {
		t.Errorf("day after AddSpace = %q, want %q", got, want)
	}
}

// =============================================================================
// RemoveSpace Tests
// =============================================================================

func TestSlotGrid_RemoveSpaceAt_ShiftsLeft(t *testing.T) {
	cfg := testConfig()
	grid := gridFromString("AABB-CC--", cfg)

	newGrid, err := grid.RemoveSpaceAt(0, 4)
	if err != nil {
		t.Fatalf("RemoveSpaceAt failed: %v", err)
	}

	got := printDayPrefix(newGrid, 0, 9)
	want := "AABBCC---"
	if got != want {
		t.Errorf("day after RemoveSpaceAt = %q, want %q", got, want)
	}
}

func TestSlotGrid_RemoveSpaceAt_NoGap(t *testing.T) {
	cfg := testConfig()
	tests := []struct {
		name string
		grid string
		slot int
		want error
	}{
		{
			name: "slot occupied",
			grid: "AABB----",
			slot: 1,
			want: ErrNoGapToRemove,
		},
		{
			name: "no tasks after gap",
			grid: "AABB----",
			slot: 4,
			want: ErrNoGapToRemove,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid := gridFromString(tt.grid, cfg)
			_, err := grid.RemoveSpaceAt(0, tt.slot)
			if !errors.Is(err, tt.want) {
				t.Fatalf("RemoveSpaceAt error = %v, want %v", err, tt.want)
			}
		})
	}
}

// =============================================================================
// Delete Tests
// =============================================================================

func TestSlotGrid_Delete(t *testing.T) {
	cfg := testConfig()

	tests := []struct {
		name       string
		initial    string
		taskLetter rune
		wantDay0   string
	}{
		{
			name:       "delete with shift",
			initial:    "AABBCC--",
			taskLetter: 'B',
			wantDay0:   "AACC----",
		},
		{
			name:       "delete first task",
			initial:    "AABBCC--",
			taskLetter: 'A',
			wantDay0:   "BBCC----",
		},
		{
			name:       "delete last task",
			initial:    "AABBCC--",
			taskLetter: 'C',
			wantDay0:   "AABB----",
		},
		{
			name:       "delete only task",
			initial:    "AA------",
			taskLetter: 'A',
			wantDay0:   "--------",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid := gridFromString(tt.initial, cfg)
			taskID := int64(tt.taskLetter - 'A')
			tsk := grid.TaskAt(0, findTaskStart(grid, taskID))

			newGrid, err := grid.Delete(tsk)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			got := printDayPrefix(newGrid, 0, len(tt.wantDay0))
			if got != tt.wantDay0 {
				t.Errorf("result = %q, want %q", got, tt.wantDay0)
			}
		})
	}
}

// =============================================================================
// Helper functions
// =============================================================================

// findTaskStart finds the start slot of a task by ID on day 0.
func findTaskStart(grid *SlotGrid, taskID int64) int {
	for s := 0; s < SlotsPerDay; s++ {
		tsk := grid.TaskAt(0, s)
		if tsk != nil && tsk.ID == taskID {
			return s
		}
	}
	return -1
}

// Ensure we use strings package (for potential future use)
var _ = strings.TrimSpace

// Package tui provides the terminal user interface for sancho.
package tui

import (
	"testing"
	"time"

	"github.com/javiermolinar/sancho/internal/config"
	"github.com/javiermolinar/sancho/internal/task"
)

func TestMovingTaskSlotCount(t *testing.T) {
	cfg := &config.Config{
		Schedule: config.ScheduleConfig{
			DayStart: "09:00",
			DayEnd:   "17:00",
		},
	}

	tests := []struct {
		name          string
		taskStart     string
		taskEnd       string
		rowHeight     int
		expectedSlots int
	}{
		{
			name:          "1 hour task with 15min slots",
			taskStart:     "09:00",
			taskEnd:       "10:00",
			rowHeight:     15,
			expectedSlots: 4, // 60min / 15min = 4 slots
		},
		{
			name:          "2 hour task with 15min slots",
			taskStart:     "09:00",
			taskEnd:       "11:00",
			rowHeight:     15,
			expectedSlots: 8, // 120min / 15min = 8 slots
		},
		{
			name:          "30min task with 30min slots",
			taskStart:     "09:00",
			taskEnd:       "09:30",
			rowHeight:     30,
			expectedSlots: 1, // 30min / 30min = 1 slot
		},
		{
			name:          "15min task with 30min slots",
			taskStart:     "09:00",
			taskEnd:       "09:15",
			rowHeight:     30,
			expectedSlots: 1, // 15min / 30min = 0.5, clamped to 1
		},
		{
			name:          "1 hour task with 60min slots",
			taskStart:     "09:00",
			taskEnd:       "10:00",
			rowHeight:     60,
			expectedSlots: 1, // 60min / 60min = 1 slot
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create model
			m := New(nil, cfg)
			m.rowHeight = tt.rowHeight

			// Create a task
			monday := time.Date(2025, 1, 6, 0, 0, 0, 0, time.Local)
			testTask := &task.Task{
				ID:             1,
				Description:    "Test task",
				Category:       task.CategoryDeep,
				ScheduledDate:  monday,
				ScheduledStart: tt.taskStart,
				ScheduledEnd:   tt.taskEnd,
				Status:         task.StatusScheduled,
			}

			// Create week with the task and convert to slot grid
			week := task.NewWeek(monday)
			_ = week.Day(0).AddTask(testTask)
			ww := task.NewWeekWindow(nil, week, nil)

			// Create a config that matches our test week
			slotConfig := SlotGridConfigFromWeekWindow(ww, cfg.Schedule.DayStart, cfg.Schedule.DayEnd, time.Now, 60)
			m.slotState = NewSlotStateManager(slotConfig)

			// Setup slotState with the grid and start move
			slotGrid := WeekWindowToSlotGrid(ww, m.slotState.Config())
			m.slotState.SetGrid(slotGrid)
			m.slotState.EnterEditMode()

			// Find the task in the grid (it should be there)
			foundTask := m.slotState.TaskAt(7, slotConfig.MinutesToSlot(task.TimeToMinutes(tt.taskStart))) // Day 7 = current week Monday
			if foundTask == nil {
				t.Fatalf("Task not found in grid at day 7, slot %d", slotConfig.MinutesToSlot(task.TimeToMinutes(tt.taskStart)))
			}
			_ = m.slotState.StartMove(foundTask)

			// Test
			got := m.movingTaskSlotCount()
			if got != tt.expectedSlots {
				t.Errorf("movingTaskSlotCount() = %d, want %d", got, tt.expectedSlots)
			}
		})
	}
}

func TestMovingTaskSlotCount_NotMoving(t *testing.T) {
	cfg := &config.Config{
		Schedule: config.ScheduleConfig{
			DayStart: "09:00",
			DayEnd:   "17:00",
		},
	}

	m := New(nil, cfg)
	m.rowHeight = 15

	// Not in move mode - should return 0
	got := m.movingTaskSlotCount()
	if got != 0 {
		t.Errorf("movingTaskSlotCount() when not moving = %d, want 0", got)
	}
}

func TestMaxSlots(t *testing.T) {
	tests := []struct {
		name      string
		dayStart  string
		dayEnd    string
		rowHeight int
		expected  int
	}{
		{name: "8h day with 30min slots", dayStart: "09:00", dayEnd: "17:00", rowHeight: 30, expected: 16},
		{name: "8h day with 15min slots", dayStart: "09:00", dayEnd: "17:00", rowHeight: 15, expected: 32},
		{name: "8h day with 60min slots", dayStart: "09:00", dayEnd: "17:00", rowHeight: 60, expected: 8},
		{name: "10h day with 30min slots", dayStart: "08:00", dayEnd: "18:00", rowHeight: 30, expected: 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Schedule: config.ScheduleConfig{
					DayStart: tt.dayStart,
					DayEnd:   tt.dayEnd,
				},
			}
			m := New(nil, cfg)
			m.rowHeight = tt.rowHeight

			got := m.maxSlots()
			if got != tt.expected {
				t.Errorf("maxSlots() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestCalculateColWidth(t *testing.T) {
	cfg := &config.Config{
		Schedule: config.ScheduleConfig{
			DayStart: "09:00",
			DayEnd:   "17:00",
		},
	}

	tests := []struct {
		name     string
		width    int
		expected int
	}{
		{name: "narrow terminal", width: 80, expected: 10},     // (80-22)/7 = 8, clamped to 10
		{name: "medium terminal", width: 120, expected: 14},    // (120-22)/7 = 14
		{name: "wide terminal", width: 180, expected: 22},      // (180-22)/7 = 22
		{name: "very wide terminal", width: 250, expected: 32}, // (250-22)/7 = 32
		{name: "zero width", width: 0, expected: defaultColWidth},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(nil, cfg)
			m.width = tt.width

			got := m.calculateColWidth()
			if got != tt.expected {
				t.Errorf("calculateColWidth() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestCalculateLayout(t *testing.T) {
	cfg := &config.Config{
		Schedule: config.ScheduleConfig{
			DayStart: "09:00",
			DayEnd:   "17:00", // 8 hours = 8/16/32 slots at 60/30/15min
		},
	}

	tests := []struct {
		name           string
		height         int
		expectedHeight int // rowHeight in minutes
		expectedLines  int // rowLines
	}{
		{name: "very tall - 15min slots", height: 100, expectedHeight: 15, expectedLines: 3},
		{name: "tall - 15min slots", height: 60, expectedHeight: 15, expectedLines: 2},
		{name: "medium - 15min slots", height: 36, expectedHeight: 15, expectedLines: 1},
		{name: "small - 15min slots", height: 28, expectedHeight: 15, expectedLines: 1},
		{name: "very small - 15min slots", height: 20, expectedHeight: 15, expectedLines: 1},
		{name: "zero height", height: 0, expectedHeight: 15, expectedLines: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(nil, cfg)
			m.height = tt.height

			m.calculateLayout()

			if m.rowHeight != tt.expectedHeight {
				t.Errorf("rowHeight = %d, want %d", m.rowHeight, tt.expectedHeight)
			}
			if m.rowLines != tt.expectedLines {
				t.Errorf("rowLines = %d, want %d", m.rowLines, tt.expectedLines)
			}
		})
	}
}

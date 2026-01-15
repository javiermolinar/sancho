// Package tui provides the terminal user interface for sancho.
package tui

import (
	"testing"
	"time"

	"github.com/javiermolinar/sancho/internal/config"
	"github.com/javiermolinar/sancho/internal/task"
	"github.com/javiermolinar/sancho/internal/tui/commands"
)

func TestInitialLoadFocusesCursorOnCurrentTaskOrTime(t *testing.T) {
	cfg := &config.Config{
		Schedule: config.ScheduleConfig{
			DayStart: "09:00",
			DayEnd:   "17:00",
		},
	}

	now := time.Date(2025, 1, 6, 10, 30, 0, 0, time.Local) // Monday
	dayStartMins := task.TimeToMinutes(cfg.Schedule.DayStart)
	expectedSlot := (now.Hour()*60 + now.Minute() - dayStartMins) / 15

	tests := []struct {
		name     string
		withTask bool
	}{
		{name: "current_task", withTask: true},
		{name: "no_task", withTask: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(nil, cfg)
			m.rowHeight = 15
			m.slotState.UpdateConfig(SlotGridConfigFromWeekWindow(nil, cfg.Schedule.DayStart, cfg.Schedule.DayEnd, func() time.Time {
				return now
			}, m.rowHeight))

			weekStart := startOfWeek(now)
			week := task.NewWeek(weekStart)
			if tt.withTask {
				currentTask := &task.Task{
					ID:             1,
					Description:    "Current task",
					Category:       task.CategoryDeep,
					ScheduledDate:  weekStart,
					ScheduledStart: "10:00",
					ScheduledEnd:   "11:00",
					Status:         task.StatusScheduled,
				}
				_ = week.Day(0).AddTask(currentTask)
			}

			ww := task.NewWeekWindow(nil, week, nil)
			updated, _ := m.Update(commands.InitialLoadMsg{Window: ww})
			model := updated.(Model)

			if model.cursor.Day != 0 {
				t.Fatalf("cursor day = %d, want 0", model.cursor.Day)
			}
			if model.cursor.Slot != expectedSlot {
				t.Fatalf("cursor slot = %d, want %d", model.cursor.Slot, expectedSlot)
			}
		})
	}
}

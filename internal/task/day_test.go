package task

import (
	"errors"
	"testing"
	"time"
)

func TestNewDay(t *testing.T) {
	date := time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)
	day := NewDay(date)

	if day.Len() != 0 {
		t.Errorf("expected empty day, got %d tasks", day.Len())
	}

	// Date should be truncated to midnight
	expected := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	if !day.Date.Equal(expected) {
		t.Errorf("expected date %v, got %v", expected, day.Date)
	}
}

func TestDay_AddTask(t *testing.T) {
	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	t.Run("add single task", func(t *testing.T) {
		day := NewDay(date)
		task := &Task{
			Description:    "Test task",
			ScheduledStart: "09:00",
			ScheduledEnd:   "10:00",
			Status:         StatusScheduled,
		}

		err := day.AddTask(task)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if day.Len() != 1 {
			t.Errorf("expected 1 task, got %d", day.Len())
		}
	})

	t.Run("add nil task", func(t *testing.T) {
		day := NewDay(date)
		err := day.AddTask(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if day.Len() != 0 {
			t.Errorf("expected 0 tasks, got %d", day.Len())
		}
	})

	t.Run("tasks sorted by start time", func(t *testing.T) {
		day := NewDay(date)

		// Add tasks out of order
		_ = day.AddTask(&Task{Description: "Second", ScheduledStart: "11:00", ScheduledEnd: "12:00", Status: StatusScheduled})
		_ = day.AddTask(&Task{Description: "First", ScheduledStart: "09:00", ScheduledEnd: "10:00", Status: StatusScheduled})
		_ = day.AddTask(&Task{Description: "Third", ScheduledStart: "14:00", ScheduledEnd: "15:00", Status: StatusScheduled})

		tasks := day.Tasks()
		if len(tasks) != 3 {
			t.Fatalf("expected 3 tasks, got %d", len(tasks))
		}
		if tasks[0].Description != "First" {
			t.Errorf("expected First, got %s", tasks[0].Description)
		}
		if tasks[1].Description != "Second" {
			t.Errorf("expected Second, got %s", tasks[1].Description)
		}
		if tasks[2].Description != "Third" {
			t.Errorf("expected Third, got %s", tasks[2].Description)
		}
	})

	t.Run("overlap with scheduled task", func(t *testing.T) {
		day := NewDay(date)
		_ = day.AddTask(&Task{Description: "Existing", ScheduledStart: "09:00", ScheduledEnd: "10:30", Status: StatusScheduled})

		err := day.AddTask(&Task{Description: "Overlapping", ScheduledStart: "10:00", ScheduledEnd: "11:00", Status: StatusScheduled})
		if err == nil {
			t.Fatal("expected overlap error, got nil")
		}
		if !errors.Is(err, ErrTimeBlockOverlap) {
			t.Errorf("expected ErrTimeBlockOverlap, got %v", err)
		}
	})

	t.Run("no overlap with cancelled task", func(t *testing.T) {
		day := NewDay(date)
		_ = day.AddTask(&Task{Description: "Cancelled", ScheduledStart: "09:00", ScheduledEnd: "10:30", Status: StatusCancelled})

		err := day.AddTask(&Task{Description: "New", ScheduledStart: "10:00", ScheduledEnd: "11:00", Status: StatusScheduled})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("no overlap with postponed task", func(t *testing.T) {
		day := NewDay(date)
		_ = day.AddTask(&Task{Description: "Postponed", ScheduledStart: "09:00", ScheduledEnd: "10:30", Status: StatusPostponed})

		err := day.AddTask(&Task{Description: "New", ScheduledStart: "10:00", ScheduledEnd: "11:00", Status: StatusScheduled})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestNewDayWithTasks(t *testing.T) {
	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	t.Run("valid tasks", func(t *testing.T) {
		tasks := []*Task{
			{Description: "First", ScheduledStart: "09:00", ScheduledEnd: "10:00", Status: StatusScheduled},
			{Description: "Second", ScheduledStart: "11:00", ScheduledEnd: "12:00", Status: StatusScheduled},
		}

		day, err := NewDayWithTasks(date, tasks)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if day.Len() != 2 {
			t.Errorf("expected 2 tasks, got %d", day.Len())
		}
	})

	t.Run("overlapping tasks", func(t *testing.T) {
		tasks := []*Task{
			{Description: "First", ScheduledStart: "09:00", ScheduledEnd: "10:30", Status: StatusScheduled},
			{Description: "Second", ScheduledStart: "10:00", ScheduledEnd: "11:00", Status: StatusScheduled},
		}

		_, err := NewDayWithTasks(date, tasks)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrTimeBlockOverlap) {
			t.Errorf("expected ErrTimeBlockOverlap, got %v", err)
		}
	})
}

func TestDay_FindOverlappingTask(t *testing.T) {
	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	day := NewDay(date)
	existing := &Task{Description: "Existing", ScheduledStart: "09:00", ScheduledEnd: "10:30", Status: StatusScheduled}
	_ = day.AddTask(existing)

	tests := []struct {
		name      string
		start     string
		end       string
		wantFound bool
	}{
		{name: "no overlap before", start: "08:00", end: "09:00", wantFound: false},
		{name: "no overlap after", start: "10:30", end: "11:00", wantFound: false},
		{name: "overlap at start", start: "08:30", end: "09:30", wantFound: true},
		{name: "overlap at end", start: "10:00", end: "11:00", wantFound: true},
		{name: "fully inside", start: "09:30", end: "10:00", wantFound: true},
		{name: "fully contains", start: "08:00", end: "11:00", wantFound: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found := day.FindOverlappingTask(tt.start, tt.end)
			if tt.wantFound && found == nil {
				t.Error("expected to find overlapping task, got nil")
			}
			if !tt.wantFound && found != nil {
				t.Errorf("expected no overlap, found %q", found.Description)
			}
		})
	}
}

func TestDay_HasOverlap(t *testing.T) {
	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	day := NewDay(date)
	_ = day.AddTask(&Task{Description: "Existing", ScheduledStart: "09:00", ScheduledEnd: "10:30", Status: StatusScheduled})

	if !day.HasOverlap("09:30", "10:00") {
		t.Error("expected overlap")
	}
	if day.HasOverlap("11:00", "12:00") {
		t.Error("expected no overlap")
	}
}

func TestDay_ScheduledTasks(t *testing.T) {
	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	day := NewDay(date)

	_ = day.AddTask(&Task{Description: "Scheduled", ScheduledStart: "09:00", ScheduledEnd: "10:00", Status: StatusScheduled})
	_ = day.AddTask(&Task{Description: "Cancelled", ScheduledStart: "11:00", ScheduledEnd: "12:00", Status: StatusCancelled})
	_ = day.AddTask(&Task{Description: "Postponed", ScheduledStart: "13:00", ScheduledEnd: "14:00", Status: StatusPostponed})

	scheduled := day.ScheduledTasks()
	if len(scheduled) != 1 {
		t.Errorf("expected 1 scheduled task, got %d", len(scheduled))
	}
	if scheduled[0].Description != "Scheduled" {
		t.Errorf("expected 'Scheduled', got %q", scheduled[0].Description)
	}
}

func TestDay_Stats(t *testing.T) {
	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	day := NewDay(date)

	_ = day.AddTask(&Task{
		Description:    "Deep work",
		Category:       CategoryDeep,
		ScheduledStart: "09:00",
		ScheduledEnd:   "11:00",
		Status:         StatusScheduled,
	})
	_ = day.AddTask(&Task{
		Description:    "Shallow work",
		Category:       CategoryShallow,
		ScheduledStart: "11:00",
		ScheduledEnd:   "12:00",
		Status:         StatusScheduled,
	})
	_ = day.AddTask(&Task{
		Description:    "Cancelled",
		Category:       CategoryDeep,
		ScheduledStart: "14:00",
		ScheduledEnd:   "15:00",
		Status:         StatusCancelled,
	})
	_ = day.AddTask(&Task{
		Description:    "Postponed",
		Category:       CategoryShallow,
		ScheduledStart: "16:00",
		ScheduledEnd:   "17:00",
		Status:         StatusPostponed,
	})

	stats := day.Stats()

	if stats.TotalBlocks != 4 {
		t.Errorf("expected 4 total blocks, got %d", stats.TotalBlocks)
	}
	if stats.DeepMinutes != 120 {
		t.Errorf("expected 120 deep minutes, got %d", stats.DeepMinutes)
	}
	if stats.ShallowMinutes != 60 {
		t.Errorf("expected 60 shallow minutes, got %d", stats.ShallowMinutes)
	}
	if stats.CancelledBlocks != 1 {
		t.Errorf("expected 1 cancelled block, got %d", stats.CancelledBlocks)
	}
	if stats.PostponedBlocks != 1 {
		t.Errorf("expected 1 postponed block, got %d", stats.PostponedBlocks)
	}
	if stats.TotalMinutes() != 180 {
		t.Errorf("expected 180 total minutes, got %d", stats.TotalMinutes())
	}
	if stats.DeepPercent() != 66 {
		t.Errorf("expected 66%% deep, got %d%%", stats.DeepPercent())
	}
}

func TestDay_StatsWithPeakHours(t *testing.T) {
	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	day := NewDay(date)

	// Deep work from 09:00-11:00, peak hours 09:00-10:00
	_ = day.AddTask(&Task{
		Description:    "Deep work",
		Category:       CategoryDeep,
		ScheduledStart: "09:00",
		ScheduledEnd:   "11:00",
		Status:         StatusScheduled,
	})

	stats := day.StatsWithPeakHours("09:00", "10:00")

	if stats.DeepMinutes != 120 {
		t.Errorf("expected 120 deep minutes, got %d", stats.DeepMinutes)
	}
	if stats.PeakDeepMinutes != 60 {
		t.Errorf("expected 60 peak deep minutes, got %d", stats.PeakDeepMinutes)
	}
}

func TestDayStats_DeepPercent(t *testing.T) {
	tests := []struct {
		name    string
		deep    int
		shallow int
		want    int
	}{
		{name: "all deep", deep: 120, shallow: 0, want: 100},
		{name: "all shallow", deep: 0, shallow: 60, want: 0},
		{name: "50/50", deep: 60, shallow: 60, want: 50},
		{name: "2:1 ratio", deep: 120, shallow: 60, want: 66},
		{name: "empty", deep: 0, shallow: 0, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := DayStats{DeepMinutes: tt.deep, ShallowMinutes: tt.shallow}
			if got := stats.DeepPercent(); got != tt.want {
				t.Errorf("DeepPercent() = %d, want %d", got, tt.want)
			}
		})
	}
}

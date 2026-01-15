package summary

import (
	"testing"
	"time"

	"github.com/javiermolinar/sancho/internal/task"
)

func TestSummarizeWeek(t *testing.T) {
	weekStart := time.Date(2025, 1, 15, 0, 0, 0, 0, time.Local) // Wednesday
	monday := time.Date(2025, 1, 13, 0, 0, 0, 0, time.Local)
	sunday := time.Date(2025, 1, 19, 0, 0, 0, 0, time.Local)

	tasks := []*task.Task{
		{
			Description:    "Deep work",
			Category:       task.CategoryDeep,
			ScheduledDate:  monday,
			ScheduledStart: "09:00",
			ScheduledEnd:   "10:00",
			Status:         task.StatusScheduled,
		},
		{
			Description:    "Shallow work",
			Category:       task.CategoryShallow,
			ScheduledDate:  monday.AddDate(0, 0, 1),
			ScheduledStart: "10:00",
			ScheduledEnd:   "10:30",
			Status:         task.StatusScheduled,
		},
		{
			Description:    "Next week",
			Category:       task.CategoryDeep,
			ScheduledDate:  sunday.AddDate(0, 0, 1),
			ScheduledStart: "09:00",
			ScheduledEnd:   "10:00",
			Status:         task.StatusScheduled,
		},
	}

	summary := SummarizeWeek(weekStart, tasks, WeekSummaryOptions{
		PeakStart: "09:00",
		PeakEnd:   "11:00",
	})

	if !summary.Start.Equal(monday) {
		t.Fatalf("start = %v, want %v", summary.Start, monday)
	}
	if !summary.End.Equal(sunday) {
		t.Fatalf("end = %v, want %v", summary.End, sunday)
	}
	if len(summary.Tasks) != 2 {
		t.Fatalf("tasks = %d, want 2", len(summary.Tasks))
	}

	stats := summary.Stats
	if stats.DeepMinutes != 60 {
		t.Fatalf("deep minutes = %d, want 60", stats.DeepMinutes)
	}
	if stats.ShallowMinutes != 30 {
		t.Fatalf("shallow minutes = %d, want 30", stats.ShallowMinutes)
	}
	if stats.TotalBlocks != 2 {
		t.Fatalf("total blocks = %d, want 2", stats.TotalBlocks)
	}
	if stats.PeakDeepMinutes != 60 {
		t.Fatalf("peak deep minutes = %d, want 60", stats.PeakDeepMinutes)
	}
}

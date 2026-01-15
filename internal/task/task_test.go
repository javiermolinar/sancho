package task

import (
	"errors"
	"testing"
	"time"

	"github.com/javiermolinar/sancho/internal/dateutil"
)

func TestNew(t *testing.T) {
	t.Run("valid task", func(t *testing.T) {
		task, err := New("Write tests", "deep", "2025-01-15", "09:00", "11:00")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if task.Description != "Write tests" {
			t.Errorf("got description %q, want %q", task.Description, "Write tests")
		}
		if task.Category != CategoryDeep {
			t.Errorf("got category %q, want %q", task.Category, CategoryDeep)
		}
		if task.ScheduledStart != "09:00" {
			t.Errorf("got start %q, want %q", task.ScheduledStart, "09:00")
		}
		if task.ScheduledEnd != "11:00" {
			t.Errorf("got end %q, want %q", task.ScheduledEnd, "11:00")
		}
		if task.Status != StatusScheduled {
			t.Errorf("got status %q, want %q", task.Status, StatusScheduled)
		}
		if task.CreatedAt.IsZero() {
			t.Error("expected CreatedAt to be set")
		}
	})

	t.Run("empty date defaults to today", func(t *testing.T) {
		task, err := New("Write tests", "deep", "", "09:00", "11:00")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		today := dateutil.TruncateToDay(time.Now())
		if !task.ScheduledDate.Equal(today) {
			t.Errorf("got date %v, want %v", task.ScheduledDate, today)
		}
	})

	t.Run("shallow category", func(t *testing.T) {
		task, err := New("Review PRs", "shallow", "", "14:00", "15:00")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if task.Category != CategoryShallow {
			t.Errorf("got category %q, want %q", task.Category, CategoryShallow)
		}
	})
}

func TestNew_Errors(t *testing.T) {
	tests := []struct {
		name        string
		description string
		category    string
		date        string
		start       string
		end         string
		wantErr     error
	}{
		{
			name:        "empty description",
			description: "",
			category:    "deep",
			date:        "",
			start:       "09:00",
			end:         "11:00",
			wantErr:     ErrEmptyDescription,
		},
		{
			name:        "invalid category",
			description: "Test",
			category:    "invalid",
			date:        "",
			start:       "09:00",
			end:         "11:00",
			wantErr:     ErrInvalidCategory,
		},
		{
			name:        "invalid date format",
			description: "Test",
			category:    "deep",
			date:        "01-15-2025",
			start:       "09:00",
			end:         "11:00",
			wantErr:     dateutil.ErrInvalidDateFormat,
		},
		{
			name:        "invalid start time",
			description: "Test",
			category:    "deep",
			date:        "",
			start:       "9:00",
			end:         "11:00",
			wantErr:     ErrInvalidTimeFormat,
		},
		{
			name:        "invalid end time",
			description: "Test",
			category:    "deep",
			date:        "",
			start:       "09:00",
			end:         "25:00",
			wantErr:     ErrInvalidTimeFormat,
		},
		{
			name:        "end before start",
			description: "Test",
			category:    "deep",
			date:        "",
			start:       "14:00",
			end:         "12:00",
			wantErr:     ErrEndBeforeStart,
		},
		{
			name:        "end equals start",
			description: "Test",
			category:    "deep",
			date:        "",
			start:       "14:00",
			end:         "14:00",
			wantErr:     ErrEndBeforeStart,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.description, tt.category, tt.date, tt.start, tt.end)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestTask_StatusChecks(t *testing.T) {
	tests := []struct {
		name        string
		status      Status
		isScheduled bool
		isCancelled bool
		isPostponed bool
	}{
		{
			name:        "scheduled",
			status:      StatusScheduled,
			isScheduled: true,
		},
		{
			name:        "cancelled",
			status:      StatusCancelled,
			isCancelled: true,
		},
		{
			name:        "postponed",
			status:      StatusPostponed,
			isPostponed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Status: tt.status}
			if task.IsScheduled() != tt.isScheduled {
				t.Errorf("IsScheduled() = %v, want %v", task.IsScheduled(), tt.isScheduled)
			}
			if task.IsCancelled() != tt.isCancelled {
				t.Errorf("IsCancelled() = %v, want %v", task.IsCancelled(), tt.isCancelled)
			}
			if task.IsPostponed() != tt.isPostponed {
				t.Errorf("IsPostponed() = %v, want %v", task.IsPostponed(), tt.isPostponed)
			}
		})
	}
}

func TestTask_CategoryChecks(t *testing.T) {
	tests := []struct {
		name      string
		category  Category
		isDeep    bool
		isShallow bool
	}{
		{
			name:     "deep",
			category: CategoryDeep,
			isDeep:   true,
		},
		{
			name:      "shallow",
			category:  CategoryShallow,
			isShallow: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Category: tt.category}
			if task.IsDeep() != tt.isDeep {
				t.Errorf("IsDeep() = %v, want %v", task.IsDeep(), tt.isDeep)
			}
			if task.IsShallow() != tt.isShallow {
				t.Errorf("IsShallow() = %v, want %v", task.IsShallow(), tt.isShallow)
			}
		})
	}
}

func TestTask_Duration(t *testing.T) {
	tests := []struct {
		name  string
		start string
		end   string
		want  int
	}{
		{name: "1 hour", start: "09:00", end: "10:00", want: 60},
		{name: "2 hours", start: "09:00", end: "11:00", want: 120},
		{name: "30 minutes", start: "09:00", end: "09:30", want: 30},
		{name: "2.5 hours", start: "09:00", end: "11:30", want: 150},
		{name: "invalid start", start: "invalid", end: "10:00", want: 0},
		{name: "invalid end", start: "09:00", end: "bad", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{ScheduledStart: tt.start, ScheduledEnd: tt.end}
			got := task.Duration()
			if got != tt.want {
				t.Errorf("Duration() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestTask_OverlapsWith(t *testing.T) {
	baseDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	otherDate := time.Date(2025, 1, 16, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name  string
		task1 *Task
		task2 *Task
		want  bool
	}{
		{
			name:  "nil other",
			task1: &Task{ScheduledDate: baseDate, ScheduledStart: "09:00", ScheduledEnd: "10:00"},
			task2: nil,
			want:  false,
		},
		{
			name:  "different days",
			task1: &Task{ScheduledDate: baseDate, ScheduledStart: "09:00", ScheduledEnd: "10:00"},
			task2: &Task{ScheduledDate: otherDate, ScheduledStart: "09:00", ScheduledEnd: "10:00"},
			want:  false,
		},
		{
			name:  "same day no overlap",
			task1: &Task{ScheduledDate: baseDate, ScheduledStart: "09:00", ScheduledEnd: "10:00"},
			task2: &Task{ScheduledDate: baseDate, ScheduledStart: "10:00", ScheduledEnd: "11:00"},
			want:  false,
		},
		{
			name:  "same day with overlap",
			task1: &Task{ScheduledDate: baseDate, ScheduledStart: "09:00", ScheduledEnd: "10:30"},
			task2: &Task{ScheduledDate: baseDate, ScheduledStart: "10:00", ScheduledEnd: "11:00"},
			want:  true,
		},
		{
			name:  "same time slot",
			task1: &Task{ScheduledDate: baseDate, ScheduledStart: "09:00", ScheduledEnd: "11:00"},
			task2: &Task{ScheduledDate: baseDate, ScheduledStart: "09:00", ScheduledEnd: "11:00"},
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.task1.OverlapsWith(tt.task2)
			if got != tt.want {
				t.Errorf("OverlapsWith() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTask_IsPast(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := today.AddDate(0, 0, -1)
	tomorrow := today.AddDate(0, 0, 1)

	tests := []struct {
		name string
		task *Task
		want bool
	}{
		{
			name: "yesterday task is past",
			task: &Task{ScheduledDate: yesterday, ScheduledEnd: "23:59"},
			want: true,
		},
		{
			name: "tomorrow task is not past",
			task: &Task{ScheduledDate: tomorrow, ScheduledEnd: "09:00"},
			want: false,
		},
		{
			name: "today task ended earlier is past",
			task: &Task{ScheduledDate: today, ScheduledEnd: "00:01"},
			want: true,
		},
		{
			name: "today task ending later is not past",
			task: &Task{ScheduledDate: today, ScheduledEnd: "23:59"},
			want: false,
		},
		{
			name: "invalid end time returns false",
			task: &Task{ScheduledDate: yesterday, ScheduledEnd: "invalid"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.task.IsPast()
			if got != tt.want {
				t.Errorf("IsPast() = %v, want %v", got, tt.want)
			}
		})
	}
}

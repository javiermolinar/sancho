package dwplanner

import (
	"testing"
	"time"

	"github.com/javiermolinar/sancho/internal/llm"
	"github.com/javiermolinar/sancho/internal/task"
)

func TestValidator_DateFormat(t *testing.T) {
	now := time.Date(2025, 1, 13, 10, 0, 0, 0, time.Local)
	v := NewValidator(now, "09:00", "17:00", nil)

	tests := []struct {
		name      string
		date      string
		wantValid bool
	}{
		{name: "valid date", date: "2025-01-13", wantValid: true},
		{name: "valid future date", date: "2025-01-20", wantValid: true},
		{name: "invalid format slashes", date: "01/13/2025", wantValid: false},
		{name: "invalid format dashes wrong order", date: "13-01-2025", wantValid: false},
		{name: "invalid month", date: "2025-13-01", wantValid: false},
		{name: "invalid day", date: "2025-01-32", wantValid: false},
		{name: "empty string", date: "", wantValid: false},
		{name: "random text", date: "next monday", wantValid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks := []llm.PlannedTask{{
				Description:    "Test task",
				Category:       "deep",
				ScheduledDate:  tt.date,
				ScheduledStart: "11:00",
				ScheduledEnd:   "12:00",
			}}

			result := v.Validate(tasks)
			if result.Valid != tt.wantValid {
				t.Errorf("Validate() valid = %v, want %v", result.Valid, tt.wantValid)
				if !result.Valid {
					for _, e := range result.Errors {
						t.Logf("Error: %s", e)
					}
				}
			}
		})
	}
}

func TestValidator_TimeFormat(t *testing.T) {
	now := time.Date(2025, 1, 13, 10, 0, 0, 0, time.Local)
	v := NewValidator(now, "09:00", "17:00", nil)

	tests := []struct {
		name      string
		start     string
		end       string
		wantValid bool
		wantField string
	}{
		{name: "valid times", start: "11:00", end: "12:00", wantValid: true},
		{name: "valid early morning", start: "00:00", end: "01:00", wantValid: true},
		{name: "valid late night", start: "22:00", end: "23:59", wantValid: true},
		{name: "invalid start format", start: "9:00", end: "10:00", wantValid: false, wantField: "scheduled_start"},
		{name: "invalid end format", start: "09:00", end: "10", wantValid: false, wantField: "scheduled_end"},
		{name: "invalid hour 25", start: "25:00", end: "26:00", wantValid: false, wantField: "scheduled_start"},
		{name: "invalid minute 60", start: "09:60", end: "10:00", wantValid: false, wantField: "scheduled_start"},
		{name: "empty start", start: "", end: "10:00", wantValid: false, wantField: "scheduled_start"},
		{name: "empty end", start: "09:00", end: "", wantValid: false, wantField: "scheduled_end"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks := []llm.PlannedTask{{
				Description:    "Test task",
				Category:       "deep",
				ScheduledDate:  "2025-01-14", // Future date to avoid past-time errors
				ScheduledStart: tt.start,
				ScheduledEnd:   tt.end,
			}}

			result := v.Validate(tasks)
			if result.Valid != tt.wantValid {
				t.Errorf("Validate() valid = %v, want %v", result.Valid, tt.wantValid)
			}

			if !tt.wantValid && tt.wantField != "" {
				found := false
				for _, e := range result.Errors {
					if e.Field == tt.wantField {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error on field %q, got errors: %v", tt.wantField, result.Errors)
				}
			}
		})
	}
}

func TestValidator_EndBeforeStart(t *testing.T) {
	now := time.Date(2025, 1, 13, 10, 0, 0, 0, time.Local)
	v := NewValidator(now, "09:00", "17:00", nil)

	tests := []struct {
		name      string
		start     string
		end       string
		wantValid bool
	}{
		{name: "end after start", start: "09:00", end: "10:00", wantValid: true},
		{name: "end equals start", start: "09:00", end: "09:00", wantValid: false},
		{name: "end before start", start: "10:00", end: "09:00", wantValid: false},
		{name: "one minute difference", start: "09:00", end: "09:01", wantValid: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks := []llm.PlannedTask{{
				Description:    "Test task",
				Category:       "deep",
				ScheduledDate:  "2025-01-14",
				ScheduledStart: tt.start,
				ScheduledEnd:   tt.end,
			}}

			result := v.Validate(tasks)
			if result.Valid != tt.wantValid {
				t.Errorf("Validate() valid = %v, want %v", result.Valid, tt.wantValid)
			}
		})
	}
}

func TestValidator_PastTime(t *testing.T) {
	// Current time is 10:30 on 2025-01-13
	now := time.Date(2025, 1, 13, 10, 30, 0, 0, time.Local)
	v := NewValidator(now, "09:00", "17:00", nil)

	tests := []struct {
		name      string
		date      string
		start     string
		wantValid bool
	}{
		{name: "today future time", date: "2025-01-13", start: "11:00", wantValid: true},
		{name: "today current time", date: "2025-01-13", start: "10:30", wantValid: true}, // Edge case: exact match is valid
		{name: "today past time", date: "2025-01-13", start: "09:00", wantValid: false},
		{name: "today just past", date: "2025-01-13", start: "10:29", wantValid: false},
		{name: "tomorrow past time ok", date: "2025-01-14", start: "09:00", wantValid: true},
		{name: "future date past time ok", date: "2025-01-20", start: "08:00", wantValid: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks := []llm.PlannedTask{{
				Description:    "Test task",
				Category:       "deep",
				ScheduledDate:  tt.date,
				ScheduledStart: tt.start,
				ScheduledEnd:   "17:00",
			}}

			result := v.Validate(tasks)
			if result.Valid != tt.wantValid {
				t.Errorf("Validate() valid = %v, want %v", result.Valid, tt.wantValid)
				for _, e := range result.Errors {
					t.Logf("Error: %s", e)
				}
			}
		})
	}
}

func TestValidator_SelfOverlap(t *testing.T) {
	now := time.Date(2025, 1, 13, 8, 0, 0, 0, time.Local)
	v := NewValidator(now, "09:00", "17:00", nil)

	tests := []struct {
		name      string
		tasks     []llm.PlannedTask
		wantValid bool
	}{
		{
			name: "no overlap",
			tasks: []llm.PlannedTask{
				{Description: "Task 1", Category: "deep", ScheduledDate: "2025-01-13", ScheduledStart: "09:00", ScheduledEnd: "10:00"},
				{Description: "Task 2", Category: "deep", ScheduledDate: "2025-01-13", ScheduledStart: "10:00", ScheduledEnd: "11:00"},
			},
			wantValid: true,
		},
		{
			name: "adjacent tasks ok",
			tasks: []llm.PlannedTask{
				{Description: "Task 1", Category: "deep", ScheduledDate: "2025-01-13", ScheduledStart: "09:00", ScheduledEnd: "10:00"},
				{Description: "Task 2", Category: "deep", ScheduledDate: "2025-01-13", ScheduledStart: "10:00", ScheduledEnd: "11:00"},
				{Description: "Task 3", Category: "deep", ScheduledDate: "2025-01-13", ScheduledStart: "11:00", ScheduledEnd: "12:00"},
			},
			wantValid: true,
		},
		{
			name: "overlap same time",
			tasks: []llm.PlannedTask{
				{Description: "Task 1", Category: "deep", ScheduledDate: "2025-01-13", ScheduledStart: "09:00", ScheduledEnd: "10:00"},
				{Description: "Task 2", Category: "deep", ScheduledDate: "2025-01-13", ScheduledStart: "09:00", ScheduledEnd: "10:00"},
			},
			wantValid: false,
		},
		{
			name: "partial overlap",
			tasks: []llm.PlannedTask{
				{Description: "Task 1", Category: "deep", ScheduledDate: "2025-01-13", ScheduledStart: "09:00", ScheduledEnd: "10:30"},
				{Description: "Task 2", Category: "deep", ScheduledDate: "2025-01-13", ScheduledStart: "10:00", ScheduledEnd: "11:00"},
			},
			wantValid: false,
		},
		{
			name: "contained task",
			tasks: []llm.PlannedTask{
				{Description: "Task 1", Category: "deep", ScheduledDate: "2025-01-13", ScheduledStart: "09:00", ScheduledEnd: "12:00"},
				{Description: "Task 2", Category: "deep", ScheduledDate: "2025-01-13", ScheduledStart: "10:00", ScheduledEnd: "11:00"},
			},
			wantValid: false,
		},
		{
			name: "different days no overlap",
			tasks: []llm.PlannedTask{
				{Description: "Task 1", Category: "deep", ScheduledDate: "2025-01-13", ScheduledStart: "09:00", ScheduledEnd: "10:00"},
				{Description: "Task 2", Category: "deep", ScheduledDate: "2025-01-14", ScheduledStart: "09:00", ScheduledEnd: "10:00"},
			},
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.Validate(tt.tasks)
			if result.Valid != tt.wantValid {
				t.Errorf("Validate() valid = %v, want %v", result.Valid, tt.wantValid)
				for _, e := range result.Errors {
					t.Logf("Error: %s", e)
				}
			}
		})
	}
}

func TestValidator_ExistingOverlap(t *testing.T) {
	now := time.Date(2025, 1, 13, 8, 0, 0, 0, time.Local)

	existingTasks := []*task.Task{
		{
			Description:    "Existing meeting",
			Category:       task.CategoryShallow,
			ScheduledDate:  time.Date(2025, 1, 13, 0, 0, 0, 0, time.Local),
			ScheduledStart: "10:00",
			ScheduledEnd:   "11:00",
			Status:         task.StatusScheduled,
		},
		{
			Description:    "Cancelled task",
			Category:       task.CategoryDeep,
			ScheduledDate:  time.Date(2025, 1, 13, 0, 0, 0, 0, time.Local),
			ScheduledStart: "14:00",
			ScheduledEnd:   "15:00",
			Status:         task.StatusCancelled,
		},
	}

	v := NewValidator(now, "09:00", "17:00", existingTasks)

	tests := []struct {
		name      string
		tasks     []llm.PlannedTask
		wantValid bool
	}{
		{
			name: "no overlap with existing",
			tasks: []llm.PlannedTask{
				{Description: "New task", Category: "deep", ScheduledDate: "2025-01-13", ScheduledStart: "09:00", ScheduledEnd: "10:00"},
			},
			wantValid: true,
		},
		{
			name: "overlap with existing meeting",
			tasks: []llm.PlannedTask{
				{Description: "New task", Category: "deep", ScheduledDate: "2025-01-13", ScheduledStart: "10:30", ScheduledEnd: "11:30"},
			},
			wantValid: false,
		},
		{
			name: "can schedule over cancelled task",
			tasks: []llm.PlannedTask{
				{Description: "New task", Category: "deep", ScheduledDate: "2025-01-13", ScheduledStart: "14:00", ScheduledEnd: "15:00"},
			},
			wantValid: true,
		},
		{
			name: "different day no overlap",
			tasks: []llm.PlannedTask{
				{Description: "New task", Category: "deep", ScheduledDate: "2025-01-14", ScheduledStart: "10:00", ScheduledEnd: "11:00"},
			},
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.Validate(tt.tasks)
			if result.Valid != tt.wantValid {
				t.Errorf("Validate() valid = %v, want %v", result.Valid, tt.wantValid)
				for _, e := range result.Errors {
					t.Logf("Error: %s", e)
				}
			}
		})
	}
}

func TestValidator_MultipleErrors(t *testing.T) {
	now := time.Date(2025, 1, 13, 10, 0, 0, 0, time.Local)
	v := NewValidator(now, "09:00", "17:00", nil)

	tasks := []llm.PlannedTask{
		{
			Description:    "Invalid task",
			Category:       "deep",
			ScheduledDate:  "invalid-date",
			ScheduledStart: "bad",
			ScheduledEnd:   "also-bad",
		},
	}

	result := v.Validate(tasks)
	if result.Valid {
		t.Error("Expected validation to fail")
	}

	// Should have at least 3 errors (date, start, end)
	if len(result.Errors) < 3 {
		t.Errorf("Expected at least 3 errors, got %d: %v", len(result.Errors), result.Errors)
	}
}

func TestValidator_Valid(t *testing.T) {
	now := time.Date(2025, 1, 13, 8, 0, 0, 0, time.Local)
	v := NewValidator(now, "09:00", "17:00", nil)

	tasks := []llm.PlannedTask{
		{
			Description:    "Write thesis introduction",
			Category:       "deep",
			ScheduledDate:  "2025-01-13",
			ScheduledStart: "09:00",
			ScheduledEnd:   "11:00",
		},
		{
			Description:    "Review PRs",
			Category:       "shallow",
			ScheduledDate:  "2025-01-13",
			ScheduledStart: "11:00",
			ScheduledEnd:   "12:00",
		},
		{
			Description:    "Email clients",
			Category:       "shallow",
			ScheduledDate:  "2025-01-14",
			ScheduledStart: "09:00",
			ScheduledEnd:   "10:00",
		},
	}

	result := v.Validate(tasks)
	if !result.Valid {
		t.Errorf("Expected validation to pass, got errors: %v", result.Errors)
	}
}

func TestValidationResult_FormatErrors(t *testing.T) {
	result := ValidationResult{
		Valid: false,
		Errors: []ValidationError{
			{TaskIndex: 0, Field: "scheduled_start", Message: "'25:00' is invalid"},
			{TaskIndex: 1, Field: "overlap", Message: "overlaps with existing task"},
		},
	}

	formatted := result.FormatErrors()
	if formatted == "" {
		t.Error("Expected formatted errors, got empty string")
	}

	// Check that it contains key parts
	if !contains(formatted, "Task 0") || !contains(formatted, "Task 1") {
		t.Errorf("Formatted errors missing task indices: %s", formatted)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Package dwplanner provides high-level deep work planning orchestration.
package dwplanner

import (
	"fmt"
	"sort"
	"time"

	"github.com/javiermolinar/sancho/internal/llm"
	"github.com/javiermolinar/sancho/internal/task"
)

// ValidationError represents a single validation error for a planned task.
type ValidationError struct {
	TaskIndex int    // Index of the task in the input slice
	Field     string // Field name: "scheduled_date", "scheduled_start", "scheduled_end", "overlap"
	Message   string // Human-readable error message
}

// String returns a formatted error message.
func (e ValidationError) String() string {
	return fmt.Sprintf("Task %d: %s - %s", e.TaskIndex, e.Field, e.Message)
}

// ValidationResult contains the result of validating LLM-planned tasks.
type ValidationResult struct {
	Valid  bool              // True if all tasks are valid
	Errors []ValidationError // List of validation errors (empty if Valid is true)
}

// FormatErrors returns a formatted string of all validation errors for LLM feedback.
func (r ValidationResult) FormatErrors() string {
	if len(r.Errors) == 0 {
		return ""
	}

	result := "Your response had these errors:\n"
	for _, e := range r.Errors {
		result += fmt.Sprintf("- %s\n", e.String())
	}
	result += "\nPlease correct these issues and respond again with valid JSON."
	return result
}

// Validator validates LLM planning responses against scheduling constraints.
type Validator struct {
	now      time.Time    // Current time for past-time validation
	dayStart string       // Workday start time (HH:MM)
	dayEnd   string       // Workday end time (HH:MM)
	existing []*task.Task // Existing scheduled tasks to check for overlaps
}

// NewValidator creates a new Validator with the given constraints.
func NewValidator(now time.Time, dayStart, dayEnd string, existing []*task.Task) *Validator {
	return &Validator{
		now:      now,
		dayStart: dayStart,
		dayEnd:   dayEnd,
		existing: existing,
	}
}

// Validate checks the LLM-planned tasks for validity.
// It validates:
// - Date format (YYYY-MM-DD)
// - Time format (HH:MM)
// - End time > start time
// - Start time not in the past (for today's tasks)
// - No overlaps between proposed tasks
// - No overlaps with existing scheduled tasks
func (v *Validator) Validate(tasks []llm.PlannedTask) ValidationResult {
	result := ValidationResult{Valid: true}

	// First pass: validate individual task fields
	validTasks := make([]struct {
		index int
		task  llm.PlannedTask
		date  time.Time
	}, 0, len(tasks))

	for i, t := range tasks {
		taskValid := true

		// Validate date format
		date, err := time.Parse("2006-01-02", t.ScheduledDate)
		if err != nil {
			result.Errors = append(result.Errors, ValidationError{
				TaskIndex: i,
				Field:     "scheduled_date",
				Message:   fmt.Sprintf("'%s' is invalid (must be YYYY-MM-DD format)", t.ScheduledDate),
			})
			taskValid = false
		}

		// Validate start time format
		if !isValidTimeFormat(t.ScheduledStart) {
			result.Errors = append(result.Errors, ValidationError{
				TaskIndex: i,
				Field:     "scheduled_start",
				Message:   fmt.Sprintf("'%s' is invalid (must be HH:MM format, 00:00-23:59)", t.ScheduledStart),
			})
			taskValid = false
		}

		// Validate end time format
		if !isValidTimeFormat(t.ScheduledEnd) {
			result.Errors = append(result.Errors, ValidationError{
				TaskIndex: i,
				Field:     "scheduled_end",
				Message:   fmt.Sprintf("'%s' is invalid (must be HH:MM format, 00:00-23:59)", t.ScheduledEnd),
			})
			taskValid = false
		}

		// Validate end > start (only if both times are valid)
		if isValidTimeFormat(t.ScheduledStart) && isValidTimeFormat(t.ScheduledEnd) {
			if t.ScheduledEnd <= t.ScheduledStart {
				result.Errors = append(result.Errors, ValidationError{
					TaskIndex: i,
					Field:     "scheduled_end",
					Message:   fmt.Sprintf("end time '%s' must be after start time '%s'", t.ScheduledEnd, t.ScheduledStart),
				})
				taskValid = false
			}
		}

		// Validate not in the past (for today's tasks)
		if err == nil && isValidTimeFormat(t.ScheduledStart) {
			if v.isInPast(date, t.ScheduledStart) {
				result.Errors = append(result.Errors, ValidationError{
					TaskIndex: i,
					Field:     "scheduled_start",
					Message:   fmt.Sprintf("start time '%s' on '%s' is in the past", t.ScheduledStart, t.ScheduledDate),
				})
				taskValid = false
			}
		}

		// Track valid tasks for overlap checking
		if taskValid && err == nil {
			validTasks = append(validTasks, struct {
				index int
				task  llm.PlannedTask
				date  time.Time
			}{index: i, task: t, date: date})
		}
	}

	// Second pass: check for overlaps between proposed tasks
	v.checkSelfOverlaps(&result, validTasks)

	// Third pass: check for overlaps with existing tasks
	v.checkExistingOverlaps(&result, validTasks)

	result.Valid = len(result.Errors) == 0
	return result
}

// isValidTimeFormat checks if a string is in HH:MM format (00:00-23:59).
func isValidTimeFormat(s string) bool {
	if len(s) != 5 {
		return false
	}
	_, err := time.Parse("15:04", s)
	return err == nil
}

// isInPast checks if a given date and time is before the current time.
func (v *Validator) isInPast(date time.Time, timeStr string) bool {
	// Only check for today's date
	today := time.Date(v.now.Year(), v.now.Month(), v.now.Day(), 0, 0, 0, 0, v.now.Location())
	taskDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, v.now.Location())

	if !taskDate.Equal(today) {
		return false // Future dates are never in the past
	}

	// Parse the time and compare with current time
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return false
	}

	// Create full datetime for comparison
	taskTime := time.Date(v.now.Year(), v.now.Month(), v.now.Day(),
		t.Hour(), t.Minute(), 0, 0, v.now.Location())

	return taskTime.Before(v.now)
}

// checkSelfOverlaps checks for overlaps between proposed tasks.
func (v *Validator) checkSelfOverlaps(result *ValidationResult, validTasks []struct {
	index int
	task  llm.PlannedTask
	date  time.Time
}) {
	// Group tasks by date
	tasksByDate := make(map[string][]struct {
		index int
		task  llm.PlannedTask
	})

	for _, vt := range validTasks {
		dateKey := vt.date.Format("2006-01-02")
		tasksByDate[dateKey] = append(tasksByDate[dateKey], struct {
			index int
			task  llm.PlannedTask
		}{index: vt.index, task: vt.task})
	}

	// Check overlaps within each date
	for _, dayTasks := range tasksByDate {
		// Sort by start time for consistent error reporting
		sort.Slice(dayTasks, func(i, j int) bool {
			return dayTasks[i].task.ScheduledStart < dayTasks[j].task.ScheduledStart
		})

		for i := 0; i < len(dayTasks); i++ {
			for j := i + 1; j < len(dayTasks); j++ {
				t1 := dayTasks[i].task
				t2 := dayTasks[j].task

				if task.TimesOverlap(t1.ScheduledStart, t1.ScheduledEnd, t2.ScheduledStart, t2.ScheduledEnd) {
					result.Errors = append(result.Errors, ValidationError{
						TaskIndex: dayTasks[j].index,
						Field:     "overlap",
						Message: fmt.Sprintf("overlaps with task '%s' (%s-%s)",
							t1.Description, t1.ScheduledStart, t1.ScheduledEnd),
					})
				}
			}
		}
	}
}

// checkExistingOverlaps checks for overlaps between proposed tasks and existing tasks.
func (v *Validator) checkExistingOverlaps(result *ValidationResult, validTasks []struct {
	index int
	task  llm.PlannedTask
	date  time.Time
}) {
	for _, vt := range validTasks {
		taskDate := time.Date(vt.date.Year(), vt.date.Month(), vt.date.Day(), 0, 0, 0, 0, v.now.Location())

		for _, existing := range v.existing {
			// Only check scheduled tasks
			if !existing.IsScheduled() {
				continue
			}

			// Only check same date
			existingDate := time.Date(existing.ScheduledDate.Year(), existing.ScheduledDate.Month(),
				existing.ScheduledDate.Day(), 0, 0, 0, 0, v.now.Location())
			if !taskDate.Equal(existingDate) {
				continue
			}

			// Check overlap
			if task.TimesOverlap(vt.task.ScheduledStart, vt.task.ScheduledEnd,
				existing.ScheduledStart, existing.ScheduledEnd) {
				result.Errors = append(result.Errors, ValidationError{
					TaskIndex: vt.index,
					Field:     "overlap",
					Message: fmt.Sprintf("overlaps with existing task '%s' (%s-%s on %s)",
						existing.Description, existing.ScheduledStart, existing.ScheduledEnd,
						existing.ScheduledDate.Format("2006-01-02")),
				})
			}
		}
	}
}

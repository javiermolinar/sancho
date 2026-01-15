// Package task defines the core domain types for sancho.
package task

import (
	"errors"
	"fmt"
	"time"

	"github.com/javiermolinar/sancho/internal/dateutil"
)

// Validation errors.
var (
	ErrEmptyDescription  = errors.New("description cannot be empty")
	ErrInvalidCategory   = errors.New("category must be 'deep' or 'shallow'")
	ErrInvalidTimeFormat = errors.New("time must be in HH:MM format")
	ErrEndBeforeStart    = errors.New("end time must be after start time")
)

// Domain errors.
var (
	ErrTimeBlockOverlap = errors.New("time block overlaps with existing task")
	ErrCannotCancelPast = errors.New("cannot cancel past tasks")
	ErrTaskNotFound     = errors.New("task not found")
)

// Status represents the state of a task.
type Status string

const (
	StatusScheduled Status = "scheduled"
	StatusPostponed Status = "postponed"
	StatusCancelled Status = "cancelled"
)

// Category represents the type of work.
type Category string

const (
	CategoryDeep    Category = "deep"
	CategoryShallow Category = "shallow"
)

// Outcome represents how the task duration compared to the estimate.
type Outcome string

const (
	OutcomeOnTime Outcome = "on_time"
	OutcomeOver   Outcome = "over"
	OutcomeUnder  Outcome = "under"
)

// Valid returns true if the outcome is a valid value.
func (o Outcome) Valid() bool {
	switch o {
	case OutcomeOnTime, OutcomeOver, OutcomeUnder:
		return true
	default:
		return false
	}
}

// Task represents a scheduled work block.
type Task struct {
	ID             int64
	Description    string
	Category       Category
	ScheduledDate  time.Time
	ScheduledStart string // "HH:MM" format
	ScheduledEnd   string // "HH:MM" format
	Status         Status
	Outcome        *Outcome // optional, nil means assumed on_time
	PostponedFrom  *int64   // FK to original task if postponed
	CreatedAt      time.Time
}

// New creates a new Task with validation.
// date can be empty (defaults to today) or in YYYY-MM-DD format.
// category must be "deep" or "shallow".
// start and end must be in HH:MM format, with end after start.
func New(description, category, date, start, end string) (*Task, error) {
	if description == "" {
		return nil, ErrEmptyDescription
	}

	cat, err := parseCategory(category)
	if err != nil {
		return nil, err
	}

	scheduledDate, err := dateutil.ParseDate(date)
	if err != nil {
		return nil, err
	}

	if err := validateTimeFormat(start); err != nil {
		return nil, fmt.Errorf("start time: %w", err)
	}

	if err := validateTimeFormat(end); err != nil {
		return nil, fmt.Errorf("end time: %w", err)
	}

	if end <= start {
		return nil, ErrEndBeforeStart
	}

	return &Task{
		Description:    description,
		Category:       cat,
		ScheduledDate:  scheduledDate,
		ScheduledStart: start,
		ScheduledEnd:   end,
		Status:         StatusScheduled,
		CreatedAt:      time.Now(),
	}, nil
}

func parseCategory(s string) (Category, error) {
	switch s {
	case "deep":
		return CategoryDeep, nil
	case "shallow":
		return CategoryShallow, nil
	default:
		return "", ErrInvalidCategory
	}
}

func validateTimeFormat(s string) error {
	if len(s) != 5 {
		return ErrInvalidTimeFormat
	}
	if _, err := time.Parse("15:04", s); err != nil {
		return ErrInvalidTimeFormat
	}
	return nil
}

// IsScheduled returns true if the task has scheduled status.
func (t *Task) IsScheduled() bool {
	return t.Status == StatusScheduled
}

// IsCancelled returns true if the task has cancelled status.
func (t *Task) IsCancelled() bool {
	return t.Status == StatusCancelled
}

// IsPostponed returns true if the task has postponed status.
func (t *Task) IsPostponed() bool {
	return t.Status == StatusPostponed
}

// IsDeep returns true if the task is categorized as deep work.
func (t *Task) IsDeep() bool {
	return t.Category == CategoryDeep
}

// IsShallow returns true if the task is categorized as shallow work.
func (t *Task) IsShallow() bool {
	return t.Category == CategoryShallow
}

// Duration returns the task duration in minutes.
func (t *Task) Duration() int {
	start, err1 := time.Parse("15:04", t.ScheduledStart)
	end, err2 := time.Parse("15:04", t.ScheduledEnd)
	if err1 != nil || err2 != nil {
		return 0
	}
	return int(end.Sub(start).Minutes())
}

// OverlapsWith returns true if this task overlaps with another task.
// Tasks must be on the same day and have overlapping time ranges.
func (t *Task) OverlapsWith(other *Task) bool {
	if other == nil {
		return false
	}
	// Must be same day
	if !t.ScheduledDate.Equal(other.ScheduledDate) {
		return false
	}
	// Check time overlap
	return TimesOverlap(t.ScheduledStart, t.ScheduledEnd, other.ScheduledStart, other.ScheduledEnd)
}

// IsPast returns true if the task's scheduled end time has passed.
func (t *Task) IsPast() bool {
	now := time.Now()
	endTime, err := time.Parse("15:04", t.ScheduledEnd)
	if err != nil {
		return false
	}

	taskEnd := time.Date(
		t.ScheduledDate.Year(),
		t.ScheduledDate.Month(),
		t.ScheduledDate.Day(),
		endTime.Hour(),
		endTime.Minute(),
		0, 0,
		now.Location(),
	)
	return now.After(taskEnd)
}

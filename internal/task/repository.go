package task

import (
	"context"
	"time"
)

// TaskTimeUpdate represents a task time change for batch updates.
type TaskTimeUpdate struct {
	ID       int64
	NewStart string
	NewEnd   string
}

// Repository defines the storage interface for tasks.
type Repository interface {
	// CreateTask adds a new task to the repository.
	CreateTask(ctx context.Context, task *Task) error

	// GetTask retrieves a task by ID.
	GetTask(ctx context.Context, id int64) (*Task, error)

	// CancelTask marks a task as cancelled.
	CancelTask(ctx context.Context, id int64) error

	// SetTaskOutcome sets the outcome of a task during review.
	SetTaskOutcome(ctx context.Context, id int64, outcome Outcome) error

	// ListTasksByDateRange returns all tasks scheduled within the date range (inclusive).
	ListTasksByDateRange(ctx context.Context, start, end time.Time) ([]*Task, error)

	// CreateTasks adds multiple tasks in a batch.
	CreateTasks(ctx context.Context, tasks []*Task) error

	// PostponeTask atomically marks the original task as postponed and creates a new task.
	// Returns the newly created task.
	PostponeTask(ctx context.Context, taskID int64, newDate time.Time, newStart, newEnd string) (*Task, error)

	// UpdateTask updates a task's scheduled times in place.
	// Used for minor adjustments like grow/shrink operations.
	// Returns ErrTimeBlockOverlap if the new times conflict with another task.
	UpdateTask(ctx context.Context, id int64, newStart, newEnd string) error

	// UpdateTaskDescription updates a task description in place.
	// Returns ErrEmptyDescription if the description is empty.
	UpdateTaskDescription(ctx context.Context, id int64, description string) error

	// BatchUpdateTaskTimes updates multiple tasks' times atomically.
	// It validates that the final state has no overlaps before applying changes.
	// Used for move operations where multiple tasks shift positions.
	BatchUpdateTaskTimes(ctx context.Context, date time.Time, updates []TaskTimeUpdate) error

	// Close releases any resources held by the repository.
	Close() error
}

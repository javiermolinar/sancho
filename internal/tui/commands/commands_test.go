package commands

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/javiermolinar/sancho/internal/task"
)

type fakeRepo struct {
	tasksByRange func(start, end time.Time) ([]*task.Task, error)
}

func (f fakeRepo) CreateTask(ctx context.Context, task *task.Task) error {
	return errors.New("not implemented")
}

func (f fakeRepo) GetTask(ctx context.Context, id int64) (*task.Task, error) {
	return nil, errors.New("not implemented")
}

func (f fakeRepo) CancelTask(ctx context.Context, id int64) error {
	return errors.New("not implemented")
}

func (f fakeRepo) SetTaskOutcome(ctx context.Context, id int64, outcome task.Outcome) error {
	return errors.New("not implemented")
}

func (f fakeRepo) ListTasksByDateRange(ctx context.Context, start, end time.Time) ([]*task.Task, error) {
	if f.tasksByRange == nil {
		return nil, errors.New("not implemented")
	}
	return f.tasksByRange(start, end)
}

func (f fakeRepo) CreateTasks(ctx context.Context, tasks []*task.Task) error {
	return errors.New("not implemented")
}

func (f fakeRepo) PostponeTask(ctx context.Context, taskID int64, newDate time.Time, newStart, newEnd string) (*task.Task, error) {
	return nil, errors.New("not implemented")
}

func (f fakeRepo) UpdateTask(ctx context.Context, id int64, newStart, newEnd string) error {
	return errors.New("not implemented")
}

func (f fakeRepo) UpdateTaskDescription(ctx context.Context, id int64, description string) error {
	return errors.New("not implemented")
}

func (f fakeRepo) BatchUpdateTaskTimes(ctx context.Context, date time.Time, updates []task.TaskTimeUpdate) error {
	return errors.New("not implemented")
}

func (f fakeRepo) Close() error {
	return nil
}

func TestLoadWeekReturnsWeekLoadedMsg(t *testing.T) {
	weekStart := time.Date(2025, 1, 6, 0, 0, 0, 0, time.Local)
	taskDate := weekStart

	repo := fakeRepo{
		tasksByRange: func(start, end time.Time) ([]*task.Task, error) {
			return []*task.Task{
				{
					ID:             1,
					Description:    "Test",
					Category:       task.CategoryDeep,
					ScheduledDate:  taskDate,
					ScheduledStart: "09:00",
					ScheduledEnd:   "10:00",
					Status:         task.StatusScheduled,
				},
			}, nil
		},
	}

	cmd := LoadWeek(repo, weekStart)
	msg := cmd()

	loaded, ok := msg.(WeekLoadedMsg)
	if !ok {
		t.Fatalf("msg type = %T, want WeekLoadedMsg", msg)
	}

	if loaded.Week == nil {
		t.Fatal("WeekLoadedMsg.Week is nil")
	}

	day := loaded.Week.Day(0)
	if day == nil {
		t.Fatal("week day 0 is nil")
	}

	tasks := day.ScheduledTasks()
	if len(tasks) != 1 {
		t.Fatalf("scheduled tasks = %d, want 1", len(tasks))
	}
	if tasks[0].Description != "Test" {
		t.Fatalf("task description = %q, want %q", tasks[0].Description, "Test")
	}
}

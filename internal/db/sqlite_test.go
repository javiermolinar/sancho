package db

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/javiermolinar/sancho/internal/task"
)

func TestCreateTask(t *testing.T) {
	repo := newTestRepo(t)

	now := time.Now()
	date := time.Date(2025, 1, 9, 0, 0, 0, 0, time.UTC)

	tsk := &task.Task{
		Description:    "Write unit tests",
		Category:       task.CategoryDeep,
		ScheduledDate:  date,
		ScheduledStart: "09:00",
		ScheduledEnd:   "11:00",
		Status:         task.StatusScheduled,
		CreatedAt:      now,
	}

	err := repo.CreateTask(context.Background(), tsk)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	if tsk.ID == 0 {
		t.Error("expected ID to be set after insert")
	}
}

func TestCreateTask_WithOutcome(t *testing.T) {
	repo := newTestRepo(t)

	date := time.Date(2025, 1, 9, 0, 0, 0, 0, time.UTC)
	outcome := task.OutcomeOver

	tsk := &task.Task{
		Description:    "Review PRs",
		Category:       task.CategoryShallow,
		ScheduledDate:  date,
		ScheduledStart: "14:00",
		ScheduledEnd:   "15:00",
		Status:         task.StatusScheduled,
		Outcome:        &outcome,
		CreatedAt:      time.Now(),
	}

	err := repo.CreateTask(context.Background(), tsk)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	if tsk.ID == 0 {
		t.Error("expected ID to be set after insert")
	}
}

func TestCreateTask_WithPostponedFrom(t *testing.T) {
	repo := newTestRepo(t)

	date := time.Date(2025, 1, 9, 0, 0, 0, 0, time.UTC)

	// Create original task
	original := &task.Task{
		Description:    "Email clients",
		Category:       task.CategoryShallow,
		ScheduledDate:  date,
		ScheduledStart: "14:00",
		ScheduledEnd:   "14:30",
		Status:         task.StatusPostponed,
		CreatedAt:      time.Now(),
	}

	err := repo.CreateTask(context.Background(), original)
	if err != nil {
		t.Fatalf("CreateTask (original) failed: %v", err)
	}

	// Create postponed task
	nextDay := date.AddDate(0, 0, 1)
	postponed := &task.Task{
		Description:    "Email clients",
		Category:       task.CategoryShallow,
		ScheduledDate:  nextDay,
		ScheduledStart: "09:00",
		ScheduledEnd:   "09:30",
		Status:         task.StatusScheduled,
		PostponedFrom:  &original.ID,
		CreatedAt:      time.Now(),
	}

	err = repo.CreateTask(context.Background(), postponed)
	if err != nil {
		t.Fatalf("CreateTask (postponed) failed: %v", err)
	}

	if postponed.ID == 0 {
		t.Error("expected ID to be set after insert")
	}
	if *postponed.PostponedFrom != original.ID {
		t.Errorf("expected PostponedFrom to be %d, got %d", original.ID, *postponed.PostponedFrom)
	}
}

func TestGetTask(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	now := time.Now().Truncate(time.Second)

	original := &task.Task{
		Description:    "Deep work session",
		Category:       task.CategoryDeep,
		ScheduledDate:  date,
		ScheduledStart: "09:00",
		ScheduledEnd:   "12:00",
		Status:         task.StatusScheduled,
		CreatedAt:      now,
	}

	err := repo.CreateTask(ctx, original)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// Get the task
	got, err := repo.GetTask(ctx, original.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if got == nil {
		t.Fatal("expected task, got nil")
	}

	// Verify fields
	if got.ID != original.ID {
		t.Errorf("ID: got %d, want %d", got.ID, original.ID)
	}
	if got.Description != original.Description {
		t.Errorf("Description: got %q, want %q", got.Description, original.Description)
	}
	if got.Category != original.Category {
		t.Errorf("Category: got %q, want %q", got.Category, original.Category)
	}
	// Compare by date components (year, month, day) since dates are stored without timezone
	// and returned in local timezone for consistency with time.Now()
	if got.ScheduledDate.Year() != original.ScheduledDate.Year() ||
		got.ScheduledDate.Month() != original.ScheduledDate.Month() ||
		got.ScheduledDate.Day() != original.ScheduledDate.Day() {
		t.Errorf("ScheduledDate: got %v, want %v", got.ScheduledDate, original.ScheduledDate)
	}
	if got.ScheduledStart != original.ScheduledStart {
		t.Errorf("ScheduledStart: got %q, want %q", got.ScheduledStart, original.ScheduledStart)
	}
	if got.ScheduledEnd != original.ScheduledEnd {
		t.Errorf("ScheduledEnd: got %q, want %q", got.ScheduledEnd, original.ScheduledEnd)
	}
	if got.Status != original.Status {
		t.Errorf("Status: got %q, want %q", got.Status, original.Status)
	}
}

func TestGetTask_NotFound(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	got, err := repo.GetTask(ctx, 9999)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for non-existent task, got %+v", got)
	}
}

func TestGetTask_WithOutcomeAndPostponedFrom(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create original task
	original := &task.Task{
		Description:    "Original task",
		Category:       task.CategoryShallow,
		ScheduledDate:  date,
		ScheduledStart: "10:00",
		ScheduledEnd:   "11:00",
		Status:         task.StatusPostponed,
		CreatedAt:      time.Now(),
	}
	err := repo.CreateTask(ctx, original)
	if err != nil {
		t.Fatalf("CreateTask (original) failed: %v", err)
	}

	// Create postponed task with outcome
	outcome := task.OutcomeOver
	postponed := &task.Task{
		Description:    "Postponed task",
		Category:       task.CategoryDeep,
		ScheduledDate:  date.AddDate(0, 0, 1),
		ScheduledStart: "14:00",
		ScheduledEnd:   "16:00",
		Status:         task.StatusScheduled,
		Outcome:        &outcome,
		PostponedFrom:  &original.ID,
		CreatedAt:      time.Now(),
	}
	err = repo.CreateTask(ctx, postponed)
	if err != nil {
		t.Fatalf("CreateTask (postponed) failed: %v", err)
	}

	// Get and verify
	got, err := repo.GetTask(ctx, postponed.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}

	if got.Outcome == nil {
		t.Error("expected Outcome to be set")
	} else if *got.Outcome != outcome {
		t.Errorf("Outcome: got %q, want %q", *got.Outcome, outcome)
	}

	if got.PostponedFrom == nil {
		t.Error("expected PostponedFrom to be set")
	} else if *got.PostponedFrom != original.ID {
		t.Errorf("PostponedFrom: got %d, want %d", *got.PostponedFrom, original.ID)
	}
}

func TestCancelTask(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	// Create a task
	tsk := &task.Task{
		Description:    "Task to cancel",
		Category:       task.CategoryDeep,
		ScheduledDate:  time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
		ScheduledStart: "09:00",
		ScheduledEnd:   "10:00",
		Status:         task.StatusScheduled,
		CreatedAt:      time.Now(),
	}
	err := repo.CreateTask(ctx, tsk)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// Cancel it
	err = repo.CancelTask(ctx, tsk.ID)
	if err != nil {
		t.Fatalf("CancelTask failed: %v", err)
	}

	// Verify status changed
	got, err := repo.GetTask(ctx, tsk.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if got.Status != task.StatusCancelled {
		t.Errorf("expected status %q, got %q", task.StatusCancelled, got.Status)
	}
}

func TestCancelTask_NotFound(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	err := repo.CancelTask(ctx, 9999)
	if err == nil {
		t.Error("expected error for non-existent task")
	}
}

func TestSetTaskOutcome(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	// Create a task
	tsk := &task.Task{
		Description:    "Task with outcome",
		Category:       task.CategoryDeep,
		ScheduledDate:  time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
		ScheduledStart: "09:00",
		ScheduledEnd:   "10:00",
		Status:         task.StatusScheduled,
		CreatedAt:      time.Now(),
	}
	err := repo.CreateTask(ctx, tsk)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// Set outcome
	err = repo.SetTaskOutcome(ctx, tsk.ID, task.OutcomeOnTime)
	if err != nil {
		t.Fatalf("SetTaskOutcome failed: %v", err)
	}

	// Verify outcome is set
	got, err := repo.GetTask(ctx, tsk.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if got.Outcome == nil {
		t.Fatal("expected Outcome to be set")
	}
	if *got.Outcome != task.OutcomeOnTime {
		t.Errorf("expected outcome %q, got %q", task.OutcomeOnTime, *got.Outcome)
	}
}

func TestSetTaskOutcome_NotFound(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	err := repo.SetTaskOutcome(ctx, 9999, task.OutcomeOver)
	if err == nil {
		t.Error("expected error for non-existent task")
	}
}

func TestListTasksByDateRange(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	// Create tasks across multiple days
	jan8 := time.Date(2025, 1, 8, 0, 0, 0, 0, time.UTC)
	jan9 := time.Date(2025, 1, 9, 0, 0, 0, 0, time.UTC)
	jan10 := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
	jan11 := time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC)

	tasks := []*task.Task{
		{
			Description:    "Task on Jan 8",
			Category:       task.CategoryDeep,
			ScheduledDate:  jan8,
			ScheduledStart: "09:00",
			ScheduledEnd:   "10:00",
			Status:         task.StatusScheduled,
			CreatedAt:      time.Now(),
		},
		{
			Description:    "Task on Jan 9 morning",
			Category:       task.CategoryShallow,
			ScheduledDate:  jan9,
			ScheduledStart: "09:00",
			ScheduledEnd:   "10:00",
			Status:         task.StatusScheduled,
			CreatedAt:      time.Now(),
		},
		{
			Description:    "Task on Jan 9 afternoon",
			Category:       task.CategoryDeep,
			ScheduledDate:  jan9,
			ScheduledStart: "14:00",
			ScheduledEnd:   "16:00",
			Status:         task.StatusScheduled,
			CreatedAt:      time.Now(),
		},
		{
			Description:    "Task on Jan 10",
			Category:       task.CategoryShallow,
			ScheduledDate:  jan10,
			ScheduledStart: "11:00",
			ScheduledEnd:   "12:00",
			Status:         task.StatusCancelled,
			CreatedAt:      time.Now(),
		},
		{
			Description:    "Task on Jan 11",
			Category:       task.CategoryDeep,
			ScheduledDate:  jan11,
			ScheduledStart: "09:00",
			ScheduledEnd:   "11:00",
			Status:         task.StatusScheduled,
			CreatedAt:      time.Now(),
		},
	}

	for _, tsk := range tasks {
		if err := repo.CreateTask(ctx, tsk); err != nil {
			t.Fatalf("CreateTask failed: %v", err)
		}
	}

	// Query Jan 9 to Jan 10 (inclusive)
	got, err := repo.ListTasksByDateRange(ctx, jan9, jan10)
	if err != nil {
		t.Fatalf("ListTasksByDateRange failed: %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(got))
	}

	// Verify ordering by date and start time
	expectedDescs := []string{"Task on Jan 9 morning", "Task on Jan 9 afternoon", "Task on Jan 10"}
	for i, tsk := range got {
		if tsk.Description != expectedDescs[i] {
			t.Errorf("task %d: expected description %q, got %q", i, expectedDescs[i], tsk.Description)
		}
	}
}

func TestListTasksByDateRange_SingleDay(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create multiple tasks on the same day
	tasks := []*task.Task{
		{
			Description:    "Afternoon task",
			Category:       task.CategoryDeep,
			ScheduledDate:  date,
			ScheduledStart: "14:00",
			ScheduledEnd:   "16:00",
			Status:         task.StatusScheduled,
			CreatedAt:      time.Now(),
		},
		{
			Description:    "Morning task",
			Category:       task.CategoryShallow,
			ScheduledDate:  date,
			ScheduledStart: "09:00",
			ScheduledEnd:   "10:00",
			Status:         task.StatusScheduled,
			CreatedAt:      time.Now(),
		},
	}

	for _, tsk := range tasks {
		if err := repo.CreateTask(ctx, tsk); err != nil {
			t.Fatalf("CreateTask failed: %v", err)
		}
	}

	got, err := repo.ListTasksByDateRange(ctx, date, date)
	if err != nil {
		t.Fatalf("ListTasksByDateRange failed: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(got))
	}

	// Verify ordering by start time (morning before afternoon)
	if got[0].Description != "Morning task" {
		t.Errorf("expected first task to be 'Morning task', got %q", got[0].Description)
	}
	if got[1].Description != "Afternoon task" {
		t.Errorf("expected second task to be 'Afternoon task', got %q", got[1].Description)
	}
}

func TestListTasksByDateRange_Empty(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	// Create a task outside the query range
	jan5 := time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC)
	tsk := &task.Task{
		Description:    "Task on Jan 5",
		Category:       task.CategoryDeep,
		ScheduledDate:  jan5,
		ScheduledStart: "09:00",
		ScheduledEnd:   "10:00",
		Status:         task.StatusScheduled,
		CreatedAt:      time.Now(),
	}
	if err := repo.CreateTask(ctx, tsk); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// Query a range with no tasks
	jan10 := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
	jan15 := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	got, err := repo.ListTasksByDateRange(ctx, jan10, jan15)
	if err != nil {
		t.Fatalf("ListTasksByDateRange failed: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(got))
	}
}

func TestListTasksByDateRange_WithNullableFields(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create original task (will be postponed)
	original := &task.Task{
		Description:    "Original task",
		Category:       task.CategoryShallow,
		ScheduledDate:  date,
		ScheduledStart: "10:00",
		ScheduledEnd:   "11:00",
		Status:         task.StatusPostponed,
		CreatedAt:      time.Now(),
	}
	if err := repo.CreateTask(ctx, original); err != nil {
		t.Fatalf("CreateTask (original) failed: %v", err)
	}

	// Create postponed task with outcome
	outcome := task.OutcomeOver
	postponed := &task.Task{
		Description:    "Postponed task",
		Category:       task.CategoryDeep,
		ScheduledDate:  date,
		ScheduledStart: "14:00",
		ScheduledEnd:   "16:00",
		Status:         task.StatusScheduled,
		Outcome:        &outcome,
		PostponedFrom:  &original.ID,
		CreatedAt:      time.Now(),
	}
	if err := repo.CreateTask(ctx, postponed); err != nil {
		t.Fatalf("CreateTask (postponed) failed: %v", err)
	}

	got, err := repo.ListTasksByDateRange(ctx, date, date)
	if err != nil {
		t.Fatalf("ListTasksByDateRange failed: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(got))
	}

	// Find the postponed task and verify nullable fields
	var foundPostponed *task.Task
	for _, tsk := range got {
		if tsk.Description == "Postponed task" {
			foundPostponed = tsk
			break
		}
	}

	if foundPostponed == nil {
		t.Fatal("did not find postponed task")
	}

	if foundPostponed.Outcome == nil {
		t.Error("expected Outcome to be set")
	} else if *foundPostponed.Outcome != task.OutcomeOver {
		t.Errorf("expected outcome %q, got %q", task.OutcomeOver, *foundPostponed.Outcome)
	}

	if foundPostponed.PostponedFrom == nil {
		t.Error("expected PostponedFrom to be set")
	} else if *foundPostponed.PostponedFrom != original.ID {
		t.Errorf("expected PostponedFrom %d, got %d", original.ID, *foundPostponed.PostponedFrom)
	}
}

func TestCreateTasks(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	date := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)
	now := time.Now()

	tasks := []*task.Task{
		{
			Description:    "Batch task 1",
			Category:       task.CategoryDeep,
			ScheduledDate:  date,
			ScheduledStart: "09:00",
			ScheduledEnd:   "10:00",
			Status:         task.StatusScheduled,
			CreatedAt:      now,
		},
		{
			Description:    "Batch task 2",
			Category:       task.CategoryShallow,
			ScheduledDate:  date,
			ScheduledStart: "10:00",
			ScheduledEnd:   "11:00",
			Status:         task.StatusScheduled,
			CreatedAt:      now,
		},
		{
			Description:    "Batch task 3",
			Category:       task.CategoryDeep,
			ScheduledDate:  date,
			ScheduledStart: "14:00",
			ScheduledEnd:   "16:00",
			Status:         task.StatusScheduled,
			CreatedAt:      now,
		},
	}

	err := repo.CreateTasks(ctx, tasks)
	if err != nil {
		t.Fatalf("CreateTasks failed: %v", err)
	}

	// Verify all tasks got IDs assigned
	for i, tsk := range tasks {
		if tsk.ID == 0 {
			t.Errorf("task %d: expected ID to be set", i)
		}
	}

	// Verify IDs are unique and sequential
	if tasks[1].ID != tasks[0].ID+1 || tasks[2].ID != tasks[1].ID+1 {
		t.Errorf("expected sequential IDs, got %d, %d, %d", tasks[0].ID, tasks[1].ID, tasks[2].ID)
	}

	// Verify tasks can be retrieved
	for _, tsk := range tasks {
		got, err := repo.GetTask(ctx, tsk.ID)
		if err != nil {
			t.Fatalf("GetTask(%d) failed: %v", tsk.ID, err)
		}
		if got == nil {
			t.Fatalf("GetTask(%d) returned nil", tsk.ID)
		}
		if got.Description != tsk.Description {
			t.Errorf("Description: got %q, want %q", got.Description, tsk.Description)
		}
	}
}

func TestCreateTasks_Empty(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	err := repo.CreateTasks(ctx, []*task.Task{})
	if err != nil {
		t.Fatalf("CreateTasks with empty slice should succeed, got: %v", err)
	}

	err = repo.CreateTasks(ctx, nil)
	if err != nil {
		t.Fatalf("CreateTasks with nil slice should succeed, got: %v", err)
	}
}

func TestCreateTasks_WithNullableFields(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	date := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)
	now := time.Now()

	// First create an original task to reference
	original := &task.Task{
		Description:    "Original for batch",
		Category:       task.CategoryShallow,
		ScheduledDate:  date,
		ScheduledStart: "08:00",
		ScheduledEnd:   "09:00",
		Status:         task.StatusPostponed,
		CreatedAt:      now,
	}
	if err := repo.CreateTask(ctx, original); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	outcome := task.OutcomeUnder
	tasks := []*task.Task{
		{
			Description:    "Batch with outcome",
			Category:       task.CategoryDeep,
			ScheduledDate:  date,
			ScheduledStart: "09:00",
			ScheduledEnd:   "10:00",
			Status:         task.StatusScheduled,
			Outcome:        &outcome,
			CreatedAt:      now,
		},
		{
			Description:    "Batch with postponed_from",
			Category:       task.CategoryShallow,
			ScheduledDate:  date,
			ScheduledStart: "10:00",
			ScheduledEnd:   "11:00",
			Status:         task.StatusScheduled,
			PostponedFrom:  &original.ID,
			CreatedAt:      now,
		},
	}

	err := repo.CreateTasks(ctx, tasks)
	if err != nil {
		t.Fatalf("CreateTasks failed: %v", err)
	}

	// Verify nullable fields were saved
	got1, _ := repo.GetTask(ctx, tasks[0].ID)
	if got1.Outcome == nil || *got1.Outcome != task.OutcomeUnder {
		t.Errorf("expected outcome %q, got %v", task.OutcomeUnder, got1.Outcome)
	}

	got2, _ := repo.GetTask(ctx, tasks[1].ID)
	if got2.PostponedFrom == nil || *got2.PostponedFrom != original.ID {
		t.Errorf("expected PostponedFrom %d, got %v", original.ID, got2.PostponedFrom)
	}
}

func TestPostponeTask(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	// Create original task
	originalDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	original := &task.Task{
		Description:    "Task to postpone",
		Category:       task.CategoryDeep,
		ScheduledDate:  originalDate,
		ScheduledStart: "09:00",
		ScheduledEnd:   "11:00",
		Status:         task.StatusScheduled,
		CreatedAt:      time.Now(),
	}
	if err := repo.CreateTask(ctx, original); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// Postpone to next day
	newDate := time.Date(2025, 1, 16, 0, 0, 0, 0, time.UTC)
	newTask, err := repo.PostponeTask(ctx, original.ID, newDate, "14:00", "16:00")
	if err != nil {
		t.Fatalf("PostponeTask failed: %v", err)
	}

	// Verify new task
	if newTask.ID == 0 {
		t.Error("expected new task to have ID")
	}
	if newTask.ID == original.ID {
		t.Error("new task should have different ID than original")
	}
	if newTask.Description != original.Description {
		t.Errorf("Description: got %q, want %q", newTask.Description, original.Description)
	}
	if newTask.Category != original.Category {
		t.Errorf("Category: got %q, want %q", newTask.Category, original.Category)
	}
	if !newTask.ScheduledDate.Equal(newDate) {
		t.Errorf("ScheduledDate: got %v, want %v", newTask.ScheduledDate, newDate)
	}
	if newTask.ScheduledStart != "14:00" {
		t.Errorf("ScheduledStart: got %q, want %q", newTask.ScheduledStart, "14:00")
	}
	if newTask.ScheduledEnd != "16:00" {
		t.Errorf("ScheduledEnd: got %q, want %q", newTask.ScheduledEnd, "16:00")
	}
	if newTask.Status != task.StatusScheduled {
		t.Errorf("Status: got %q, want %q", newTask.Status, task.StatusScheduled)
	}
	if newTask.PostponedFrom == nil || *newTask.PostponedFrom != original.ID {
		t.Errorf("PostponedFrom: got %v, want %d", newTask.PostponedFrom, original.ID)
	}

	// Verify original task is now marked as postponed
	updatedOriginal, err := repo.GetTask(ctx, original.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if updatedOriginal.Status != task.StatusPostponed {
		t.Errorf("original status: got %q, want %q", updatedOriginal.Status, task.StatusPostponed)
	}

	// Verify new task can be retrieved from DB
	retrieved, err := repo.GetTask(ctx, newTask.ID)
	if err != nil {
		t.Fatalf("GetTask (new) failed: %v", err)
	}
	if retrieved.PostponedFrom == nil || *retrieved.PostponedFrom != original.ID {
		t.Errorf("retrieved PostponedFrom: got %v, want %d", retrieved.PostponedFrom, original.ID)
	}
}

func TestPostponeTask_NotFound(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	newDate := time.Date(2025, 1, 16, 0, 0, 0, 0, time.UTC)
	_, err := repo.PostponeTask(ctx, 9999, newDate, "14:00", "16:00")
	if err == nil {
		t.Error("expected error for non-existent task")
	}
}

func TestPostponeTask_PreservesCategory(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	// Create a shallow task
	originalDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	original := &task.Task{
		Description:    "Shallow task to postpone",
		Category:       task.CategoryShallow,
		ScheduledDate:  originalDate,
		ScheduledStart: "10:00",
		ScheduledEnd:   "10:30",
		Status:         task.StatusScheduled,
		CreatedAt:      time.Now(),
	}
	if err := repo.CreateTask(ctx, original); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	newDate := time.Date(2025, 1, 17, 0, 0, 0, 0, time.UTC)
	newTask, err := repo.PostponeTask(ctx, original.ID, newDate, "11:00", "11:30")
	if err != nil {
		t.Fatalf("PostponeTask failed: %v", err)
	}

	if newTask.Category != task.CategoryShallow {
		t.Errorf("Category: got %q, want %q", newTask.Category, task.CategoryShallow)
	}
}

// newTestRepo creates a temporary SQLite repository for testing.
func newTestRepo(t *testing.T) *SQLite {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create test repo: %v", err)
	}

	t.Cleanup(func() {
		_ = repo.Close()
	})

	return repo
}

func TestCreateTask_OverlapError(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create first task 09:00-11:00
	first := &task.Task{
		Description:    "First task",
		Category:       task.CategoryDeep,
		ScheduledDate:  date,
		ScheduledStart: "09:00",
		ScheduledEnd:   "11:00",
		Status:         task.StatusScheduled,
		CreatedAt:      time.Now(),
	}
	if err := repo.CreateTask(ctx, first); err != nil {
		t.Fatalf("CreateTask (first) failed: %v", err)
	}

	tests := []struct {
		name  string
		start string
		end   string
	}{
		{"exact same time", "09:00", "11:00"},
		{"starts during existing", "10:00", "12:00"},
		{"ends during existing", "08:00", "10:00"},
		{"contained within existing", "09:30", "10:30"},
		{"contains existing", "08:00", "12:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overlapping := &task.Task{
				Description:    "Overlapping task",
				Category:       task.CategoryShallow,
				ScheduledDate:  date,
				ScheduledStart: tt.start,
				ScheduledEnd:   tt.end,
				Status:         task.StatusScheduled,
				CreatedAt:      time.Now(),
			}
			err := repo.CreateTask(ctx, overlapping)
			if err == nil {
				t.Error("expected overlap error, got nil")
			}
			if !errors.Is(err, task.ErrTimeBlockOverlap) {
				t.Errorf("expected ErrTimeBlockOverlap, got: %v", err)
			}
		})
	}
}

func TestCreateTask_NoOverlapWithAdjacentTasks(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create first task 10:00-11:00
	first := &task.Task{
		Description:    "First task",
		Category:       task.CategoryDeep,
		ScheduledDate:  date,
		ScheduledStart: "10:00",
		ScheduledEnd:   "11:00",
		Status:         task.StatusScheduled,
		CreatedAt:      time.Now(),
	}
	if err := repo.CreateTask(ctx, first); err != nil {
		t.Fatalf("CreateTask (first) failed: %v", err)
	}

	// Task immediately before (09:00-10:00) should succeed
	before := &task.Task{
		Description:    "Before task",
		Category:       task.CategoryShallow,
		ScheduledDate:  date,
		ScheduledStart: "09:00",
		ScheduledEnd:   "10:00",
		Status:         task.StatusScheduled,
		CreatedAt:      time.Now(),
	}
	if err := repo.CreateTask(ctx, before); err != nil {
		t.Errorf("adjacent task before should succeed: %v", err)
	}

	// Task immediately after (11:00-12:00) should succeed
	after := &task.Task{
		Description:    "After task",
		Category:       task.CategoryShallow,
		ScheduledDate:  date,
		ScheduledStart: "11:00",
		ScheduledEnd:   "12:00",
		Status:         task.StatusScheduled,
		CreatedAt:      time.Now(),
	}
	if err := repo.CreateTask(ctx, after); err != nil {
		t.Errorf("adjacent task after should succeed: %v", err)
	}
}

func TestCreateTask_NoOverlapOnDifferentDays(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	jan15 := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	jan16 := time.Date(2025, 1, 16, 0, 0, 0, 0, time.UTC)

	// Create task on Jan 15
	first := &task.Task{
		Description:    "Jan 15 task",
		Category:       task.CategoryDeep,
		ScheduledDate:  jan15,
		ScheduledStart: "09:00",
		ScheduledEnd:   "11:00",
		Status:         task.StatusScheduled,
		CreatedAt:      time.Now(),
	}
	if err := repo.CreateTask(ctx, first); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// Same time slot on different day should succeed
	second := &task.Task{
		Description:    "Jan 16 task",
		Category:       task.CategoryDeep,
		ScheduledDate:  jan16,
		ScheduledStart: "09:00",
		ScheduledEnd:   "11:00",
		Status:         task.StatusScheduled,
		CreatedAt:      time.Now(),
	}
	if err := repo.CreateTask(ctx, second); err != nil {
		t.Errorf("same time on different day should succeed: %v", err)
	}
}

func TestCreateTask_NoOverlapWithCancelledTask(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create and cancel a task
	cancelled := &task.Task{
		Description:    "Cancelled task",
		Category:       task.CategoryDeep,
		ScheduledDate:  date,
		ScheduledStart: "09:00",
		ScheduledEnd:   "11:00",
		Status:         task.StatusScheduled,
		CreatedAt:      time.Now(),
	}
	if err := repo.CreateTask(ctx, cancelled); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}
	if err := repo.CancelTask(ctx, cancelled.ID); err != nil {
		t.Fatalf("CancelTask failed: %v", err)
	}

	// Same time slot should now be available
	newTask := &task.Task{
		Description:    "New task at same time",
		Category:       task.CategoryDeep,
		ScheduledDate:  date,
		ScheduledStart: "09:00",
		ScheduledEnd:   "11:00",
		Status:         task.StatusScheduled,
		CreatedAt:      time.Now(),
	}
	if err := repo.CreateTask(ctx, newTask); err != nil {
		t.Errorf("should allow task in cancelled slot: %v", err)
	}
}

func TestCreateTasks_BatchOverlapError(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	date := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)

	// Batch with overlapping tasks
	tasks := []*task.Task{
		{
			Description:    "Task 1",
			Category:       task.CategoryDeep,
			ScheduledDate:  date,
			ScheduledStart: "09:00",
			ScheduledEnd:   "11:00",
			Status:         task.StatusScheduled,
			CreatedAt:      time.Now(),
		},
		{
			Description:    "Task 2 overlaps with 1",
			Category:       task.CategoryShallow,
			ScheduledDate:  date,
			ScheduledStart: "10:00",
			ScheduledEnd:   "12:00",
			Status:         task.StatusScheduled,
			CreatedAt:      time.Now(),
		},
	}

	err := repo.CreateTasks(ctx, tasks)
	if err == nil {
		t.Error("expected overlap error, got nil")
	}
	if !errors.Is(err, task.ErrTimeBlockOverlap) {
		t.Errorf("expected ErrTimeBlockOverlap, got: %v", err)
	}
}

func TestCreateTasks_OverlapWithExisting(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	date := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)

	// Create existing task
	existing := &task.Task{
		Description:    "Existing task",
		Category:       task.CategoryDeep,
		ScheduledDate:  date,
		ScheduledStart: "09:00",
		ScheduledEnd:   "11:00",
		Status:         task.StatusScheduled,
		CreatedAt:      time.Now(),
	}
	if err := repo.CreateTask(ctx, existing); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// Batch that overlaps with existing
	tasks := []*task.Task{
		{
			Description:    "New task 1",
			Category:       task.CategoryShallow,
			ScheduledDate:  date,
			ScheduledStart: "10:00",
			ScheduledEnd:   "12:00",
			Status:         task.StatusScheduled,
			CreatedAt:      time.Now(),
		},
	}

	err := repo.CreateTasks(ctx, tasks)
	if err == nil {
		t.Error("expected overlap error, got nil")
	}
	if !errors.Is(err, task.ErrTimeBlockOverlap) {
		t.Errorf("expected ErrTimeBlockOverlap, got: %v", err)
	}
}

func TestPostponeTask_OverlapError(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	nextDate := time.Date(2025, 1, 16, 0, 0, 0, 0, time.UTC)

	// Create task to postpone
	toPostpone := &task.Task{
		Description:    "Task to postpone",
		Category:       task.CategoryDeep,
		ScheduledDate:  date,
		ScheduledStart: "09:00",
		ScheduledEnd:   "11:00",
		Status:         task.StatusScheduled,
		CreatedAt:      time.Now(),
	}
	if err := repo.CreateTask(ctx, toPostpone); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// Create existing task on target date
	existing := &task.Task{
		Description:    "Existing on target date",
		Category:       task.CategoryShallow,
		ScheduledDate:  nextDate,
		ScheduledStart: "14:00",
		ScheduledEnd:   "16:00",
		Status:         task.StatusScheduled,
		CreatedAt:      time.Now(),
	}
	if err := repo.CreateTask(ctx, existing); err != nil {
		t.Fatalf("CreateTask (existing) failed: %v", err)
	}

	// Try to postpone to overlapping slot
	_, err := repo.PostponeTask(ctx, toPostpone.ID, nextDate, "15:00", "17:00")
	if err == nil {
		t.Error("expected overlap error, got nil")
	}
	if !errors.Is(err, task.ErrTimeBlockOverlap) {
		t.Errorf("expected ErrTimeBlockOverlap, got: %v", err)
	}

	// Verify original task is still scheduled (transaction rolled back)
	original, _ := repo.GetTask(ctx, toPostpone.ID)
	if original.Status != task.StatusScheduled {
		t.Errorf("original should still be scheduled, got %q", original.Status)
	}
}

func TestUpdateTask(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create task
	tsk := &task.Task{
		Description:    "Task to update",
		Category:       task.CategoryDeep,
		ScheduledDate:  date,
		ScheduledStart: "09:00",
		ScheduledEnd:   "10:00",
		Status:         task.StatusScheduled,
		CreatedAt:      time.Now(),
	}
	if err := repo.CreateTask(ctx, tsk); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// Update times (grow by 15 min)
	err := repo.UpdateTask(ctx, tsk.ID, "09:00", "10:15")
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}

	// Verify update
	updated, err := repo.GetTask(ctx, tsk.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if updated.ScheduledStart != "09:00" {
		t.Errorf("ScheduledStart: got %q, want %q", updated.ScheduledStart, "09:00")
	}
	if updated.ScheduledEnd != "10:15" {
		t.Errorf("ScheduledEnd: got %q, want %q", updated.ScheduledEnd, "10:15")
	}
}

func TestUpdateTaskDescription(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	tsk, err := task.New("Original", "deep", "2025-01-15", "09:00", "10:00")
	if err != nil {
		t.Fatalf("New task failed: %v", err)
	}

	if err := repo.CreateTask(ctx, tsk); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	if err := repo.UpdateTaskDescription(ctx, tsk.ID, "Updated"); err != nil {
		t.Fatalf("UpdateTaskDescription failed: %v", err)
	}

	got, err := repo.GetTask(ctx, tsk.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if got == nil {
		t.Fatalf("expected task, got nil")
	}
	if got.Description != "Updated" {
		t.Fatalf("description = %q, want %q", got.Description, "Updated")
	}
}

func TestUpdateTaskDescription_Empty(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	tsk, err := task.New("Original", "deep", "2025-01-15", "09:00", "10:00")
	if err != nil {
		t.Fatalf("New task failed: %v", err)
	}

	if err := repo.CreateTask(ctx, tsk); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	err = repo.UpdateTaskDescription(ctx, tsk.ID, " ")
	if !errors.Is(err, task.ErrEmptyDescription) {
		t.Fatalf("error = %v, want %v", err, task.ErrEmptyDescription)
	}
}

func TestUpdateTaskDescription_NotFound(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	err := repo.UpdateTaskDescription(ctx, 9999, "Updated")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestUpdateTask_NotFound(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	err := repo.UpdateTask(ctx, 9999, "09:00", "10:00")
	if err == nil {
		t.Error("expected error for non-existent task")
	}
}

func TestUpdateTask_OverlapError(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create two adjacent tasks
	first := &task.Task{
		Description:    "First task",
		Category:       task.CategoryDeep,
		ScheduledDate:  date,
		ScheduledStart: "09:00",
		ScheduledEnd:   "10:00",
		Status:         task.StatusScheduled,
		CreatedAt:      time.Now(),
	}
	if err := repo.CreateTask(ctx, first); err != nil {
		t.Fatalf("CreateTask (first) failed: %v", err)
	}

	second := &task.Task{
		Description:    "Second task",
		Category:       task.CategoryShallow,
		ScheduledDate:  date,
		ScheduledStart: "10:00",
		ScheduledEnd:   "11:00",
		Status:         task.StatusScheduled,
		CreatedAt:      time.Now(),
	}
	if err := repo.CreateTask(ctx, second); err != nil {
		t.Fatalf("CreateTask (second) failed: %v", err)
	}

	// Try to grow first task into second task's time
	err := repo.UpdateTask(ctx, first.ID, "09:00", "10:30")
	if err == nil {
		t.Error("expected overlap error, got nil")
	}
	if !errors.Is(err, task.ErrTimeBlockOverlap) {
		t.Errorf("expected ErrTimeBlockOverlap, got: %v", err)
	}

	// Verify first task is unchanged
	unchanged, _ := repo.GetTask(ctx, first.ID)
	if unchanged.ScheduledEnd != "10:00" {
		t.Errorf("task should be unchanged, got end time %q", unchanged.ScheduledEnd)
	}
}

func TestUpdateTask_NoSelfOverlap(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create task
	tsk := &task.Task{
		Description:    "Task to grow",
		Category:       task.CategoryDeep,
		ScheduledDate:  date,
		ScheduledStart: "09:00",
		ScheduledEnd:   "11:00",
		Status:         task.StatusScheduled,
		CreatedAt:      time.Now(),
	}
	if err := repo.CreateTask(ctx, tsk); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// Should be able to update to same times (no self-overlap)
	err := repo.UpdateTask(ctx, tsk.ID, "09:00", "11:00")
	if err != nil {
		t.Errorf("updating to same times should succeed: %v", err)
	}

	// Should be able to shrink
	err = repo.UpdateTask(ctx, tsk.ID, "09:00", "10:00")
	if err != nil {
		t.Errorf("shrinking should succeed: %v", err)
	}
}

func TestParseDate_LocalTimezone(t *testing.T) {
	// This tests that parseDate returns dates in local timezone,
	// which is critical for matching with time.Now()-based dates in the TUI.
	tests := []struct {
		name  string
		input string
	}{
		{"date only", "2025-01-15"},
		{"date only different month", "2025-06-20"},
		{"date only end of year", "2025-12-31"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parseDate(tt.input)
			if err != nil {
				t.Fatalf("parseDate(%q) failed: %v", tt.input, err)
			}

			// The parsed date should be in local timezone
			if parsed.Location() != time.Local {
				t.Errorf("parseDate(%q) location = %v, want %v", tt.input, parsed.Location(), time.Local)
			}

			// Create a date the same way the TUI does (using time.Now's location)
			localMidnight := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.Local)

			// They should be equal
			if !parsed.Equal(localMidnight) {
				t.Errorf("parseDate(%q) = %v, want %v (should match local midnight)", tt.input, parsed, localMidnight)
			}
		})
	}
}

func TestParseDate_AllFormats(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"date only", "2025-01-15", false},
		{"datetime with Z", "2025-01-15T09:00:00Z", false},
		{"datetime without tz", "2025-01-15 09:00:00", false},
		{"RFC3339", "2025-01-15T09:00:00+05:00", false},
		{"invalid format", "15/01/2025", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseDate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDate(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestListTasksByDateRange_ReturnsLocalTimezone(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	// Create task with local timezone date
	localDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.Local)
	tsk := &task.Task{
		Description:    "Test task",
		Category:       task.CategoryDeep,
		ScheduledDate:  localDate,
		ScheduledStart: "09:00",
		ScheduledEnd:   "10:00",
		Status:         task.StatusScheduled,
		CreatedAt:      time.Now(),
	}

	if err := repo.CreateTask(ctx, tsk); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// Fetch tasks
	got, err := repo.ListTasksByDateRange(ctx, localDate, localDate)
	if err != nil {
		t.Fatalf("ListTasksByDateRange failed: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 task, got %d", len(got))
	}

	// The returned task's ScheduledDate should be in Local timezone
	if got[0].ScheduledDate.Location() != time.Local {
		t.Errorf("expected ScheduledDate location to be Local, got %v", got[0].ScheduledDate.Location())
	}

	// The date should match the input date (same year, month, day)
	if got[0].ScheduledDate.Year() != localDate.Year() ||
		got[0].ScheduledDate.Month() != localDate.Month() ||
		got[0].ScheduledDate.Day() != localDate.Day() {
		t.Errorf("date mismatch: got %v, want %v", got[0].ScheduledDate, localDate)
	}

	// Equal() should return true since both are in local timezone at midnight
	if !got[0].ScheduledDate.Equal(localDate) {
		t.Errorf("ScheduledDate.Equal() failed: got %v, want %v", got[0].ScheduledDate, localDate)
	}
}

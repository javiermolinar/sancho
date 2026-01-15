package integration

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/javiermolinar/sancho/internal/db"
	"github.com/javiermolinar/sancho/internal/task"
)

// openRepo creates a fresh repository for each test with automatic cleanup.
func openRepo(t *testing.T) *db.SQLite {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	repo, err := db.New(dbPath)
	if err != nil {
		t.Fatalf("failed to open repo: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	return repo
}

// mustParseDate parses a date string or fails the test.
func mustParseDate(t *testing.T, s string) time.Time {
	t.Helper()
	date, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("failed to parse date %q: %v", s, err)
	}
	return date
}

// createTask is a helper to create and insert a task.
func createTask(t *testing.T, repo *db.SQLite, desc, category, date, start, end string) *task.Task {
	t.Helper()
	ctx := context.Background()
	tsk, err := task.New(desc, category, date, start, end)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}
	if err := repo.CreateTask(ctx, tsk); err != nil {
		t.Fatalf("failed to insert task: %v", err)
	}
	return tsk
}

func TestCreateTask(t *testing.T) {
	repo := openRepo(t)
	ctx := context.Background()

	tsk, err := task.New("Integration test task", "deep", "2025-01-20", "08:00", "09:00")
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	if err := repo.CreateTask(ctx, tsk); err != nil {
		t.Fatalf("failed to insert task: %v", err)
	}

	if tsk.ID == 0 {
		t.Error("expected task ID to be set after insert")
	}

	// Verify the task was actually inserted
	got, err := repo.GetTask(ctx, tsk.ID)
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}
	if got == nil {
		t.Fatalf("task %d not found in database", tsk.ID)
	}
	if got.Description != "Integration test task" {
		t.Errorf("Description: got %q, want %q", got.Description, "Integration test task")
	}
	if got.ScheduledStart != "08:00" {
		t.Errorf("ScheduledStart: got %q, want %q", got.ScheduledStart, "08:00")
	}
	if got.ScheduledEnd != "09:00" {
		t.Errorf("ScheduledEnd: got %q, want %q", got.ScheduledEnd, "09:00")
	}
	if got.Category != task.CategoryDeep {
		t.Errorf("Category: got %q, want %q", got.Category, task.CategoryDeep)
	}
	if got.Status != task.StatusScheduled {
		t.Errorf("Status: got %q, want %q", got.Status, task.StatusScheduled)
	}
}

func TestCreateTask_WithCategory(t *testing.T) {
	repo := openRepo(t)
	ctx := context.Background()

	tsk, err := task.New("Shallow work task", "shallow", "2025-01-20", "10:00", "10:30")
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	if err := repo.CreateTask(ctx, tsk); err != nil {
		t.Fatalf("failed to insert task: %v", err)
	}

	got, err := repo.GetTask(ctx, tsk.ID)
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}
	if got == nil {
		t.Fatalf("task %d not found in database", tsk.ID)
	}
	if got.Category != task.CategoryShallow {
		t.Errorf("Category: got %q, want %q", got.Category, task.CategoryShallow)
	}
}

func TestNewTask_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		desc    string
		cat     string
		date    string
		start   string
		end     string
		wantErr error
	}{
		{
			name:    "empty description",
			desc:    "",
			cat:     "deep",
			date:    "2025-01-20",
			start:   "09:00",
			end:     "10:00",
			wantErr: task.ErrEmptyDescription,
		},
		{
			name:    "invalid category",
			desc:    "test",
			cat:     "invalid",
			date:    "2025-01-20",
			start:   "09:00",
			end:     "10:00",
			wantErr: task.ErrInvalidCategory,
		},
		{
			name:    "end before start",
			desc:    "test",
			cat:     "deep",
			date:    "2025-01-20",
			start:   "10:00",
			end:     "09:00",
			wantErr: task.ErrEndBeforeStart,
		},
		{
			name:    "invalid start time format",
			desc:    "test",
			cat:     "deep",
			date:    "2025-01-20",
			start:   "9:00",
			end:     "10:00",
			wantErr: task.ErrInvalidTimeFormat,
		},
		{
			name:    "invalid end time format",
			desc:    "test",
			cat:     "deep",
			date:    "2025-01-20",
			start:   "09:00",
			end:     "10",
			wantErr: task.ErrInvalidTimeFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := task.New(tt.desc, tt.cat, tt.date, tt.start, tt.end)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetTask_NotFound(t *testing.T) {
	repo := openRepo(t)
	ctx := context.Background()

	got, err := repo.GetTask(ctx, 99999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for non-existent task, got %+v", got)
	}
}

func TestCancelTask(t *testing.T) {
	repo := openRepo(t)
	ctx := context.Background()

	tsk := createTask(t, repo, "Task to cancel", "deep", "2025-01-21", "11:00", "12:00")

	if err := repo.CancelTask(ctx, tsk.ID); err != nil {
		t.Fatalf("failed to cancel task: %v", err)
	}

	got, err := repo.GetTask(ctx, tsk.ID)
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}
	if got == nil {
		t.Fatalf("task %d not found in database", tsk.ID)
	}
	if got.Status != task.StatusCancelled {
		t.Errorf("Status: got %q, want %q", got.Status, task.StatusCancelled)
	}
}

func TestCancelTask_NotFound(t *testing.T) {
	repo := openRepo(t)
	ctx := context.Background()

	err := repo.CancelTask(ctx, 99999)
	if err == nil {
		t.Fatal("expected error for non-existent task")
	}
	// Error message should mention "not found"
	if !contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error message, got: %v", err)
	}
}

func TestSetTaskOutcome(t *testing.T) {
	repo := openRepo(t)
	ctx := context.Background()

	tsk := createTask(t, repo, "Task with outcome", "deep", "2025-01-21", "13:00", "14:00")

	if err := repo.SetTaskOutcome(ctx, tsk.ID, task.OutcomeOnTime); err != nil {
		t.Fatalf("failed to set outcome: %v", err)
	}

	got, err := repo.GetTask(ctx, tsk.ID)
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}
	if got == nil {
		t.Fatalf("task %d not found in database", tsk.ID)
	}
	if got.Outcome == nil {
		t.Fatal("expected Outcome to be set")
	}
	if *got.Outcome != task.OutcomeOnTime {
		t.Errorf("Outcome: got %q, want %q", *got.Outcome, task.OutcomeOnTime)
	}
}

func TestSetTaskOutcome_AllValues(t *testing.T) {
	outcomes := []task.Outcome{
		task.OutcomeOnTime,
		task.OutcomeOver,
		task.OutcomeUnder,
	}

	for _, outcome := range outcomes {
		t.Run(string(outcome), func(t *testing.T) {
			repo := openRepo(t)
			ctx := context.Background()

			tsk := createTask(t, repo, "Task for "+string(outcome), "deep", "2025-01-21", "09:00", "10:00")

			if err := repo.SetTaskOutcome(ctx, tsk.ID, outcome); err != nil {
				t.Fatalf("failed to set outcome: %v", err)
			}

			got, err := repo.GetTask(ctx, tsk.ID)
			if err != nil {
				t.Fatalf("failed to get task: %v", err)
			}
			if got.Outcome == nil {
				t.Fatal("expected Outcome to be set")
			}
			if *got.Outcome != outcome {
				t.Errorf("Outcome: got %q, want %q", *got.Outcome, outcome)
			}
		})
	}
}

func TestSetTaskOutcome_InvalidOutcome(t *testing.T) {
	invalidOutcome := task.Outcome("invalid")
	if invalidOutcome.Valid() {
		t.Error("expected invalid outcome to return false from Valid()")
	}

	// Valid outcomes should return true
	validOutcomes := []task.Outcome{task.OutcomeOnTime, task.OutcomeOver, task.OutcomeUnder}
	for _, o := range validOutcomes {
		if !o.Valid() {
			t.Errorf("expected %q to be valid", o)
		}
	}
}

func TestSetTaskOutcome_NotFound(t *testing.T) {
	repo := openRepo(t)
	ctx := context.Background()

	err := repo.SetTaskOutcome(ctx, 99999, task.OutcomeOnTime)
	if err == nil {
		t.Fatal("expected error for non-existent task")
	}
	if !contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error message, got: %v", err)
	}
}

func TestListTasksByDateRange(t *testing.T) {
	repo := openRepo(t)
	ctx := context.Background()

	// Create tasks on specific dates
	createTask(t, repo, "List test morning", "deep", "2025-02-01", "09:00", "10:00")
	createTask(t, repo, "List test afternoon", "deep", "2025-02-01", "14:00", "15:00")
	createTask(t, repo, "List test next day", "deep", "2025-02-02", "10:00", "11:00")

	// List single day
	start := mustParseDate(t, "2025-02-01")
	end := mustParseDate(t, "2025-02-01")
	tasks, err := repo.ListTasksByDateRange(ctx, start, end)
	if err != nil {
		t.Fatalf("failed to list tasks: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}

	// Verify ordering (morning should appear before afternoon)
	if tasks[0].Description != "List test morning" {
		t.Errorf("expected first task to be 'List test morning', got %q", tasks[0].Description)
	}
	if tasks[1].Description != "List test afternoon" {
		t.Errorf("expected second task to be 'List test afternoon', got %q", tasks[1].Description)
	}
}

func TestListTasksByDateRange_MultiDay(t *testing.T) {
	repo := openRepo(t)
	ctx := context.Background()

	// Create tasks across multiple days
	createTask(t, repo, "Range test day1", "deep", "2025-03-01", "09:00", "10:00")
	createTask(t, repo, "Range test day2", "deep", "2025-03-02", "09:00", "10:00")
	createTask(t, repo, "Range test day3", "deep", "2025-03-03", "09:00", "10:00")
	createTask(t, repo, "Range test outside", "deep", "2025-03-05", "09:00", "10:00")

	// List date range (should include days 1-3, not day 5)
	start := mustParseDate(t, "2025-03-01")
	end := mustParseDate(t, "2025-03-03")
	tasks, err := repo.ListTasksByDateRange(ctx, start, end)
	if err != nil {
		t.Fatalf("failed to list tasks: %v", err)
	}

	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}

	// Verify correct tasks are included
	descriptions := make(map[string]bool)
	for _, tsk := range tasks {
		descriptions[tsk.Description] = true
	}

	for _, expected := range []string{"Range test day1", "Range test day2", "Range test day3"} {
		if !descriptions[expected] {
			t.Errorf("expected task %q to be in results", expected)
		}
	}
	if descriptions["Range test outside"] {
		t.Error("task 'Range test outside' should not be in results")
	}
}

func TestListTasksByDateRange_Empty(t *testing.T) {
	repo := openRepo(t)
	ctx := context.Background()

	// List a date range with no tasks
	start := mustParseDate(t, "2099-01-01")
	end := mustParseDate(t, "2099-01-31")
	tasks, err := repo.ListTasksByDateRange(ctx, start, end)
	if err != nil {
		t.Fatalf("failed to list tasks: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(tasks))
	}
}

func TestPostponeTask(t *testing.T) {
	repo := openRepo(t)
	ctx := context.Background()

	original := createTask(t, repo, "Task to postpone", "deep", "2025-04-01", "09:00", "10:00")

	// Postpone to new date/time
	newDate := mustParseDate(t, "2025-04-02")
	newTask, err := repo.PostponeTask(ctx, original.ID, newDate, "14:00", "15:00")
	if err != nil {
		t.Fatalf("failed to postpone task: %v", err)
	}

	// Verify original task is now postponed
	got, err := repo.GetTask(ctx, original.ID)
	if err != nil {
		t.Fatalf("failed to get original task: %v", err)
	}
	if got.Status != task.StatusPostponed {
		t.Errorf("original status: got %q, want %q", got.Status, task.StatusPostponed)
	}

	// Verify new task was created correctly
	if newTask.ID == 0 {
		t.Error("expected new task ID to be set")
	}
	if newTask.ID == original.ID {
		t.Error("new task should have different ID than original")
	}
	if newTask.Description != original.Description {
		t.Errorf("new task description: got %q, want %q", newTask.Description, original.Description)
	}
	if newTask.Category != original.Category {
		t.Errorf("new task category: got %q, want %q", newTask.Category, original.Category)
	}
	if newTask.ScheduledStart != "14:00" {
		t.Errorf("new task start: got %q, want %q", newTask.ScheduledStart, "14:00")
	}
	if newTask.ScheduledEnd != "15:00" {
		t.Errorf("new task end: got %q, want %q", newTask.ScheduledEnd, "15:00")
	}
	if newTask.Status != task.StatusScheduled {
		t.Errorf("new task status: got %q, want %q", newTask.Status, task.StatusScheduled)
	}
	if newTask.PostponedFrom == nil || *newTask.PostponedFrom != original.ID {
		t.Errorf("new task PostponedFrom: got %v, want %d", newTask.PostponedFrom, original.ID)
	}

	// Verify new task appears on the new date
	tasks, err := repo.ListTasksByDateRange(ctx, newDate, newDate)
	if err != nil {
		t.Fatalf("failed to list tasks: %v", err)
	}
	found := false
	for _, tsk := range tasks {
		if tsk.ID == newTask.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("postponed task not found on new date")
	}
}

func TestPostponeTask_NotFound(t *testing.T) {
	repo := openRepo(t)
	ctx := context.Background()

	newDate := mustParseDate(t, "2025-04-10")
	_, err := repo.PostponeTask(ctx, 99999, newDate, "14:00", "15:00")
	if err == nil {
		t.Fatal("expected error for non-existent task")
	}
	if !contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error message, got: %v", err)
	}
}

func TestCreateTasks_Batch(t *testing.T) {
	repo := openRepo(t)
	ctx := context.Background()

	tasks := []*task.Task{}
	for i, desc := range []string{"Batch task 1", "Batch task 2", "Batch task 3"} {
		start, end := getTimeSlot(i)
		tsk, err := task.New(desc, "deep", "2025-05-10", start, end)
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}
		tasks = append(tasks, tsk)
	}

	if err := repo.CreateTasks(ctx, tasks); err != nil {
		t.Fatalf("failed to create batch tasks: %v", err)
	}

	// Verify all tasks were created with IDs
	for _, tsk := range tasks {
		if tsk.ID == 0 {
			t.Errorf("expected task %q to have ID set", tsk.Description)
		}
	}

	// Verify they can be retrieved
	start := mustParseDate(t, "2025-05-10")
	got, err := repo.ListTasksByDateRange(ctx, start, start)
	if err != nil {
		t.Fatalf("failed to list tasks: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(got))
	}
}

func TestTimeBlockOverlap(t *testing.T) {
	repo := openRepo(t)
	ctx := context.Background()

	// Create first task
	createTask(t, repo, "First task", "deep", "2025-06-01", "09:00", "10:00")

	// Try to create overlapping task
	overlapping, err := task.New("Overlapping task", "deep", "2025-06-01", "09:30", "10:30")
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	err = repo.CreateTask(ctx, overlapping)
	if err == nil {
		t.Fatal("expected error for overlapping time block")
	}
	if !errors.Is(err, task.ErrTimeBlockOverlap) {
		t.Errorf("expected ErrTimeBlockOverlap, got: %v", err)
	}
}

func TestTimeBlockOverlap_Batch(t *testing.T) {
	repo := openRepo(t)
	ctx := context.Background()

	// Try to create batch with overlapping tasks
	task1, _ := task.New("Batch overlap 1", "deep", "2025-06-02", "09:00", "10:00")
	task2, _ := task.New("Batch overlap 2", "deep", "2025-06-02", "09:30", "10:30")

	err := repo.CreateTasks(ctx, []*task.Task{task1, task2})
	if err == nil {
		t.Fatal("expected error for overlapping batch tasks")
	}
	if !errors.Is(err, task.ErrTimeBlockOverlap) {
		t.Errorf("expected ErrTimeBlockOverlap, got: %v", err)
	}
}

// TestFullWorkflow tests a complete task lifecycle: create, list, set outcome, postpone, cancel
func TestFullWorkflow(t *testing.T) {
	repo := openRepo(t)
	ctx := context.Background()

	// 1. Create multiple tasks
	task1 := createTask(t, repo, "Deep work session", "deep", "2025-05-01", "09:00", "12:00")
	task2 := createTask(t, repo, "Email catchup", "shallow", "2025-05-01", "13:00", "14:00")
	task3 := createTask(t, repo, "Code review", "shallow", "2025-05-01", "14:00", "15:00")

	// 2. List tasks for the day
	start := mustParseDate(t, "2025-05-01")
	tasks, err := repo.ListTasksByDateRange(ctx, start, start)
	if err != nil {
		t.Fatalf("failed to list tasks: %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}

	// 3. Set outcome on completed task
	if err := repo.SetTaskOutcome(ctx, task1.ID, task.OutcomeOnTime); err != nil {
		t.Fatalf("failed to set outcome: %v", err)
	}
	got1, _ := repo.GetTask(ctx, task1.ID)
	if got1.Outcome == nil || *got1.Outcome != task.OutcomeOnTime {
		t.Errorf("task1 outcome: got %v, want %q", got1.Outcome, task.OutcomeOnTime)
	}

	// 4. Postpone a task to next day
	nextDay := mustParseDate(t, "2025-05-02")
	_, err = repo.PostponeTask(ctx, task2.ID, nextDay, "10:00", "11:00")
	if err != nil {
		t.Fatalf("failed to postpone task: %v", err)
	}
	got2, _ := repo.GetTask(ctx, task2.ID)
	if got2.Status != task.StatusPostponed {
		t.Errorf("task2 status: got %q, want %q", got2.Status, task.StatusPostponed)
	}

	// Verify new task was created for next day
	nextDayTasks, err := repo.ListTasksByDateRange(ctx, nextDay, nextDay)
	if err != nil {
		t.Fatalf("failed to list next day: %v", err)
	}
	found := false
	for _, tsk := range nextDayTasks {
		if tsk.Description == "Email catchup" {
			found = true
			break
		}
	}
	if !found {
		t.Error("postponed task not found on next day")
	}

	// 5. Cancel a task
	if err := repo.CancelTask(ctx, task3.ID); err != nil {
		t.Fatalf("failed to cancel task: %v", err)
	}
	got3, _ := repo.GetTask(ctx, task3.ID)
	if got3.Status != task.StatusCancelled {
		t.Errorf("task3 status: got %q, want %q", got3.Status, task.StatusCancelled)
	}

	// 6. List original day - verify task1 is still there
	finalTasks, err := repo.ListTasksByDateRange(ctx, start, start)
	if err != nil {
		t.Fatalf("failed to list final: %v", err)
	}
	foundSancho := false
	for _, tsk := range finalTasks {
		if tsk.Description == "Deep work session" {
			foundSancho = true
			break
		}
	}
	if !foundSancho {
		t.Error("expected 'Deep work session' to still be in original day listing")
	}
}

// contains checks if a string contains a substring.
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

// getTimeSlot returns non-overlapping time slots for batch testing.
func getTimeSlot(i int) (start, end string) {
	starts := []string{"09:00", "10:00", "11:00"}
	ends := []string{"09:30", "10:30", "11:30"}
	return starts[i], ends[i]
}

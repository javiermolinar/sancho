package integration

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/javiermolinar/sancho/internal/db"
	"github.com/javiermolinar/sancho/internal/task"
)

func TestTimezoneDebug(t *testing.T) {
	// Create temp db and add a task for today
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := db.New(dbPath)
	if err != nil {
		t.Fatalf("Error opening db: %v", err)
	}
	defer func() { _ = repo.Close() }()

	// Get current week
	now := time.Now()
	t.Logf("Current time: %v", now)
	t.Logf("Current location: %v", now.Location())

	// Calculate Monday of this week (same as TUI does)
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := now.AddDate(0, 0, -(weekday - 1))
	monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
	sunday := monday.AddDate(0, 0, 6)

	t.Logf("Week start (Monday): %v", monday)
	t.Logf("Week end (Sunday): %v", sunday)

	// Create a task for today
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	testTask := &task.Task{
		Description:    "Test task for today",
		Category:       task.CategoryDeep,
		ScheduledDate:  today,
		ScheduledStart: "10:00",
		ScheduledEnd:   "11:00",
		Status:         task.StatusScheduled,
		CreatedAt:      now,
	}
	if err := repo.CreateTask(context.Background(), testTask); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}
	t.Logf("Created task with ScheduledDate: %v (location: %v)", testTask.ScheduledDate, testTask.ScheduledDate.Location())

	// Fetch tasks
	ctx := context.Background()
	tasks, err := repo.ListTasksByDateRange(ctx, monday, sunday)
	if err != nil {
		t.Fatalf("Error fetching tasks: %v", err)
	}

	t.Logf("Fetched %d tasks:", len(tasks))
	for _, tsk := range tasks {
		t.Logf("  Task #%d: %s on %v (%s - %s)",
			tsk.ID, tsk.Description, tsk.ScheduledDate, tsk.ScheduledStart, tsk.ScheduledEnd)
		t.Logf("    Date location: %v", tsk.ScheduledDate.Location())
	}

	if len(tasks) == 0 {
		t.Fatalf("Expected at least 1 task, got 0")
	}

	// Create Week from tasks (same as TUI does)
	week := task.NewWeekFromTasks(monday, tasks)
	t.Logf("Week created from monday: %v (location: %v)", week.StartDate, week.StartDate.Location())

	totalScheduled := 0
	for i := 0; i < 7; i++ {
		day := week.Day(i)
		scheduled := day.ScheduledTasks()
		t.Logf("  Day %d (%v, location: %v): %d scheduled tasks", i, day.Date, day.Date.Location(), len(scheduled))
		for _, tsk := range scheduled {
			t.Logf("    - %s (%s-%s)", tsk.Description, tsk.ScheduledStart, tsk.ScheduledEnd)
		}
		totalScheduled += len(scheduled)
	}

	if totalScheduled == 0 {
		// Debug: check DayByDate
		t.Logf("\nDirect DayByDate check:")
		taskDate := tasks[0].ScheduledDate
		t.Logf("  Task date: %v (location: %v)", taskDate, taskDate.Location())
		t.Logf("  Looking for day...")
		day := week.DayByDate(taskDate)
		if day == nil {
			t.Logf("  DayByDate returned nil!")
			// Check each day
			for i := 0; i < 7; i++ {
				d := week.Day(i)
				t.Logf("  Day %d: %v (location: %v), Equal=%v",
					i, d.Date, d.Date.Location(), d.Date.Equal(taskDate))
				// Also compare year/month/day
				sameDateComponents := d.Date.Year() == taskDate.Year() &&
					d.Date.Month() == taskDate.Month() &&
					d.Date.Day() == taskDate.Day()
				t.Logf("    Same Y/M/D components: %v", sameDateComponents)
			}
		} else {
			t.Logf("  DayByDate found day: %v", day.Date)
		}
		t.Fatalf("Expected tasks to be assigned to Week days, but totalScheduled=0")
	}

	fmt.Printf("SUCCESS: %d tasks properly assigned to week days\n", totalScheduled)
}

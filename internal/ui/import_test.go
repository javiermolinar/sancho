package ui

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/javiermolinar/sancho/internal/db"
	"github.com/javiermolinar/sancho/internal/task"
)

func TestImportTasks(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "source.db")
	destPath := filepath.Join(dir, "dest.db")

	sourceRepo, err := db.New(sourcePath)
	if err != nil {
		t.Fatalf("creating source repo: %v", err)
	}
	defer func() { _ = sourceRepo.Close() }()

	date := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
	original := &task.Task{
		Description:    "Original task",
		Category:       task.CategoryDeep,
		ScheduledDate:  date,
		ScheduledStart: "09:00",
		ScheduledEnd:   "10:00",
		Status:         task.StatusPostponed,
		CreatedAt:      time.Now(),
	}
	if err := sourceRepo.CreateTask(ctx, original); err != nil {
		t.Fatalf("CreateTask (original) failed: %v", err)
	}

	outcome := task.OutcomeOnTime
	postponed := &task.Task{
		Description:    "Postponed task",
		Category:       task.CategoryShallow,
		ScheduledDate:  date.AddDate(0, 0, 1),
		ScheduledStart: "11:00",
		ScheduledEnd:   "12:00",
		Status:         task.StatusScheduled,
		Outcome:        &outcome,
		PostponedFrom:  &original.ID,
		CreatedAt:      time.Now(),
	}
	if err := sourceRepo.CreateTask(ctx, postponed); err != nil {
		t.Fatalf("CreateTask (postponed) failed: %v", err)
	}

	destRepo, err := db.New(destPath)
	if err != nil {
		t.Fatalf("creating destination repo: %v", err)
	}
	defer func() { _ = destRepo.Close() }()

	count, err := importTasks(ctx, destRepo, sourcePath)
	if err != nil {
		t.Fatalf("importTasks failed: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 imported tasks, got %d", count)
	}

	imported, err := destRepo.ListAllTasks(ctx)
	if err != nil {
		t.Fatalf("ListAllTasks failed: %v", err)
	}

	if len(imported) != 2 {
		t.Fatalf("expected 2 tasks in destination, got %d", len(imported))
	}

	var importedOriginal *task.Task
	var importedPostponed *task.Task
	for _, tsk := range imported {
		switch tsk.Description {
		case "Original task":
			importedOriginal = tsk
		case "Postponed task":
			importedPostponed = tsk
		}
	}

	if importedOriginal == nil {
		t.Fatal("missing original task after import")
	}
	if importedPostponed == nil {
		t.Fatal("missing postponed task after import")
	}

	if importedPostponed.Outcome == nil {
		t.Fatal("expected outcome to be set on postponed task")
	}
	if *importedPostponed.Outcome != outcome {
		t.Fatalf("expected outcome %q, got %q", outcome, *importedPostponed.Outcome)
	}

	if importedPostponed.PostponedFrom == nil {
		t.Fatal("expected PostponedFrom to be set on postponed task")
	}
	if *importedPostponed.PostponedFrom != importedOriginal.ID {
		t.Fatalf("expected PostponedFrom %d, got %d", importedOriginal.ID, *importedPostponed.PostponedFrom)
	}
}

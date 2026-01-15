package llm

import (
	"strings"
	"testing"
	"time"
)

func TestBuildInitialMessages_IncludesWeekdayContext(t *testing.T) {
	planner := NewPlanner(nil)
	req := PlanRequest{
		Input:    "Plan tasks for tomorrow",
		Date:     time.Date(2026, 1, 8, 9, 30, 0, 0, time.UTC), // Thursday
		DayStart: "09:00",
		DayEnd:   "18:00",
		RecentTasks: []ExistingTask{
			{
				Date:        "2026-01-07",
				Start:       "10:00",
				End:         "11:00",
				Description: "Review backlog",
				Category:    "deep",
			},
		},
	}

	msgs := planner.BuildInitialMessages(req)
	if len(msgs) != 1 {
		t.Fatalf("messages = %d, want 1", len(msgs))
	}

	content := msgs[0].Content
	if !strings.Contains(content, "Today: Thursday (2026-01-08, weekday)") {
		t.Fatalf("missing today context: %s", content)
	}
	if !strings.Contains(content, "Tomorrow: Friday (2026-01-09, weekday)") {
		t.Fatalf("missing tomorrow context: %s", content)
	}
	if !strings.Contains(content, "Recent schedule history (last 14 days):") {
		t.Fatalf("missing recent history header: %s", content)
	}
	if !strings.Contains(content, "2026-01-07 10:00-11:00: Review backlog [deep]") {
		t.Fatalf("missing recent history entry: %s", content)
	}
}

func TestBuildInitialMessages_CompactPrompt(t *testing.T) {
	planner := NewPlanner(nil)
	req := PlanRequest{
		Input:            "Plan tasks for today",
		Date:             time.Date(2026, 1, 8, 9, 30, 0, 0, time.UTC),
		DayStart:         "09:00",
		DayEnd:           "18:00",
		UseCompactPrompt: true,
	}

	msgs := planner.BuildInitialMessages(req)
	if len(msgs) != 1 {
		t.Fatalf("messages = %d, want 1", len(msgs))
	}

	content := msgs[0].Content
	if strings.Contains(content, "CRITICAL DATE RULES") {
		t.Fatalf("expected compact prompt without critical rules: %s", content)
	}
	if strings.Contains(content, "Date parsing:") {
		t.Fatalf("expected compact prompt without date parsing: %s", content)
	}
	if !strings.Contains(content, "Existing scheduled tasks:") {
		t.Fatalf("missing existing tasks section: %s", content)
	}
}

func TestSortedExistingTasks_ByDateTime(t *testing.T) {
	tasks := []ExistingTask{
		{Date: "2026-01-08", Start: "09:00", End: "10:00", Description: "B", Category: "deep"},
		{Date: "2026-01-07", Start: "11:00", End: "12:00", Description: "C", Category: "shallow"},
		{Date: "2026-01-07", Start: "09:00", End: "10:00", Description: "A", Category: "deep"},
	}

	sorted := sortedExistingTasks(tasks)

	if sorted[0].Date != "2026-01-07" || sorted[0].Start != "09:00" {
		t.Fatalf("first task = %+v, want 2026-01-07 09:00", sorted[0])
	}
	if sorted[1].Date != "2026-01-07" || sorted[1].Start != "11:00" {
		t.Fatalf("second task = %+v, want 2026-01-07 11:00", sorted[1])
	}
	if sorted[2].Date != "2026-01-08" || sorted[2].Start != "09:00" {
		t.Fatalf("third task = %+v, want 2026-01-08 09:00", sorted[2])
	}
}

func TestSuggestedTimeWindows_Median(t *testing.T) {
	tasks := []ExistingTask{
		{Date: "2026-01-06", Start: "12:00", End: "13:00", Description: "Train", Category: "deep"},
		{Date: "2026-01-07", Start: "12:30", End: "13:30", Description: "Train", Category: "deep"},
		{Date: "2026-01-08", Start: "13:00", End: "14:00", Description: "Train", Category: "deep"},
	}

	suggestions := suggestedTimeWindows(tasks)
	if len(suggestions) != 1 {
		t.Fatalf("suggestions = %d, want 1", len(suggestions))
	}
	if suggestions[0] != "Train: ~12:30-13:30 (n=3)" {
		t.Fatalf("suggestion = %q, want %q", suggestions[0], "Train: ~12:30-13:30 (n=3)")
	}
}

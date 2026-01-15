package llm

import (
	"testing"
	"time"
)

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "raw json object",
			input:    `{"tasks": []}`,
			expected: `{"tasks": []}`,
		},
		{
			name:     "json with leading text",
			input:    `Here is the response: {"tasks": [{"description": "test"}]}`,
			expected: `{"tasks": [{"description": "test"}]}`,
		},
		{
			name:     "json in code block",
			input:    "```json\n{\"tasks\": []}\n```",
			expected: `{"tasks": []}`,
		},
		{
			name:     "json in plain code block",
			input:    "```\n{\"tasks\": []}\n```",
			expected: `{"tasks": []}`,
		},
		{
			name:     "json array",
			input:    `[{"id": 1}, {"id": 2}]`,
			expected: `[{"id": 1}, {"id": 2}]`,
		},
		{
			name:     "nested json",
			input:    `{"outer": {"inner": {"deep": true}}}`,
			expected: `{"outer": {"inner": {"deep": true}}}`,
		},
		{
			name: "markdown with explanation",
			input: `Here's my analysis:

` + "```json" + `
{
  "tasks": [
    {"description": "Write code", "category": "deep"}
  ]
}
` + "```" + `

Let me know if you need anything else.`,
			expected: `{
  "tasks": [
    {"description": "Write code", "category": "deep"}
  ]
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractJSON(tt.input)
			if got != tt.expected {
				t.Errorf("extractJSON() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestPlanResponseToTasks(t *testing.T) {
	resp := &PlanResponse{
		Tasks: []PlannedTask{
			{
				Description:    "Write thesis",
				Category:       "deep",
				ScheduledDate:  "2025-01-10",
				ScheduledStart: "09:00",
				ScheduledEnd:   "11:00",
			},
			{
				Description:    "Check emails",
				Category:       "shallow",
				ScheduledDate:  "2025-01-10",
				ScheduledStart: "11:00",
				ScheduledEnd:   "11:30",
			},
			{
				Description:    "Review PRs",
				Category:       "deep",
				ScheduledDate:  "2025-01-13",
				ScheduledStart: "09:00",
				ScheduledEnd:   "10:00",
			},
		},
		Warnings:    []string{"Busy day ahead"},
		Suggestions: []string{"Take a break at noon"},
	}

	tasks, err := resp.ToTasks()
	if err != nil {
		t.Fatalf("ToTasks() error = %v", err)
	}

	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}

	todayDate := parseTestDate("2025-01-10")
	nextWorkday := parseTestDate("2025-01-13")

	// Check first task (deep, today)
	if tasks[0].Description != "Write thesis" {
		t.Errorf("task[0].Description = %q, want %q", tasks[0].Description, "Write thesis")
	}
	if tasks[0].Category != "deep" {
		t.Errorf("task[0].Category = %q, want %q", tasks[0].Category, "deep")
	}
	if tasks[0].ScheduledStart != "09:00" {
		t.Errorf("task[0].ScheduledStart = %q, want %q", tasks[0].ScheduledStart, "09:00")
	}
	if !tasks[0].ScheduledDate.Equal(todayDate) {
		t.Errorf("task[0].ScheduledDate = %v, want %v", tasks[0].ScheduledDate, todayDate)
	}

	// Check second task (shallow, today)
	if tasks[1].Category != "shallow" {
		t.Errorf("task[1].Category = %q, want %q", tasks[1].Category, "shallow")
	}

	// Check third task (deep, next_workday)
	if tasks[2].Description != "Review PRs" {
		t.Errorf("task[2].Description = %q, want %q", tasks[2].Description, "Review PRs")
	}
	if !tasks[2].ScheduledDate.Equal(nextWorkday) {
		t.Errorf("task[2].ScheduledDate = %v, want %v", tasks[2].ScheduledDate, nextWorkday)
	}
}

func TestPlanResponseToTasksWithFallback(t *testing.T) {
	resp := &PlanResponse{
		Tasks: []PlannedTask{
			{
				Description:    "Write thesis",
				Category:       "deep",
				ScheduledDate:  "today",
				ScheduledStart: "09:00",
				ScheduledEnd:   "11:00",
			},
			{
				Description:    "Review PRs",
				Category:       "deep",
				ScheduledDate:  "next_workday",
				ScheduledStart: "09:00",
				ScheduledEnd:   "10:00",
			},
		},
	}

	todayDate := parseTestDate("2025-01-10")
	nextWorkday := parseTestDate("2025-01-13")
	tasks, err := resp.ToTasksWithFallback(todayDate, nextWorkday)
	if err != nil {
		t.Fatalf("ToTasksWithFallback() error = %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}

	if !tasks[0].ScheduledDate.Equal(todayDate) {
		t.Errorf("task[0].ScheduledDate = %v, want %v", tasks[0].ScheduledDate, todayDate)
	}
	if !tasks[1].ScheduledDate.Equal(nextWorkday) {
		t.Errorf("task[1].ScheduledDate = %v, want %v", tasks[1].ScheduledDate, nextWorkday)
	}
}

func parseTestDate(s string) (t time.Time) {
	t, _ = time.Parse("2006-01-02", s)
	return
}

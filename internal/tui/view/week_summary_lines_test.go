package view

import (
	"strings"
	"testing"
	"time"

	"github.com/javiermolinar/sancho/internal/summary"
	"github.com/javiermolinar/sancho/internal/task"
)

func TestBuildWeekTasksLines(t *testing.T) {
	monday := time.Date(2025, 1, 6, 0, 0, 0, 0, time.Local)
	tasks := []*task.Task{
		{
			Description:    "Deep work",
			Category:       task.CategoryDeep,
			ScheduledDate:  monday,
			ScheduledStart: "09:00",
			ScheduledEnd:   "10:00",
			Status:         task.StatusScheduled,
		},
	}
	summaryData := summary.SummarizeWeek(monday, tasks, summary.WeekSummaryOptions{})

	lines := BuildWeekTasksLines(summaryData)
	text := linesToText(lines)

	if !strings.Contains(text, "Mon Jan 6") {
		t.Fatalf("expected day header in tasks text, got %q", text)
	}
	if !strings.Contains(text, "[D] 09:00-10:00 Deep work") {
		t.Fatalf("expected task line in tasks text, got %q", text)
	}
}

func TestBuildWeekSummaryLinesIncludesInsight(t *testing.T) {
	monday := time.Date(2025, 1, 6, 0, 0, 0, 0, time.Local)
	tasks := []*task.Task{
		{
			Description:    "Deep work",
			Category:       task.CategoryDeep,
			ScheduledDate:  monday,
			ScheduledStart: "09:00",
			ScheduledEnd:   "10:00",
			Status:         task.StatusScheduled,
		},
	}
	summaryData := summary.SummarizeWeek(monday, tasks, summary.WeekSummaryOptions{})
	summaryData.Insight = "INSIGHT LINE"

	lines := BuildWeekSummaryLines(summaryData, false)
	text := linesToText(lines)

	if !strings.Contains(text, "INSIGHT") {
		t.Fatalf("expected insight header in summary text, got %q", text)
	}
	if !strings.Contains(text, "INSIGHT LINE") {
		t.Fatalf("expected insight content in summary text, got %q", text)
	}
}

// Package view provides rendering helpers for the TUI.
package view

import (
	"fmt"
	"strings"

	"github.com/javiermolinar/sancho/internal/summary"
	"github.com/javiermolinar/sancho/internal/task"
)

// BuildWeekSummaryLines builds summary lines for the week summary modal.
func BuildWeekSummaryLines(summary *summary.WeekSummary, showPeak bool) []WeekSummaryLine {
	lines := make([]WeekSummaryLine, 0, 16)
	dateLine := fmt.Sprintf("%s - %s", summary.Start.Format("Mon Jan 2"), summary.End.Format("Mon Jan 2, 2006"))
	lines = append(lines, WeekSummaryLine{Text: dateLine, Style: WeekSummaryLineMeta})
	lines = append(lines, WeekSummaryLine{Text: ""})

	if len(summary.Tasks) == 0 {
		lines = append(lines, WeekSummaryLine{Text: "No time blocks scheduled for this week."})
		return lines
	}

	stats := summary.Stats
	lines = append(lines, WeekSummaryLine{
		Text:  fmt.Sprintf("Deep: %s (%d%%)", FormatDuration(stats.DeepMinutes), stats.DeepPercent()),
		Style: WeekSummaryLineBody,
	})
	lines = append(lines, WeekSummaryLine{
		Text:  fmt.Sprintf("Shallow: %s", FormatDuration(stats.ShallowMinutes)),
		Style: WeekSummaryLineBody,
	})
	lines = append(lines, WeekSummaryLine{
		Text:  fmt.Sprintf("Ratio: %s | Blocks: %d", stats.Ratio(), stats.TotalBlocks),
		Style: WeekSummaryLineBody,
	})

	if bestDay, bestDeep := stats.BestDay(); bestDay >= 0 {
		bestLine := fmt.Sprintf("Best day: %s (%s deep)", task.WeekdayName(bestDay), FormatDuration(bestDeep))
		lines = append(lines, WeekSummaryLine{Text: bestLine})
	}

	if showPeak && stats.DeepMinutes > 0 {
		peakLine := fmt.Sprintf("Peak: %d%% (%s of %s deep)",
			stats.PeakPercent(),
			FormatDuration(stats.PeakDeepMinutes),
			FormatDuration(stats.DeepMinutes))
		lines = append(lines, WeekSummaryLine{Text: peakLine})
	}

	if stats.CancelledBlocks > 0 || stats.PostponedBlocks > 0 {
		line := fmt.Sprintf("Cancelled: %d | Postponed: %d", stats.CancelledBlocks, stats.PostponedBlocks)
		lines = append(lines, WeekSummaryLine{Text: line, Style: WeekSummaryLineMeta})
	}

	if summary.Insight != "" {
		lines = append(lines, WeekSummaryLine{Text: ""})
		lines = append(lines, WeekSummaryLine{Text: "INSIGHT", Style: WeekSummaryLineSection})
		for _, line := range strings.Split(summary.Insight, "\n") {
			lines = append(lines, WeekSummaryLine{Text: line})
		}
	}

	return lines
}

// BuildWeekTasksLines builds task lines for the week summary modal.
func BuildWeekTasksLines(summary *summary.WeekSummary) []WeekSummaryLine {
	lines := make([]WeekSummaryLine, 0, 24)
	dateLine := fmt.Sprintf("Week: %s - %s", summary.Start.Format("Mon Jan 2"), summary.End.Format("Mon Jan 2, 2006"))
	lines = append(lines, WeekSummaryLine{Text: dateLine, Style: WeekSummaryLineMeta})
	lines = append(lines, WeekSummaryLine{Text: ""})

	if len(summary.Tasks) == 0 {
		lines = append(lines, WeekSummaryLine{Text: "No time blocks scheduled for this week."})
		return lines
	}

	var currentDate string
	for _, t := range summary.Tasks {
		date := t.ScheduledDate.Format("2006-01-02")
		dayName := t.ScheduledDate.Format("Mon Jan 2")
		if date != currentDate {
			if currentDate != "" {
				lines = append(lines, WeekSummaryLine{Text: ""})
			}
			lines = append(lines, WeekSummaryLine{Text: dayName, Style: WeekSummaryLineSection})
			currentDate = date
		}

		status := weekSummaryStatusSymbol(t.Status)
		category := "[S]"
		if t.Category == task.CategoryDeep {
			category = "[D]"
		}
		line := fmt.Sprintf("  %s %s %s-%s %s", status, category, t.ScheduledStart, t.ScheduledEnd, t.Description)
		lines = append(lines, WeekSummaryLine{Text: line})
	}

	return lines
}

// BuildWeekTasksCopyText renders week summary lines to plain text for copying.
func BuildWeekTasksCopyText(lines []WeekSummaryLine) string {
	return linesToText(lines)
}

func linesToText(lines []WeekSummaryLine) string {
	parts := make([]string, 0, len(lines))
	for _, line := range lines {
		parts = append(parts, line.Text)
	}
	return strings.Join(parts, "\n")
}

func weekSummaryStatusSymbol(s task.Status) string {
	switch s {
	case task.StatusScheduled:
		return "○"
	case task.StatusCancelled:
		return "✗"
	case task.StatusPostponed:
		return "→"
	default:
		return "?"
	}
}

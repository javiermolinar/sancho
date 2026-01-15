// Package tui provides the terminal user interface for sancho.
package tui

import (
	"fmt"
	"strings"

	"github.com/javiermolinar/sancho/internal/summary"
	"github.com/javiermolinar/sancho/internal/task"
	"github.com/javiermolinar/sancho/internal/tui/view"
)

const weekSummaryFallbackWidth = 60

func (m Model) renderWeekSummaryModal() string {
	if m.weekSummary == nil {
		return ""
	}
	body := m.weekSummaryBody()
	footer := m.weekSummaryFooter()
	return view.RenderModalFrame("Week Summary", body, footer, m.modalStyles())
}

func (m Model) weekSummaryFooter() string {
	switch m.weekSummaryView {
	case weekSummaryViewTasks:
		return view.RenderModalButtonsCompact(m.modalStyles(), "[s] Summary", "[y] Copy", "[Esc] Close")
	default:
		return view.RenderModalButtonsCompact(m.modalStyles(), "[w] Tasks", "[y] Copy", "[Esc] Close")
	}
}

func (m Model) weekSummaryBody() string {
	bodyWidth := view.ModalContentWidth(m.styles.ModalStyle, weekSummaryFallbackWidth)
	lines := m.weekSummarySummaryText
	if m.weekSummaryView == weekSummaryViewTasks {
		lines = m.weekSummaryTasksText
	}
	return view.RenderWeekSummaryBody(lines, view.WeekSummaryStyles{
		BodyStyle:         m.styles.ModalBodyStyle,
		MetaStyle:         m.styles.ModalMetaStyle,
		SectionTitleStyle: m.styles.ModalSectionTitleStyle,
	}, bodyWidth)
}

func buildWeekSummaryLines(summary *summary.WeekSummary, showPeak bool) []view.WeekSummaryLine {
	lines := make([]view.WeekSummaryLine, 0, 16)
	dateLine := fmt.Sprintf("%s - %s", summary.Start.Format("Mon Jan 2"), summary.End.Format("Mon Jan 2, 2006"))
	lines = append(lines, view.WeekSummaryLine{Text: dateLine, Style: view.WeekSummaryLineMeta})
	lines = append(lines, view.WeekSummaryLine{Text: ""})

	if len(summary.Tasks) == 0 {
		lines = append(lines, view.WeekSummaryLine{Text: "No time blocks scheduled for this week."})
		return lines
	}

	stats := summary.Stats
	lines = append(lines, view.WeekSummaryLine{
		Text:  fmt.Sprintf("Deep: %s (%d%%)", formatDuration(stats.DeepMinutes), stats.DeepPercent()),
		Style: view.WeekSummaryLineBody,
	})
	lines = append(lines, view.WeekSummaryLine{
		Text:  fmt.Sprintf("Shallow: %s", formatDuration(stats.ShallowMinutes)),
		Style: view.WeekSummaryLineBody,
	})
	lines = append(lines, view.WeekSummaryLine{
		Text:  fmt.Sprintf("Ratio: %s | Blocks: %d", stats.Ratio(), stats.TotalBlocks),
		Style: view.WeekSummaryLineBody,
	})

	if bestDay, bestDeep := stats.BestDay(); bestDay >= 0 {
		bestLine := fmt.Sprintf("Best day: %s (%s deep)", task.WeekdayName(bestDay), formatDuration(bestDeep))
		lines = append(lines, view.WeekSummaryLine{Text: bestLine})
	}

	if showPeak && stats.DeepMinutes > 0 {
		peakLine := fmt.Sprintf("Peak: %d%% (%s of %s deep)",
			stats.PeakPercent(),
			formatDuration(stats.PeakDeepMinutes),
			formatDuration(stats.DeepMinutes))
		lines = append(lines, view.WeekSummaryLine{Text: peakLine})
	}

	if stats.CancelledBlocks > 0 || stats.PostponedBlocks > 0 {
		line := fmt.Sprintf("Cancelled: %d | Postponed: %d", stats.CancelledBlocks, stats.PostponedBlocks)
		lines = append(lines, view.WeekSummaryLine{Text: line, Style: view.WeekSummaryLineMeta})
	}

	if summary.Insight != "" {
		lines = append(lines, view.WeekSummaryLine{Text: ""})
		lines = append(lines, view.WeekSummaryLine{Text: "INSIGHT", Style: view.WeekSummaryLineSection})
		for _, line := range strings.Split(summary.Insight, "\n") {
			lines = append(lines, view.WeekSummaryLine{Text: line})
		}
	}

	return lines
}

func buildWeekTasksLines(summary *summary.WeekSummary) []view.WeekSummaryLine {
	lines := make([]view.WeekSummaryLine, 0, 24)
	dateLine := fmt.Sprintf("Week: %s - %s", summary.Start.Format("Mon Jan 2"), summary.End.Format("Mon Jan 2, 2006"))
	lines = append(lines, view.WeekSummaryLine{Text: dateLine, Style: view.WeekSummaryLineMeta})
	lines = append(lines, view.WeekSummaryLine{Text: ""})

	if len(summary.Tasks) == 0 {
		lines = append(lines, view.WeekSummaryLine{Text: "No time blocks scheduled for this week."})
		return lines
	}

	var currentDate string
	for _, t := range summary.Tasks {
		date := t.ScheduledDate.Format("2006-01-02")
		dayName := t.ScheduledDate.Format("Mon Jan 2")
		if date != currentDate {
			if currentDate != "" {
				lines = append(lines, view.WeekSummaryLine{Text: ""})
			}
			lines = append(lines, view.WeekSummaryLine{Text: dayName, Style: view.WeekSummaryLineSection})
			currentDate = date
		}

		status := weekSummaryStatusSymbol(t.Status)
		category := "[S]"
		if t.Category == task.CategoryDeep {
			category = "[D]"
		}
		line := fmt.Sprintf("  %s %s %s-%s %s", status, category, t.ScheduledStart, t.ScheduledEnd, t.Description)
		lines = append(lines, view.WeekSummaryLine{Text: line})
	}

	return lines
}

func buildWeekTasksCopyText(lines []view.WeekSummaryLine) string {
	return linesToText(lines)
}

func linesToText(lines []view.WeekSummaryLine) string {
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

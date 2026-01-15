// Package tui provides the terminal user interface for sancho.
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/javiermolinar/sancho/internal/summary"
	"github.com/javiermolinar/sancho/internal/task"
	"github.com/javiermolinar/sancho/internal/tui/view"
)

const weekSummaryFallbackWidth = 60

type weekSummaryLineStyle int

const (
	weekSummaryLineBody weekSummaryLineStyle = iota
	weekSummaryLineMeta
	weekSummaryLineSection
)

type weekSummaryLine struct {
	text  string
	style weekSummaryLineStyle
}

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
	bodyWidth := modalContentWidth(m.styles.ModalStyle)
	lines := m.weekSummarySummaryText
	if m.weekSummaryView == weekSummaryViewTasks {
		lines = m.weekSummaryTasksText
	}

	rendered := make([]string, 0, len(lines))
	for _, line := range lines {
		rendered = append(rendered, wrapWeekSummaryLine(m, line, bodyWidth)...)
	}
	return strings.Join(rendered, "\n")
}

func wrapWeekSummaryLine(m Model, line weekSummaryLine, width int) []string {
	switch line.style {
	case weekSummaryLineSection:
		return wrapModalText(m.styles.ModalSectionTitleStyle, line.text, width)
	case weekSummaryLineMeta:
		return wrapModalText(m.styles.ModalMetaStyle, line.text, width)
	default:
		return wrapModalText(m.styles.ModalBodyStyle, line.text, width)
	}
}

func buildWeekSummaryLines(summary *summary.WeekSummary, showPeak bool) []weekSummaryLine {
	lines := make([]weekSummaryLine, 0, 16)
	dateLine := fmt.Sprintf("%s - %s", summary.Start.Format("Mon Jan 2"), summary.End.Format("Mon Jan 2, 2006"))
	lines = append(lines, weekSummaryLine{text: dateLine, style: weekSummaryLineMeta})
	lines = append(lines, weekSummaryLine{text: ""})

	if len(summary.Tasks) == 0 {
		lines = append(lines, weekSummaryLine{text: "No time blocks scheduled for this week."})
		return lines
	}

	stats := summary.Stats
	lines = append(lines, weekSummaryLine{
		text:  fmt.Sprintf("Deep: %s (%d%%)", formatDuration(stats.DeepMinutes), stats.DeepPercent()),
		style: weekSummaryLineBody,
	})
	lines = append(lines, weekSummaryLine{
		text:  fmt.Sprintf("Shallow: %s", formatDuration(stats.ShallowMinutes)),
		style: weekSummaryLineBody,
	})
	lines = append(lines, weekSummaryLine{
		text:  fmt.Sprintf("Ratio: %s | Blocks: %d", stats.Ratio(), stats.TotalBlocks),
		style: weekSummaryLineBody,
	})

	if bestDay, bestDeep := stats.BestDay(); bestDay >= 0 {
		bestLine := fmt.Sprintf("Best day: %s (%s deep)", task.WeekdayName(bestDay), formatDuration(bestDeep))
		lines = append(lines, weekSummaryLine{text: bestLine})
	}

	if showPeak && stats.DeepMinutes > 0 {
		peakLine := fmt.Sprintf("Peak: %d%% (%s of %s deep)",
			stats.PeakPercent(),
			formatDuration(stats.PeakDeepMinutes),
			formatDuration(stats.DeepMinutes))
		lines = append(lines, weekSummaryLine{text: peakLine})
	}

	if stats.CancelledBlocks > 0 || stats.PostponedBlocks > 0 {
		line := fmt.Sprintf("Cancelled: %d | Postponed: %d", stats.CancelledBlocks, stats.PostponedBlocks)
		lines = append(lines, weekSummaryLine{text: line, style: weekSummaryLineMeta})
	}

	if summary.Insight != "" {
		lines = append(lines, weekSummaryLine{text: ""})
		lines = append(lines, weekSummaryLine{text: "INSIGHT", style: weekSummaryLineSection})
		for _, line := range strings.Split(summary.Insight, "\n") {
			lines = append(lines, weekSummaryLine{text: line})
		}
	}

	return lines
}

func buildWeekTasksLines(summary *summary.WeekSummary) []weekSummaryLine {
	lines := make([]weekSummaryLine, 0, 24)
	dateLine := fmt.Sprintf("Week: %s - %s", summary.Start.Format("Mon Jan 2"), summary.End.Format("Mon Jan 2, 2006"))
	lines = append(lines, weekSummaryLine{text: dateLine, style: weekSummaryLineMeta})
	lines = append(lines, weekSummaryLine{text: ""})

	if len(summary.Tasks) == 0 {
		lines = append(lines, weekSummaryLine{text: "No time blocks scheduled for this week."})
		return lines
	}

	var currentDate string
	for _, t := range summary.Tasks {
		date := t.ScheduledDate.Format("2006-01-02")
		dayName := t.ScheduledDate.Format("Mon Jan 2")
		if date != currentDate {
			if currentDate != "" {
				lines = append(lines, weekSummaryLine{text: ""})
			}
			lines = append(lines, weekSummaryLine{text: dayName, style: weekSummaryLineSection})
			currentDate = date
		}

		status := weekSummaryStatusSymbol(t.Status)
		category := "[S]"
		if t.Category == task.CategoryDeep {
			category = "[D]"
		}
		line := fmt.Sprintf("  %s %s %s-%s %s", status, category, t.ScheduledStart, t.ScheduledEnd, t.Description)
		lines = append(lines, weekSummaryLine{text: line})
	}

	return lines
}

func buildWeekTasksCopyText(lines []weekSummaryLine) string {
	return linesToText(lines)
}

func linesToText(lines []weekSummaryLine) string {
	parts := make([]string, 0, len(lines))
	for _, line := range lines {
		parts = append(parts, line.text)
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

func modalContentWidth(style lipgloss.Style) int {
	width := style.GetWidth()
	if width <= 0 {
		return weekSummaryFallbackWidth
	}
	contentWidth := width - 4
	return max(10, contentWidth)
}

func wrapModalText(style lipgloss.Style, text string, width int) []string {
	if width <= 0 {
		return []string{style.Render("")}
	}
	lines := view.WrapTextToWidths(text, width, width)
	if len(lines) == 0 {
		return []string{style.Render("")}
	}

	wrapped := make([]string, 0, len(lines))
	for _, line := range lines {
		wrapped = append(wrapped, style.Render(line))
	}
	return wrapped
}

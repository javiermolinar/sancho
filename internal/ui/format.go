package ui

import (
	"fmt"
	"strings"

	"github.com/javiermolinar/sancho/internal/task"
)

// Stats holds aggregated statistics for a set of tasks.
type Stats struct {
	DeepMinutes     int
	ShallowMinutes  int
	PeakDeepMinutes int
	TotalBlocks     int
	CancelledBlocks int
	PostponedBlocks int
	DayStats        map[string]DayStats
}

// DayStats holds statistics for a single day.
type DayStats struct {
	DeepMinutes    int
	ShallowMinutes int
	Blocks         int
}

// TotalMinutes returns the sum of deep and shallow minutes.
func (s Stats) TotalMinutes() int {
	return s.DeepMinutes + s.ShallowMinutes
}

// DeepPercent returns the percentage of time spent on deep work.
func (s Stats) DeepPercent() int {
	if s.TotalMinutes() == 0 {
		return 0
	}
	return (s.DeepMinutes * 100) / s.TotalMinutes()
}

// PeakPercent returns the percentage of deep work during peak hours.
func (s Stats) PeakPercent() int {
	if s.DeepMinutes == 0 {
		return 0
	}
	return (s.PeakDeepMinutes * 100) / s.DeepMinutes
}

// Ratio returns the deep:shallow ratio as a string.
func (s Stats) Ratio() string {
	switch {
	case s.ShallowMinutes > 0:
		r := float64(s.DeepMinutes) / float64(s.ShallowMinutes)
		return fmt.Sprintf("%.1f:1", r)
	case s.DeepMinutes > 0:
		return "∞:1"
	default:
		return "0:0"
	}
}

// BestDay returns the day with the most deep work minutes.
func (s Stats) BestDay() (day string, minutes int) {
	for d, ds := range s.DayStats {
		if ds.DeepMinutes > minutes {
			minutes = ds.DeepMinutes
			day = d
		}
	}
	return day, minutes
}

// PrintOpts configures task printing behavior.
type PrintOpts struct {
	PeakStart    string // Peak hours start time (HH:MM)
	PeakEnd      string // Peak hours end time (HH:MM)
	Verbose      bool   // Show full descriptions
	ShowDuration bool   // Show duration column
	ShowPeak     bool   // Show peak indicator column
	MaxDescWidth int    // Maximum description width (0 = auto)
}

// HasPeakHours returns true if peak hours are configured.
func (o PrintOpts) HasPeakHours() bool {
	return o.PeakStart != "" && o.PeakEnd != ""
}

// CalcMaxDescWidth calculates the maximum description width based on options.
func (o PrintOpts) CalcMaxDescWidth(defaultWidth int) int {
	if o.MaxDescWidth > 0 {
		return o.MaxDescWidth
	}
	if !o.Verbose {
		return defaultWidth
	}
	tw := termWidth()
	// Base: "  ⚡ ○  HH:MM-HH:MM  [D]  " = ~26 chars
	// Duration suffix: "  Xh" = ~6 chars
	overhead := 26
	if o.ShowDuration {
		overhead += 6
	}
	available := tw - overhead
	if available > defaultWidth {
		return available
	}
	return defaultWidth
}

// PrintTaskRow prints a single task row with consistent formatting.
func PrintTaskRow(t *task.Task, opts PrintOpts, maxDescWidth int) {
	symbol := statusSymbol(t.Status)

	// Format category
	var catFormatted string
	if t.Category == task.CategoryDeep {
		catFormatted = formatDeep("[D]")
	} else {
		catFormatted = formatShallow("[S]")
	}

	// Peak indicator (⚡ is 2 columns wide)
	var peakIndicator string
	if opts.ShowPeak && opts.HasPeakHours() {
		if OverlapMinutes(t.ScheduledStart, t.ScheduledEnd, opts.PeakStart, opts.PeakEnd) > 0 {
			peakIndicator = "⚡ "
		} else {
			peakIndicator = "   "
		}
	}

	// Truncate description
	desc := t.Description
	if len(desc) > maxDescWidth {
		desc = desc[:maxDescWidth-3] + "..."
	}

	// Build format string based on options
	if opts.ShowDuration {
		duration := formatMuted(FormatDuration(TaskDurationMinutes(t)))
		if opts.ShowPeak {
			fmt.Printf("  %s%s  %s-%s  %s  %-*s  %s\n",
				peakIndicator, symbol, t.ScheduledStart, t.ScheduledEnd,
				catFormatted, maxDescWidth, desc, duration)
		} else {
			fmt.Printf("    %s  %s-%s  %s  %-*s  %s\n",
				symbol, t.ScheduledStart, t.ScheduledEnd,
				catFormatted, maxDescWidth, desc, duration)
		}
	} else {
		if opts.ShowPeak {
			fmt.Printf("  %s%s  %s-%s  %s  %s\n",
				peakIndicator, symbol, t.ScheduledStart, t.ScheduledEnd,
				catFormatted, desc)
		} else {
			fmt.Printf("    %s  %s-%s  %s  %s\n",
				symbol, t.ScheduledStart, t.ScheduledEnd,
				catFormatted, desc)
		}
	}
}

// AccumulateStats updates stats based on a task.
func AccumulateStats(stats *Stats, t *task.Task, dayKey string, opts PrintOpts) {
	minutes := TaskDurationMinutes(t)
	stats.TotalBlocks++

	if stats.DayStats == nil {
		stats.DayStats = make(map[string]DayStats)
	}
	ds := stats.DayStats[dayKey]
	ds.Blocks++

	switch t.Status {
	case task.StatusCancelled:
		stats.CancelledBlocks++
	case task.StatusPostponed:
		stats.PostponedBlocks++
	default:
		if t.Category == task.CategoryDeep {
			stats.DeepMinutes += minutes
			ds.DeepMinutes += minutes
			if opts.HasPeakHours() {
				stats.PeakDeepMinutes += OverlapMinutes(
					t.ScheduledStart, t.ScheduledEnd,
					opts.PeakStart, opts.PeakEnd)
			}
		} else {
			stats.ShallowMinutes += minutes
			ds.ShallowMinutes += minutes
		}
	}
	stats.DayStats[dayKey] = ds
}

// PrintStats prints the stats summary line.
func PrintStats(stats Stats, showPeakAlignment bool) {
	deepStr := formatDeep(fmt.Sprintf("Deep: %s", FormatDuration(stats.DeepMinutes)))
	shallowStr := formatShallow(fmt.Sprintf("Shallow: %s", FormatDuration(stats.ShallowMinutes)))
	fmt.Printf("%s | %s | Total: %d blocks\n", deepStr, shallowStr, stats.TotalBlocks)

	if showPeakAlignment && stats.DeepMinutes > 0 {
		fmt.Printf("Peak alignment: %s (%s of %s deep during peak hours)\n",
			formatStats(fmt.Sprintf("%d%%", stats.PeakPercent())),
			FormatDuration(stats.PeakDeepMinutes),
			FormatDuration(stats.DeepMinutes))
	}
}

// PrintStatsExtended prints extended stats (for week view).
func PrintStatsExtended(stats task.WeekStats, showPeakAlignment bool) {
	deepStr := formatDeep(fmt.Sprintf("Deep: %s (%d%%)", FormatDuration(stats.DeepMinutes), stats.DeepPercent()))
	shallowStr := formatShallow(fmt.Sprintf("Shallow: %s", FormatDuration(stats.ShallowMinutes)))

	fmt.Printf("  %s  |  %s  |  Ratio: %s  |  Blocks: %d\n",
		deepStr, shallowStr, stats.Ratio(), stats.TotalBlocks)

	if bestDay, bestDeep := stats.BestDay(); bestDay >= 0 {
		dayName := task.WeekdayName(bestDay)
		fmt.Printf("  Best day: %s (%s deep)\n", dayName, formatStats(FormatDuration(bestDeep)))
	}

	if showPeakAlignment && stats.DeepMinutes > 0 {
		fmt.Printf("  Peak alignment: %s (%s of %s deep during peak hours)\n",
			formatStats(fmt.Sprintf("%d%%", stats.PeakPercent())),
			FormatDuration(stats.PeakDeepMinutes),
			FormatDuration(stats.DeepMinutes))
	}

	if stats.CancelledBlocks > 0 || stats.PostponedBlocks > 0 {
		fmt.Printf("  %s\n", formatMuted(fmt.Sprintf("Cancelled: %d  |  Postponed: %d",
			stats.CancelledBlocks, stats.PostponedBlocks)))
	}
}

// FlowBar creates an ASCII progress bar showing deep work percentage.
func FlowBar(deepMinutes, totalMinutes, width int) string {
	if totalMinutes == 0 {
		return "[" + strings.Repeat("░", width) + "] (0% Focused)"
	}

	pct := (deepMinutes * 100) / totalMinutes
	filled := (deepMinutes * width) / totalMinutes

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("[%s] %s", formatDeep(bar), formatStats(fmt.Sprintf("(%d%% Focused)", pct)))
}

// FormatDuration formats minutes as a human-readable duration.
func FormatDuration(minutes int) string {
	if minutes == 0 {
		return "0m"
	}
	hours := minutes / 60
	mins := minutes % 60
	if hours == 0 {
		return fmt.Sprintf("%dm", mins)
	}
	if mins == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh%dm", hours, mins)
}

// TaskDurationMinutes calculates the duration of a task in minutes.
//
// Deprecated: Use t.Duration() instead.
func TaskDurationMinutes(t *task.Task) int {
	return t.Duration()
}

// OverlapMinutes calculates the overlapping minutes between two time ranges.
// All times are in "HH:MM" format.
func OverlapMinutes(start1, end1, start2, end2 string) int {
	return task.OverlapMinutes(start1, end1, start2, end2)
}

// TimeToMinutes converts "HH:MM" to minutes since midnight.
func TimeToMinutes(t string) int {
	return task.TimeToMinutes(t)
}

// statusSymbol returns the status indicator for a task.
// Defined in list.go - reused here for consistency.

// PrintInsightWrapped formats and prints insight text preserving structure.
func PrintInsightWrapped(text string, width int) {
	// Strip markdown code blocks
	text = stripMarkdownCodeBlocks(text)

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			fmt.Println()
			continue
		}

		// Detect and format special line types
		prefix, content, contentWidth, skip := parseInsightLine(trimmed, width)
		if skip {
			fmt.Println()
			fmt.Println(formatHeader("  " + content))
			continue
		}

		wrapAndPrint(content, prefix, contentWidth)
	}
}

// parseInsightLine parses a line and returns formatting info.
// Returns: prefix, content, contentWidth, isHeader
func parseInsightLine(trimmed string, width int) (prefix, content string, contentWidth int, isHeader bool) {
	prefix = "  "
	content = trimmed
	contentWidth = width - 2

	switch {
	case strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* "):
		// Bullet point
		prefix = "    • "
		content = strings.TrimPrefix(strings.TrimPrefix(trimmed, "- "), "* ")
		contentWidth = width - 6

	case strings.HasPrefix(trimmed, "#"):
		// Header - strip # and signal to print as header
		content = strings.TrimLeft(trimmed, "# ")
		isHeader = true

	case strings.HasPrefix(trimmed, ">"):
		// Blockquote
		content = strings.TrimPrefix(trimmed, "> ")
		prefix = "  │ "
		contentWidth = width - 4

	case isNumberedItem(trimmed):
		// Numbered item (1. or 10.)
		idx := strings.Index(trimmed, ".")
		prefix = "  " + trimmed[:idx+1] + " "
		content = strings.TrimSpace(trimmed[idx+1:])
		contentWidth = width - len(prefix)
	}

	return prefix, content, contentWidth, isHeader
}

// isNumberedItem checks if a line starts with a number followed by a period.
func isNumberedItem(s string) bool {
	if len(s) < 3 {
		return false
	}
	if s[0] < '1' || s[0] > '9' {
		return false
	}
	if s[1] == '.' {
		return true
	}
	if s[1] >= '0' && s[1] <= '9' && len(s) > 3 && s[2] == '.' {
		return true
	}
	return false
}

// wrapAndPrint wraps text to width and prints with the given prefix.
func wrapAndPrint(text, prefix string, width int) {
	words := strings.Fields(text)
	if len(words) == 0 {
		return
	}

	line := ""
	continuationPrefix := strings.Repeat(" ", len(prefix))
	isFirstLine := true

	for _, word := range words {
		switch {
		case line == "":
			line = word
		case len(line)+1+len(word) <= width:
			line += " " + word
		default:
			// Print current line and start new one
			printLine(prefix, continuationPrefix, line, isFirstLine)
			isFirstLine = false
			line = word
		}
	}

	if line != "" {
		printLine(prefix, continuationPrefix, line, isFirstLine)
	}
}

func printLine(prefix, continuationPrefix, line string, isFirstLine bool) {
	if isFirstLine {
		fmt.Println(formatInsight(prefix + line))
	} else {
		fmt.Println(formatInsight(continuationPrefix + line))
	}
}

// stripMarkdownCodeBlocks removes ```...``` fences from text.
func stripMarkdownCodeBlocks(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	inCodeBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			continue // Skip the fence line
		}
		if !inCodeBlock {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

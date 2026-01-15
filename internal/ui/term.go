package ui

import (
	"os"

	"github.com/fatih/color"
	"golang.org/x/term"
)

// Color definitions for consistent styling across the UI.
var (
	// Deep work: bold cyan for focus/calm
	colorDeep = color.New(color.FgCyan, color.Bold)

	// Shallow work: dim/grey for administrative
	colorShallow = color.New(color.FgWhite, color.Faint)

	// Insight/results: yellow to make it pop
	colorInsight = color.New(color.FgYellow)

	// Headers: bold
	colorHeader = color.New(color.Bold)

	// Stats: green for positive metrics
	colorStats = color.New(color.FgGreen)

	// Muted: for secondary information
	colorMuted = color.New(color.FgWhite, color.Faint)
)

// termWidth returns the terminal width, or a default if detection fails.
func termWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 80 // sensible default
	}
	return width
}

// DisableColor disables all color output.
func DisableColor() {
	color.NoColor = true
}

// EnableColor enables color output (if terminal supports it).
func EnableColor() {
	color.NoColor = false
}

// formatDeep formats text for deep work category.
func formatDeep(s string) string {
	return colorDeep.Sprint(s)
}

// formatShallow formats text for shallow work category.
func formatShallow(s string) string {
	return colorShallow.Sprint(s)
}

// formatInsight formats text for insight/coaching output.
func formatInsight(s string) string {
	return colorInsight.Sprint(s)
}

// formatHeader formats text as a header.
func formatHeader(s string) string {
	return colorHeader.Sprint(s)
}

// formatStats formats text for statistics.
func formatStats(s string) string {
	return colorStats.Sprint(s)
}

// formatMuted formats text as secondary/muted.
func formatMuted(s string) string {
	return colorMuted.Sprint(s)
}

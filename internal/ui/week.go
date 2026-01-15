package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/javiermolinar/sancho/internal/summary"
	"github.com/javiermolinar/sancho/internal/task"
)

func (a *App) weekCmd() *cobra.Command {
	var model string
	var noInsight bool
	var verbose bool
	var noColor bool

	cmd := &cobra.Command{
		Use:   "week",
		Short: "Show and evaluate this week's time blocks",
		Long: `Display this week's scheduled time blocks with stats and insights.

Shows Monday through Sunday of the current ISO week in a table format,
calculates deep/shallow work stats, and optionally provides LLM coaching.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			if noColor {
				DisableColor()
			}

			ctx := context.Background()
			if model == "" {
				model = a.config.LLM.Model
			}

			weekSummary, err := summary.BuildWeekSummary(ctx, a.repo, summary.BuildWeekSummaryOptions{
				WeekStart:      time.Now(),
				PeakStart:      a.config.Schedule.PeakHoursStart,
				PeakEnd:        a.config.Schedule.PeakHoursEnd,
				IncludeInsight: !noInsight,
				Provider:       a.config.LLM.Provider,
				Model:          model,
				BaseURL:        a.config.LLM.BaseURL,
			})
			if err != nil {
				return fmt.Errorf("building week summary: %w", err)
			}

			if len(weekSummary.Tasks) == 0 {
				fmt.Println("No time blocks scheduled for this week.")
				return nil
			}

			// Print header
			header := fmt.Sprintf("WEEK: %s - %s", weekSummary.Start.Format("Mon Jan 2"), weekSummary.End.Format("Mon Jan 2, 2006"))
			fmt.Printf("\n  %s\n", formatHeader(header))
			fmt.Println(strings.Repeat("─", 74))

			// Configure print options
			opts := PrintOpts{
				PeakStart:    a.config.Schedule.PeakHoursStart,
				PeakEnd:      a.config.Schedule.PeakHoursEnd,
				Verbose:      verbose,
				ShowDuration: true,
				ShowPeak:     a.config.HasPeakHours(),
			}
			maxDescWidth := opts.CalcMaxDescWidth(40)

			// Print tasks grouped by day
			printWeekTable(weekSummary.Tasks, opts, maxDescWidth)

			// Print stats
			fmt.Println(strings.Repeat("─", 74))
			PrintStatsExtended(weekSummary.Stats, a.config.HasPeakHours())

			// Show flow bar
			if weekSummary.Stats.TotalMinutes() > 0 {
				fmt.Printf("  Flow: %s\n", FlowBar(weekSummary.Stats.DeepMinutes, weekSummary.Stats.TotalMinutes(), 20))
			}

			// Get LLM insight if not disabled
			if !noInsight && weekSummary.Insight != "" {
				fmt.Println()
				fmt.Printf("  %s\n", formatHeader("INSIGHT"))
				fmt.Println(strings.Repeat("─", 74))
				PrintInsightWrapped(weekSummary.Insight, 72)
			}

			fmt.Println()
			return nil
		},
	}

	cmd.Flags().StringVar(&model, "model", "", "LLM model to use (default from config)")
	cmd.Flags().BoolVar(&noInsight, "no-insight", false, "Skip LLM insight")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show full task descriptions")
	cmd.Flags().BoolVar(&noColor, "no-color", false, "Disable color output")
	return cmd
}

func printWeekTable(tasks []*task.Task, opts PrintOpts, maxDescWidth int) {
	var currentDate string
	for _, t := range tasks {
		date := t.ScheduledDate.Format("2006-01-02")
		dayName := t.ScheduledDate.Format("Mon Jan 2")

		// Print day header if new day
		if date != currentDate {
			if currentDate != "" {
				fmt.Println()
			}
			fmt.Printf("  %s\n", formatHeader(dayName))
			currentDate = date
		}

		// Print task row and accumulate stats
		PrintTaskRow(t, opts, maxDescWidth)
	}
}

package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/javiermolinar/sancho/internal/dateutil"
)

func (a *App) showCmd() *cobra.Command {
	var verbose bool
	var noColor bool

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show today's time blocks",
		Long: `Display today's scheduled time blocks in a simple format.

This is a quick view without LLM evaluation. Use 'sancho week' for
weekly stats and insights.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			if noColor {
				DisableColor()
			}

			ctx := context.Background()
			today := dateutil.TruncateToDay(time.Now())

			tasks, err := a.repo.ListTasksByDateRange(ctx, today, today)
			if err != nil {
				return fmt.Errorf("fetching tasks: %w", err)
			}

			if len(tasks) == 0 {
				fmt.Println("No time blocks scheduled for today.")
				return nil
			}

			fmt.Printf("=== %s ===\n\n", formatHeader(today.Format("Monday, January 2, 2006")))

			// Configure print options
			opts := PrintOpts{
				PeakStart: a.config.Schedule.PeakHoursStart,
				PeakEnd:   a.config.Schedule.PeakHoursEnd,
				Verbose:   verbose,
				ShowPeak:  a.config.HasPeakHours(),
			}
			maxDescWidth := opts.CalcMaxDescWidth(50)

			// Print tasks and accumulate stats
			var stats Stats
			dayKey := today.Format("Mon Jan 2")
			for _, t := range tasks {
				PrintTaskRow(t, opts, maxDescWidth)
				AccumulateStats(&stats, t, dayKey, opts)
			}

			// Print stats
			fmt.Println()
			PrintStats(stats, opts.HasPeakHours())

			// Show flow bar
			if stats.TotalMinutes() > 0 {
				fmt.Printf("Flow: %s\n", FlowBar(stats.DeepMinutes, stats.TotalMinutes(), 20))
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show full task descriptions")
	cmd.Flags().BoolVar(&noColor, "no-color", false, "Disable color output")
	return cmd
}

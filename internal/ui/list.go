package ui

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/javiermolinar/sancho/internal/dateutil"
	"github.com/javiermolinar/sancho/internal/task"
)

func (a *App) listCmd() *cobra.Command {
	var (
		startDate string
		endDate   string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks in a date range",
		Long: `List all tasks scheduled within a date range.

If no dates are specified, lists today's tasks.
If only --start is specified, lists tasks for that single day.
If both --start and --end are specified, lists tasks in that range (inclusive).`,
		Example: `  sancho list
  sancho list --start=2025-01-15
  sancho list --start=2025-01-15 --end=2025-01-20`,
		RunE: func(_ *cobra.Command, _ []string) error {
			dateRange, err := dateutil.NewDateRange(startDate, endDate)
			if err != nil {
				return err
			}

			tasks, err := a.repo.ListTasksByDateRange(context.Background(), dateRange.Start, dateRange.End)
			if err != nil {
				return fmt.Errorf("listing tasks: %w", err)
			}

			if len(tasks) == 0 {
				fmt.Println("No tasks found in the specified date range.")
				return nil
			}

			// Print tasks grouped by date
			var currentDate string
			for _, t := range tasks {
				date := t.ScheduledDate.Format("2006-01-02")
				if date != currentDate {
					if currentDate != "" {
						fmt.Println()
					}
					fmt.Printf("=== %s ===\n", date)
					currentDate = date
				}

				status := statusSymbol(t.Status)
				category := string(t.Category)[0:1] // "d" or "s"
				fmt.Printf("  %s #%d [%s] %s-%s %s\n",
					status,
					t.ID,
					category,
					t.ScheduledStart,
					t.ScheduledEnd,
					t.Description,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&startDate, "start", "", "Start date (YYYY-MM-DD, defaults to today)")
	cmd.Flags().StringVar(&endDate, "end", "", "End date (YYYY-MM-DD, defaults to start date)")

	return cmd
}

func statusSymbol(s task.Status) string {
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

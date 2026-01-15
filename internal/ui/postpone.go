package ui

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/javiermolinar/sancho/internal/dateutil"
)

func (a *App) postponeCmd() *cobra.Command {
	var (
		date  string
		start string
		end   string
	)

	cmd := &cobra.Command{
		Use:   "postpone <task-id>",
		Short: "Postpone a task to a new date/time",
		Long: `Postpone a task to a new date and time.

This marks the original task as postponed and creates a new task
with the same description and category at the specified date/time.
The new task will have a reference to the original task.`,
		Example: `  sancho postpone 123 --date=2025-01-16 --start=14:00 --end=16:00
  sancho postpone 123 --start=09:00 --end=11:00  # defaults to today`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := a.ensureRepo(); err != nil {
				return err
			}

			taskID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid task ID: %w", err)
			}

			newDate, err := dateutil.ParseDate(date)
			if err != nil {
				return fmt.Errorf("invalid date: %w", err)
			}

			// Validate time format
			if start == "" || end == "" {
				return fmt.Errorf("--start and --end are required")
			}

			newTask, err := a.repo.PostponeTask(context.Background(), taskID, newDate, start, end)
			if err != nil {
				return err
			}

			fmt.Printf("Postponed task #%d → #%d: %s [%s] %s %s-%s\n",
				taskID,
				newTask.ID,
				newTask.Description,
				newTask.Category,
				newTask.ScheduledDate.Format("2006-01-02"),
				newTask.ScheduledStart,
				newTask.ScheduledEnd,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&date, "date", "", "New date (YYYY-MM-DD, defaults to today)")
	cmd.Flags().StringVar(&start, "start", "", "New start time (HH:MM, required)")
	cmd.Flags().StringVar(&end, "end", "", "New end time (HH:MM, required)")

	_ = cmd.MarkFlagRequired("start")
	_ = cmd.MarkFlagRequired("end")

	return cmd
}

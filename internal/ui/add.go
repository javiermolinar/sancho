package ui

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/javiermolinar/sancho/internal/task"
)

func (a *App) addCmd() *cobra.Command {
	var (
		date     string
		start    string
		end      string
		category string
	)

	cmd := &cobra.Command{
		Use:   "add [description]",
		Short: "Add a new task",
		Long: `Add a new task to your schedule.

Example:
  sancho add "Write documentation" --date=2025-01-10 --start=09:00 --end=11:00 --category=deep`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			t, err := task.New(args[0], category, date, start, end)
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := a.repo.CreateTask(ctx, t); err != nil {
				return fmt.Errorf("creating task: %w", err)
			}

			fmt.Printf("Created task #%d: %s [%s] %s %s-%s\n",
				t.ID,
				t.Description,
				t.Category,
				t.ScheduledDate.Format("2006-01-02"),
				t.ScheduledStart,
				t.ScheduledEnd,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&date, "date", "", "Scheduled date (YYYY-MM-DD, default: today)")
	cmd.Flags().StringVar(&start, "start", "", "Start time (HH:MM, required)")
	cmd.Flags().StringVar(&end, "end", "", "End time (HH:MM, required)")
	cmd.Flags().StringVar(&category, "category", "deep", "Category: deep or shallow")

	_ = cmd.MarkFlagRequired("start")
	_ = cmd.MarkFlagRequired("end")

	return cmd
}

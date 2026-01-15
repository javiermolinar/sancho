package ui

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/javiermolinar/sancho/internal/task"
)

func (a *App) outcomeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "outcome [task-id] [on_time|over|under]",
		Short: "Set the outcome of a completed task",
		Long: `Set how the task went during review.

Outcomes:
  on_time - Task was completed within the scheduled time
  over    - Task took longer than scheduled
  under   - Task was completed faster than scheduled

Example:
  sancho outcome 42 on_time`,
		Args: cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := a.ensureRepo(); err != nil {
				return err
			}

			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid task ID: %w", err)
			}

			outcome := task.Outcome(args[1])
			if !outcome.Valid() {
				return fmt.Errorf("invalid outcome %q: must be on_time, over, or under", args[1])
			}

			ctx := context.Background()
			if err := a.repo.SetTaskOutcome(ctx, id, outcome); err != nil {
				return fmt.Errorf("setting outcome: %w", err)
			}

			fmt.Printf("Set outcome for task #%d: %s\n", id, outcome)
			return nil
		},
	}
}

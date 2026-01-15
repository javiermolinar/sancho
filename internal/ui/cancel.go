package ui

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func (a *App) cancelCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cancel [task-id]",
		Short: "Cancel a scheduled task",
		Long: `Cancel a task by its ID.

Example:
  sancho cancel 42`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := a.ensureRepo(); err != nil {
				return err
			}

			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid task ID: %w", err)
			}

			ctx := context.Background()
			if err := a.repo.CancelTask(ctx, id); err != nil {
				return fmt.Errorf("cancelling task: %w", err)
			}

			fmt.Printf("Cancelled task #%d\n", id)
			return nil
		},
	}
}

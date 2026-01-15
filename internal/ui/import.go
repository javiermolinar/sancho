package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/javiermolinar/sancho/internal/db"
	"github.com/javiermolinar/sancho/internal/task"
)

func (a *App) importCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import [database_path]",
		Short: "Import tasks from another database",
		Long: `Import all tasks from another Sancho database into the current one.

Example:
  sancho import /path/to/other.db`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := a.ensureRepo(); err != nil {
				return err
			}

			sourcePath, err := resolvePath(args[0])
			if err != nil {
				return err
			}
			destPath, err := resolvePath(a.config.Storage.DBPath)
			if err != nil {
				return err
			}

			if sourcePath == destPath {
				return fmt.Errorf("source database matches current database")
			}

			info, err := os.Stat(sourcePath)
			if err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("source database does not exist: %s", sourcePath)
				}
				return fmt.Errorf("checking source database: %w", err)
			}
			if info.IsDir() {
				return fmt.Errorf("source database path is a directory: %s", sourcePath)
			}

			count, err := importTasks(context.Background(), a.repo, sourcePath)
			if err != nil {
				return err
			}

			fmt.Printf("Imported %d tasks from %s\n", count, sourcePath)
			return nil
		},
	}

	return cmd
}

func importTasks(ctx context.Context, dest task.Repository, sourcePath string) (int, error) {
	sourceRepo, err := db.New(sourcePath)
	if err != nil {
		return 0, fmt.Errorf("opening source database: %w", err)
	}
	defer func() { _ = sourceRepo.Close() }()

	tasks, err := sourceRepo.ListAllTasks(ctx)
	if err != nil {
		return 0, fmt.Errorf("listing source tasks: %w", err)
	}

	imported := 0
	idMap := make(map[int64]int64, len(tasks))
	for _, sourceTask := range tasks {
		newTask := &task.Task{
			Description:    sourceTask.Description,
			Category:       sourceTask.Category,
			ScheduledDate:  sourceTask.ScheduledDate,
			ScheduledStart: sourceTask.ScheduledStart,
			ScheduledEnd:   sourceTask.ScheduledEnd,
			Status:         sourceTask.Status,
			Outcome:        sourceTask.Outcome,
			CreatedAt:      sourceTask.CreatedAt,
		}

		if sourceTask.PostponedFrom != nil {
			newID, ok := idMap[*sourceTask.PostponedFrom]
			if !ok {
				return imported, fmt.Errorf("postponed from task %d not imported yet", *sourceTask.PostponedFrom)
			}
			newTask.PostponedFrom = &newID
		}

		if err := dest.CreateTask(ctx, newTask); err != nil {
			return imported, fmt.Errorf("importing task %q: %w", sourceTask.Description, err)
		}

		idMap[sourceTask.ID] = newTask.ID
		imported++
	}

	return imported, nil
}

func resolvePath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("empty path")
	}

	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolving home directory: %w", err)
		}
		path = filepath.Join(home, path[2:])
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolving path: %w", err)
	}

	return absPath, nil
}

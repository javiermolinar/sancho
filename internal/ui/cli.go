package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/javiermolinar/sancho/internal/config"
	"github.com/javiermolinar/sancho/internal/db"
	"github.com/javiermolinar/sancho/internal/task"
	"github.com/javiermolinar/sancho/internal/tui"
)

var (
	// Version is set at build time
	Version = "dev"
	// Commit is set at build time
	Commit = "none"
)

// App holds the CLI application state.
type App struct {
	repo   task.Repository
	config *config.Config
	root   *cobra.Command
	debug  bool // Enable debug logging
}

// NewApp creates a new CLI application with the given repository and config.
func NewApp(repo task.Repository, cfg *config.Config) *App {
	a := &App{repo: repo, config: cfg}

	a.root = &cobra.Command{
		Use:   "sancho",
		Short: "A CLI tool for deep work scheduling",
		Long: `Sancho is a CLI tool implementing Cal Newport's deep work methodology.

It helps you plan your day with focused work blocks, manage tasks,
and track your productivity over time.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return tui.RunWithDebug(a.repo, a.config, a.debug)
		},
	}

	// Add global flags
	a.root.PersistentFlags().BoolVar(&a.debug, "debug", false, "Enable debug logging (logs to temp file)")

	a.root.AddCommand(a.versionCmd())
	a.root.AddCommand(a.configCmd())
	a.root.AddCommand(a.addCmd())
	a.root.AddCommand(a.cancelCmd())
	a.root.AddCommand(a.outcomeCmd())
	a.root.AddCommand(a.listCmd())
	a.root.AddCommand(a.postponeCmd())
	a.root.AddCommand(a.planCmd())
	a.root.AddCommand(a.weekCmd())
	a.root.AddCommand(a.showCmd())
	a.root.AddCommand(a.importCmd())

	return a
}

func (a *App) versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("sancho %s (commit: %s)\n", Version, Commit)
		},
	}
}

// Execute runs the CLI application.
func (a *App) Execute() error {
	return a.root.Execute()
}

// Close releases any resources held by the app.
func (a *App) Close() error {
	if a.repo == nil {
		return nil
	}
	err := a.repo.Close()
	a.repo = nil
	return err
}

func (a *App) ensureRepo() error {
	if a.repo != nil {
		return nil
	}
	dbDir := filepath.Dir(a.config.Storage.DBPath)
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		return fmt.Errorf("creating data directory: %w", err)
	}
	repo, err := db.New(a.config.Storage.DBPath)
	if err != nil {
		return fmt.Errorf("initializing database: %w", err)
	}
	a.repo = repo
	return nil
}

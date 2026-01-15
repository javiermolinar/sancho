package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/javiermolinar/sancho/internal/config"
	"github.com/javiermolinar/sancho/internal/db"
	"github.com/javiermolinar/sancho/internal/ui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Ensure database directory exists
	dbDir := filepath.Dir(cfg.Storage.DBPath)
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		return fmt.Errorf("creating data directory: %w", err)
	}

	// Initialize repository
	repo, err := db.New(cfg.Storage.DBPath)
	if err != nil {
		return fmt.Errorf("initializing database: %w", err)
	}
	defer func() { _ = repo.Close() }()

	app := ui.NewApp(repo, cfg)
	return app.Execute()
}

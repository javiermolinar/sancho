// Package tui provides the terminal user interface for sancho.
package tui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/javiermolinar/sancho/internal/config"
	"github.com/javiermolinar/sancho/internal/db"
	"github.com/javiermolinar/sancho/internal/task"
)

// InitState tracks whether startup initialization is required.
type InitState struct {
	NeedsInit     bool
	ConfigMissing bool
	DBMissing     bool
	ConfigPath    string
	DBPath        string
}

// DetectInitState checks for missing config or database files.
func DetectInitState(cfg *config.Config) (InitState, error) {
	state := InitState{
		ConfigPath: config.DefaultConfigPath(),
		DBPath:     cfg.Storage.DBPath,
	}

	configMissing, err := pathMissing(state.ConfigPath)
	if err != nil {
		return InitState{}, fmt.Errorf("checking config path: %w", err)
	}
	dbMissing, err := pathMissing(state.DBPath)
	if err != nil {
		return InitState{}, fmt.Errorf("checking db path: %w", err)
	}

	state.ConfigMissing = configMissing
	state.DBMissing = dbMissing
	state.NeedsInit = configMissing || dbMissing
	return state, nil
}

func pathMissing(path string) (bool, error) {
	if path == "" {
		return true, nil
	}
	_, err := os.Stat(path)
	if err == nil {
		return false, nil
	}
	if os.IsNotExist(err) {
		return true, nil
	}
	return false, err
}

func openRepo(dbPath string) (task.Repository, error) {
	if dbPath == "" {
		return nil, fmt.Errorf("db path is empty")
	}
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating data directory: %w", err)
	}
	repo, err := db.New(dbPath)
	if err != nil {
		return nil, fmt.Errorf("initializing database: %w", err)
	}
	return repo, nil
}

func (m Model) initializeStorage() (Model, error) {
	if m.initState.ConfigMissing {
		if err := m.config.SaveTo(m.initState.ConfigPath); err != nil {
			return m, fmt.Errorf("saving config: %w", err)
		}
	}

	if m.repo == nil {
		repo, err := openRepo(m.initState.DBPath)
		if err != nil {
			return m, err
		}
		m.repo = repo
	}

	return m, nil
}

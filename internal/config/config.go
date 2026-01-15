// Package config handles configuration loading from files, defaults, and environment variables.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Config holds the application configuration.
type Config struct {
	Schedule ScheduleConfig `toml:"schedule"`
	LLM      LLMConfig      `toml:"llm"`
	Storage  StorageConfig  `toml:"storage"`
	UI       UIConfig       `toml:"ui"`
}

// UIConfig holds TUI settings.
type UIConfig struct {
	Theme string `toml:"theme"` // "mocha", "macchiato", "frappe", "latte"
}

// ScheduleConfig holds workday scheduling settings.
type ScheduleConfig struct {
	Workdays       []string `toml:"workdays"`         // e.g., ["monday", "tuesday", ...]
	DayStart       string   `toml:"day_start"`        // e.g., "09:00"
	DayEnd         string   `toml:"day_end"`          // e.g., "17:00"
	PeakHoursStart string   `toml:"peak_hours_start"` // e.g., "09:00" (optional)
	PeakHoursEnd   string   `toml:"peak_hours_end"`   // e.g., "12:00" (optional)
}

// LLMConfig holds LLM provider settings.
type LLMConfig struct {
	Provider string `toml:"provider"` // "copilot", "ollama", etc.
	Model    string `toml:"model"`    // e.g., "gpt-4o"
	BaseURL  string `toml:"base_url"` // e.g., "http://localhost:11434"
}

// StorageConfig holds database settings.
type StorageConfig struct {
	DBPath string `toml:"db_path"`
}

// Default returns the default configuration.
func Default() *Config {
	return &Config{
		Schedule: ScheduleConfig{
			Workdays:       []string{"monday", "tuesday", "wednesday", "thursday", "friday"},
			DayStart:       "09:00",
			DayEnd:         "17:00",
			PeakHoursStart: "", // Empty means no peak hours configured
			PeakHoursEnd:   "",
		},
		LLM: LLMConfig{
			Provider: "copilot",
			Model:    "gpt-4o",
			BaseURL:  "http://localhost:11434",
		},
		Storage: StorageConfig{
			DBPath: defaultDBPath(),
		},
		UI: UIConfig{
			Theme: "frappe", // Default to Catppuccin Mocha
		},
	}
}

// defaultDBPath returns the default database path.
func defaultDBPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "sancho.db"
	}
	return filepath.Join(home, ".local", "share", "sancho", "sancho.db")
}

// DefaultConfigPath returns the default config file path.
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "config.toml"
	}
	return filepath.Join(home, ".config", "sancho", "config.toml")
}

// Load loads configuration from the default path, merging with defaults and env vars.
func Load() (*Config, error) {
	return LoadFrom(DefaultConfigPath())
}

// LoadFrom loads configuration from the specified path.
// It starts with defaults, overlays file config if it exists, then applies env overrides.
func LoadFrom(path string) (*Config, error) {
	cfg := Default()

	// Try to load from file (not an error if it doesn't exist)
	if err := loadFromFile(path, cfg); err != nil {
		return nil, err
	}

	// Apply environment variable overrides
	applyEnvOverrides(cfg)

	// Expand paths
	cfg.Storage.DBPath = expandPath(cfg.Storage.DBPath)

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// loadFromFile loads config from a file if it exists.
func loadFromFile(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, use defaults
		}
		return fmt.Errorf("reading config file: %w", err)
	}

	if err := toml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("parsing config file: %w", err)
	}

	return nil
}

// applyEnvOverrides applies environment variable overrides to the config.
// Environment variables take precedence over file config.
func applyEnvOverrides(cfg *Config) {
	// Schedule overrides
	if v := os.Getenv("DEEPWORK_DAY_START"); v != "" {
		cfg.Schedule.DayStart = v
	}
	if v := os.Getenv("DEEPWORK_DAY_END"); v != "" {
		cfg.Schedule.DayEnd = v
	}
	if v := os.Getenv("DEEPWORK_WORKDAYS"); v != "" {
		cfg.Schedule.Workdays = strings.Split(v, ",")
	}
	if v := os.Getenv("DEEPWORK_PEAK_HOURS_START"); v != "" {
		cfg.Schedule.PeakHoursStart = v
	}
	if v := os.Getenv("DEEPWORK_PEAK_HOURS_END"); v != "" {
		cfg.Schedule.PeakHoursEnd = v
	}

	// LLM overrides
	if v := os.Getenv("DEEPWORK_LLM_PROVIDER"); v != "" {
		cfg.LLM.Provider = v
	}
	if v := os.Getenv("DEEPWORK_LLM_MODEL"); v != "" {
		cfg.LLM.Model = v
	}
	if v := os.Getenv("DEEPWORK_LLM_BASE_URL"); v != "" {
		cfg.LLM.BaseURL = v
	}

	// Storage overrides
	if v := os.Getenv("DEEPWORK_DB_PATH"); v != "" {
		cfg.Storage.DBPath = v
	}

	// UI overrides
	if v := os.Getenv("DEEPWORK_UI_THEME"); v != "" {
		cfg.UI.Theme = v
	}
}

// expandPath expands ~ to the user's home directory.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if err := validateTime(c.Schedule.DayStart, "day_start"); err != nil {
		return err
	}
	if err := validateTime(c.Schedule.DayEnd, "day_end"); err != nil {
		return err
	}
	if c.Schedule.DayStart >= c.Schedule.DayEnd {
		return errors.New("day_start must be before day_end")
	}

	// Validate peak hours if configured (both must be set or neither)
	hasStart := c.Schedule.PeakHoursStart != ""
	hasEnd := c.Schedule.PeakHoursEnd != ""
	if hasStart != hasEnd {
		return errors.New("both peak_hours_start and peak_hours_end must be set, or neither")
	}
	if hasStart && hasEnd {
		if err := validateTime(c.Schedule.PeakHoursStart, "peak_hours_start"); err != nil {
			return err
		}
		if err := validateTime(c.Schedule.PeakHoursEnd, "peak_hours_end"); err != nil {
			return err
		}
		if c.Schedule.PeakHoursStart >= c.Schedule.PeakHoursEnd {
			return errors.New("peak_hours_start must be before peak_hours_end")
		}
	}

	if len(c.Schedule.Workdays) == 0 {
		return errors.New("at least one workday must be configured")
	}
	for _, day := range c.Schedule.Workdays {
		if !isValidWeekday(day) {
			return fmt.Errorf("invalid workday: %s", day)
		}
	}
	if c.Storage.DBPath == "" {
		return errors.New("db_path must be set")
	}
	return nil
}

// validateTime checks if a time string is in HH:MM format.
func validateTime(t, field string) error {
	if len(t) != 5 || t[2] != ':' {
		return fmt.Errorf("%s must be in HH:MM format, got %q", field, t)
	}
	hour := t[0:2]
	min := t[3:5]
	if !isDigits(hour) || !isDigits(min) {
		return fmt.Errorf("%s must be in HH:MM format, got %q", field, t)
	}
	return nil
}

func isDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

var validWeekdays = map[string]bool{
	"monday":    true,
	"tuesday":   true,
	"wednesday": true,
	"thursday":  true,
	"friday":    true,
	"saturday":  true,
	"sunday":    true,
}

func isValidWeekday(day string) bool {
	return validWeekdays[strings.ToLower(day)]
}

// IsWorkday returns true if the given weekday name is a configured workday.
func (c *Config) IsWorkday(weekday string) bool {
	weekday = strings.ToLower(weekday)
	for _, d := range c.Schedule.Workdays {
		if strings.ToLower(d) == weekday {
			return true
		}
	}
	return false
}

// HasPeakHours returns true if peak hours are configured.
func (c *Config) HasPeakHours() bool {
	return c.Schedule.PeakHoursStart != "" && c.Schedule.PeakHoursEnd != ""
}

// Save writes the configuration to the default path.
func (c *Config) Save() error {
	return c.SaveTo(DefaultConfigPath())
}

// SaveTo writes the configuration to the specified path.
func (c *Config) SaveTo(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := toml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

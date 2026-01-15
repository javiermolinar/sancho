package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.Schedule.DayStart != "09:00" {
		t.Errorf("expected day_start 09:00, got %s", cfg.Schedule.DayStart)
	}
	if cfg.Schedule.DayEnd != "17:00" {
		t.Errorf("expected day_end 17:00, got %s", cfg.Schedule.DayEnd)
	}
	if len(cfg.Schedule.Workdays) != 5 {
		t.Errorf("expected 5 workdays, got %d", len(cfg.Schedule.Workdays))
	}
	if cfg.LLM.Provider != "copilot" {
		t.Errorf("expected provider copilot, got %s", cfg.LLM.Provider)
	}
	if cfg.LLM.Model != "gpt-4o" {
		t.Errorf("expected model gpt-4o, got %s", cfg.LLM.Model)
	}
	if cfg.LLM.BaseURL != "http://localhost:11434" {
		t.Errorf("expected base_url http://localhost:11434, got %s", cfg.LLM.BaseURL)
	}
}

func TestLoadFrom_FileNotExists(t *testing.T) {
	cfg, err := LoadFrom("/nonexistent/path/config.toml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return defaults
	if cfg.Schedule.DayStart != "09:00" {
		t.Errorf("expected default day_start, got %s", cfg.Schedule.DayStart)
	}
}

func TestLoadFrom_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `
[schedule]
workdays = ["monday", "tuesday", "wednesday"]
day_start = "08:00"
day_end = "16:00"

[llm]
provider = "openai"
model = "gpt-4o-mini"
base_url = "http://localhost:11435"

[storage]
db_path = "/tmp/test.db"
`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadFrom(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Schedule.DayStart != "08:00" {
		t.Errorf("expected day_start 08:00, got %s", cfg.Schedule.DayStart)
	}
	if cfg.Schedule.DayEnd != "16:00" {
		t.Errorf("expected day_end 16:00, got %s", cfg.Schedule.DayEnd)
	}
	if len(cfg.Schedule.Workdays) != 3 {
		t.Errorf("expected 3 workdays, got %d", len(cfg.Schedule.Workdays))
	}
	if cfg.LLM.Provider != "openai" {
		t.Errorf("expected provider openai, got %s", cfg.LLM.Provider)
	}
	if cfg.LLM.Model != "gpt-4o-mini" {
		t.Errorf("expected model gpt-4o-mini, got %s", cfg.LLM.Model)
	}
	if cfg.LLM.BaseURL != "http://localhost:11435" {
		t.Errorf("expected base_url http://localhost:11435, got %s", cfg.LLM.BaseURL)
	}
	if cfg.Storage.DBPath != "/tmp/test.db" {
		t.Errorf("expected db_path /tmp/test.db, got %s", cfg.Storage.DBPath)
	}
}

func TestLoadFrom_EnvOverrides(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `
[schedule]
day_start = "08:00"
day_end = "16:00"

[storage]
db_path = "/tmp/test.db"
`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Set env vars
	t.Setenv("DEEPWORK_DAY_START", "10:00")
	t.Setenv("DEEPWORK_LLM_MODEL", "gpt-3.5-turbo")
	t.Setenv("DEEPWORK_LLM_BASE_URL", "http://localhost:11436")

	cfg, err := LoadFrom(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Env should override file
	if cfg.Schedule.DayStart != "10:00" {
		t.Errorf("expected day_start 10:00 from env, got %s", cfg.Schedule.DayStart)
	}
	// File value should be kept when no env override
	if cfg.Schedule.DayEnd != "16:00" {
		t.Errorf("expected day_end 16:00 from file, got %s", cfg.Schedule.DayEnd)
	}
	// Env should override default
	if cfg.LLM.Model != "gpt-3.5-turbo" {
		t.Errorf("expected model gpt-3.5-turbo from env, got %s", cfg.LLM.Model)
	}
	if cfg.LLM.BaseURL != "http://localhost:11436" {
		t.Errorf("expected base_url http://localhost:11436 from env, got %s", cfg.LLM.BaseURL)
	}
}

func TestValidate_InvalidDayStart(t *testing.T) {
	cfg := Default()
	cfg.Schedule.DayStart = "9:00" // Missing leading zero

	err := cfg.Validate()
	if err == nil {
		t.Error("expected validation error for invalid day_start")
	}
}

func TestValidate_DayStartAfterDayEnd(t *testing.T) {
	cfg := Default()
	cfg.Schedule.DayStart = "18:00"
	cfg.Schedule.DayEnd = "09:00"

	err := cfg.Validate()
	if err == nil {
		t.Error("expected validation error when day_start >= day_end")
	}
}

func TestValidate_InvalidWorkday(t *testing.T) {
	cfg := Default()
	cfg.Schedule.Workdays = []string{"monday", "funday"}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected validation error for invalid workday")
	}
}

func TestValidate_EmptyWorkdays(t *testing.T) {
	cfg := Default()
	cfg.Schedule.Workdays = []string{}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected validation error for empty workdays")
	}
}

func TestIsWorkday(t *testing.T) {
	cfg := Default()

	tests := []struct {
		day  string
		want bool
	}{
		{"monday", true},
		{"Monday", true},
		{"FRIDAY", true},
		{"saturday", false},
		{"sunday", false},
	}

	for _, tc := range tests {
		t.Run(tc.day, func(t *testing.T) {
			got := cfg.IsWorkday(tc.day)
			if got != tc.want {
				t.Errorf("IsWorkday(%q) = %v, want %v", tc.day, got, tc.want)
			}
		})
	}
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input string
		want  string
	}{
		{"~/test.db", filepath.Join(home, "test.db")},
		{"/absolute/path.db", "/absolute/path.db"},
		{"relative/path.db", "relative/path.db"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := expandPath(tc.input)
			if got != tc.want {
				t.Errorf("expandPath(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	cfg := Default()
	cfg.Schedule.DayStart = "07:30"
	cfg.Schedule.DayEnd = "15:30"
	cfg.Schedule.Workdays = []string{"monday", "tuesday", "wednesday", "thursday"}

	if err := cfg.SaveTo(configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	loaded, err := LoadFrom(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if loaded.Schedule.DayStart != "07:30" {
		t.Errorf("expected day_start 07:30, got %s", loaded.Schedule.DayStart)
	}
	if loaded.Schedule.DayEnd != "15:30" {
		t.Errorf("expected day_end 15:30, got %s", loaded.Schedule.DayEnd)
	}
	if len(loaded.Schedule.Workdays) != 4 {
		t.Errorf("expected 4 workdays, got %d", len(loaded.Schedule.Workdays))
	}
}

func TestValidate_PeakHoursOnlyStartSet(t *testing.T) {
	cfg := Default()
	cfg.Schedule.PeakHoursStart = "09:00"
	// PeakHoursEnd is empty

	err := cfg.Validate()
	if err == nil {
		t.Error("expected validation error when only peak_hours_start is set")
	}
}

func TestValidate_PeakHoursOnlyEndSet(t *testing.T) {
	cfg := Default()
	cfg.Schedule.PeakHoursEnd = "12:00"
	// PeakHoursStart is empty

	err := cfg.Validate()
	if err == nil {
		t.Error("expected validation error when only peak_hours_end is set")
	}
}

func TestValidate_PeakHoursStartAfterEnd(t *testing.T) {
	cfg := Default()
	cfg.Schedule.PeakHoursStart = "14:00"
	cfg.Schedule.PeakHoursEnd = "09:00"

	err := cfg.Validate()
	if err == nil {
		t.Error("expected validation error when peak_hours_start >= peak_hours_end")
	}
}

func TestValidate_PeakHoursInvalidFormat(t *testing.T) {
	cfg := Default()
	cfg.Schedule.PeakHoursStart = "9:00" // Missing leading zero
	cfg.Schedule.PeakHoursEnd = "12:00"

	err := cfg.Validate()
	if err == nil {
		t.Error("expected validation error for invalid peak_hours_start format")
	}
}

func TestValidate_PeakHoursValid(t *testing.T) {
	cfg := Default()
	cfg.Schedule.PeakHoursStart = "09:00"
	cfg.Schedule.PeakHoursEnd = "12:00"

	err := cfg.Validate()
	if err != nil {
		t.Errorf("expected no error for valid peak hours, got: %v", err)
	}
}

func TestHasPeakHours(t *testing.T) {
	cfg := Default()

	if cfg.HasPeakHours() {
		t.Error("expected HasPeakHours() = false for default config")
	}

	cfg.Schedule.PeakHoursStart = "09:00"
	cfg.Schedule.PeakHoursEnd = "12:00"

	if !cfg.HasPeakHours() {
		t.Error("expected HasPeakHours() = true when peak hours configured")
	}
}

func TestLoadFrom_WithPeakHours(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `
[schedule]
workdays = ["monday", "tuesday", "wednesday", "thursday", "friday"]
day_start = "09:00"
day_end = "17:00"
peak_hours_start = "09:00"
peak_hours_end = "12:00"

[storage]
db_path = "/tmp/test.db"
`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadFrom(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Schedule.PeakHoursStart != "09:00" {
		t.Errorf("expected peak_hours_start 09:00, got %q", cfg.Schedule.PeakHoursStart)
	}
	if cfg.Schedule.PeakHoursEnd != "12:00" {
		t.Errorf("expected peak_hours_end 12:00, got %q", cfg.Schedule.PeakHoursEnd)
	}
	if !cfg.HasPeakHours() {
		t.Error("expected HasPeakHours() = true")
	}
}

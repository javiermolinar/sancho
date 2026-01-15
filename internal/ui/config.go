package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/javiermolinar/sancho/internal/config"
	"github.com/javiermolinar/sancho/internal/tui/theme"
)

func (a *App) configCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "View or edit configuration",
		Long: `Interactive configuration management.

If no config file exists, creates one with default values.
Otherwise, displays current config and allows editing.

Example:
  sancho config`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runConfigInteractive()
		},
	}
}

func runConfigInteractive() error {
	configPath := config.DefaultConfigPath()
	fmt.Printf("Config file: %s\n\n", configPath)

	// Load existing config or create defaults
	cfg, err := config.LoadFrom(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Check if file exists
	_, fileErr := os.Stat(configPath)
	isNew := os.IsNotExist(fileErr)

	if isNew {
		fmt.Println("No config file found. Creating with default values...")
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		fmt.Printf("Created %s\n\n", configPath)
	}

	// Display current config
	printConfig(cfg)

	// Ask if user wants to edit
	if !promptYesNo("\nWould you like to edit the configuration?") {
		return nil
	}

	// Interactive editing
	reader := bufio.NewReader(os.Stdin)

	cfg.Schedule.DayStart = promptValue(reader, "Day start", cfg.Schedule.DayStart)
	cfg.Schedule.DayEnd = promptValue(reader, "Day end", cfg.Schedule.DayEnd)
	cfg.Schedule.Workdays = promptSlice(reader, "Workdays (comma-separated)", cfg.Schedule.Workdays)
	cfg.Schedule.PeakHoursStart = promptValue(reader, "Peak hours start (empty to disable)", cfg.Schedule.PeakHoursStart)
	cfg.Schedule.PeakHoursEnd = promptValue(reader, "Peak hours end (empty to disable)", cfg.Schedule.PeakHoursEnd)
	cfg.LLM.Provider = promptValue(reader, "LLM provider", cfg.LLM.Provider)
	cfg.LLM.Model = promptValue(reader, "LLM model", cfg.LLM.Model)
	cfg.LLM.BaseURL = promptValue(reader, "LLM base URL (Ollama/LM Studio)", cfg.LLM.BaseURL)
	cfg.Storage.DBPath = promptValue(reader, "Database path", cfg.Storage.DBPath)
	cfg.UI.Theme = promptTheme(reader, cfg.UI.Theme)

	// Validate before saving
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Save
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Println("\nConfiguration saved!")
	return nil
}

func printConfig(cfg *config.Config) {
	fmt.Println("Current configuration:")
	fmt.Println("──────────────────────")
	fmt.Println("[schedule]")
	fmt.Printf("  day_start        = %s\n", cfg.Schedule.DayStart)
	fmt.Printf("  day_end          = %s\n", cfg.Schedule.DayEnd)
	fmt.Printf("  workdays         = %s\n", strings.Join(cfg.Schedule.Workdays, ", "))
	if cfg.HasPeakHours() {
		fmt.Printf("  peak_hours_start = %s\n", cfg.Schedule.PeakHoursStart)
		fmt.Printf("  peak_hours_end   = %s\n", cfg.Schedule.PeakHoursEnd)
	}
	fmt.Println("\n[llm]")
	fmt.Printf("  provider         = %s\n", cfg.LLM.Provider)
	fmt.Printf("  model            = %s\n", cfg.LLM.Model)
	fmt.Printf("  base_url         = %s\n", cfg.LLM.BaseURL)
	fmt.Println("\n[storage]")
	fmt.Printf("  db_path          = %s\n", cfg.Storage.DBPath)
	fmt.Println("\n[ui]")
	fmt.Printf("  theme            = %s\n", cfg.UI.Theme)
}

func promptYesNo(question string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N]: ", question)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

func promptValue(reader *bufio.Reader, label, current string) string {
	if current == "" {
		fmt.Printf("  %s: ", label)
	} else {
		fmt.Printf("  %s [%s]: ", label, current)
	}
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return current
	}
	return input
}

func promptSlice(reader *bufio.Reader, label string, current []string) []string {
	currentStr := strings.Join(current, ", ")
	fmt.Printf("  %s [%s]: ", label, currentStr)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return current
	}
	parts := strings.Split(input, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func promptTheme(reader *bufio.Reader, current string) string {
	options := strings.Join(theme.Available(), ", ")
	label := fmt.Sprintf("UI theme (%s)", options)
	for {
		value := strings.ToLower(promptValue(reader, label, current))
		if theme.IsAvailable(value) {
			return value
		}
		fmt.Printf("  Invalid theme %q. Available: %s\n", value, options)
	}
}

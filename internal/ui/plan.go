package ui

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/javiermolinar/sancho/internal/dwplanner"
	"github.com/javiermolinar/sancho/internal/llm"
)

const maxRetries = 3

func (a *App) planCmd() *cobra.Command {
	var (
		modelFlag string
		dryRun    bool
	)

	cmd := &cobra.Command{
		Use:   "plan [description]",
		Short: "Plan tasks from natural language input",
		Long: `Use AI to extract and schedule tasks from a natural language description.

The LLM understands natural language dates like:
  - "today", "tomorrow", "next Monday"
  - "in 2 days", "next week"
  - "2025-01-15" (explicit YYYY-MM-DD)

Examples:
  sancho plan "Write thesis introduction, review PRs, email clients"
  sancho plan "3 hours of coding tomorrow, 1 hour meeting next Monday"
  sancho plan "Focus on documentation today" --dry-run

Interactive mode:
  After the AI proposes a schedule, you can:
  - [a]ccept: Save the tasks to your schedule
  - [m]odify: Provide feedback to adjust the proposal
  - [c]ancel: Exit without saving`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.ensureRepo(); err != nil {
				return err
			}

			input := strings.Join(args, " ")

			// Use config default for model if not overridden
			model := modelFlag
			if model == "" {
				model = a.config.LLM.Model
			}
			provider := a.config.LLM.Provider
			baseURL := a.config.LLM.BaseURL

			// Create LLM client
			client, err := llm.NewClient(provider, model, baseURL)
			if err != nil {
				return fmt.Errorf("creating LLM client: %w", err)
			}

			// Create planner
			p := dwplanner.New(client, a.config, a.repo)

			// Initial planning
			fmt.Println("Planning tasks...")
			result, err := p.PlanWithRetry(context.Background(), dwplanner.PlanRequest{
				Input: input,
			}, maxRetries)
			if err != nil {
				return fmt.Errorf("planning: %w", err)
			}

			// Interactive loop
			reader := bufio.NewReader(os.Stdin)
			for {
				// Display results
				a.displayPlanResult(result)

				// Show validation errors if any
				if result.HasValidationErrors() {
					fmt.Println("\nValidation errors (LLM retry limit reached):")
					for _, ve := range result.ValidationErrors {
						fmt.Printf("  - %s\n", ve.Message)
					}
				}

				// If dry run, show and exit
				if dryRun {
					fmt.Println("\n(Dry run - tasks not saved)")
					return nil
				}

				// Prompt for action
				fmt.Print("\n[a]ccept / [m]odify / [c]ancel: ")
				choice, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("reading input: %w", err)
				}
				choice = strings.TrimSpace(strings.ToLower(choice))

				switch choice {
				case "a", "accept":
					if result.HasValidationErrors() {
						fmt.Println("Cannot save: there are unresolved validation errors.")
						fmt.Println("Please [m]odify the plan or [c]ancel.")
						continue
					}

					// Save tasks
					if err := p.Save(context.Background(), result); err != nil {
						return fmt.Errorf("saving tasks: %w", err)
					}
					fmt.Printf("\n%d tasks saved to database\n", result.TotalTasks())
					return nil

				case "m", "modify":
					fmt.Print("What would you like to change? ")
					modification, err := reader.ReadString('\n')
					if err != nil {
						return fmt.Errorf("reading input: %w", err)
					}
					modification = strings.TrimSpace(modification)
					if modification == "" {
						fmt.Println("No modification provided, showing current plan...")
						continue
					}

					fmt.Println("\nReplanning...")
					result, err = p.ContinuePlanning(context.Background(), modification, maxRetries)
					if err != nil {
						return fmt.Errorf("replanning: %w", err)
					}
					// Loop back to display new result

				case "c", "cancel":
					fmt.Println("Planning cancelled.")
					return nil

				default:
					fmt.Println("Invalid choice. Please enter 'a', 'm', or 'c'.")
				}
			}
		},
	}

	cmd.Flags().StringVar(&modelFlag, "model", "", "LLM model to use (from config if not set)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show planned tasks without saving")

	return cmd
}

// displayPlanResult shows the planning result to the user.
func (a *App) displayPlanResult(result *dwplanner.PlanResult) {
	fmt.Println()

	// Show any workday notes
	if result.IsNonWorkday {
		fmt.Printf("Note: %s is not a configured workday\n", result.TodayDate.Format("Monday"))
	}

	fmt.Printf("Planning context: %s\n", result.TodayDate.Format("Monday, January 2, 2006"))
	fmt.Printf("Available today: %s - %s (%dh %dm)\n",
		result.EffectiveStart, result.EffectiveEnd,
		result.AvailableMinutes/60, result.AvailableMinutes%60)

	// Show warnings and suggestions
	if len(result.Warnings) > 0 {
		fmt.Println("\nWarnings:")
		for _, w := range result.Warnings {
			fmt.Printf("  ! %s\n", w)
		}
	}

	if len(result.Suggestions) > 0 {
		fmt.Println("\nSuggestions:")
		for _, s := range result.Suggestions {
			fmt.Printf("  * %s\n", s)
		}
	}

	// Show tasks grouped by date
	if result.TotalTasks() == 0 {
		fmt.Println("\nNo tasks proposed.")
		return
	}

	for _, dateStr := range result.SortedDates {
		tasks := result.TasksByDate[dateStr]
		if len(tasks) == 0 {
			continue
		}

		// Parse date for display
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			fmt.Printf("\n%s:\n", dateStr)
		} else {
			fmt.Printf("\n%s:\n", date.Format("Monday, January 2"))
		}
		fmt.Println(strings.Repeat("-", 60))
		displayTasks(tasks)
	}

	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("Total: %d tasks", result.TotalTasks())
	if len(result.SortedDates) > 1 {
		fmt.Printf(" across %d days", len(result.SortedDates))
	}
	fmt.Println()
}

func displayTasks(tasks []dwplanner.PlannedTask) {
	for _, t := range tasks {
		categoryIcon := "[D]" // deep
		if t.Category == "shallow" {
			categoryIcon = "[S]"
		}
		fmt.Printf("  %s %s-%s  %s\n",
			categoryIcon,
			t.ScheduledStart,
			t.ScheduledEnd,
			t.Description,
		)
	}
}

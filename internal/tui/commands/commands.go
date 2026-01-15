// Package commands provides TUI command constructors and message types.
package commands

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/javiermolinar/sancho/internal/config"
	"github.com/javiermolinar/sancho/internal/dwplanner"
	"github.com/javiermolinar/sancho/internal/llm"
	"github.com/javiermolinar/sancho/internal/summary"
	"github.com/javiermolinar/sancho/internal/task"
)

// WeekLoadedMsg is sent when week data is loaded.
type WeekLoadedMsg struct {
	Week *task.Week
}

// InitialLoadMsg is sent when all 3 weeks are loaded initially.
type InitialLoadMsg struct {
	Window *task.WeekWindow
}

// WeekShiftedMsg is sent when a new edge week is loaded after navigation.
type WeekShiftedMsg struct {
	Week    *task.Week
	Forward bool // true if shifted forward, false if backward
}

// ErrMsg is sent when an error occurs.
type ErrMsg struct {
	Err error
}

// StatusMsgCmd is sent for temporary status messages.
type StatusMsgCmd struct {
	Msg string
}

// ClearStatusMsg is sent to clear the status message.
type ClearStatusMsg struct{}

// PlanStartedMsg is sent when planning starts.
type PlanStartedMsg struct{}

// PlanResultMsg is sent when planning completes.
type PlanResultMsg struct {
	Result  *dwplanner.PlanResult
	Planner *dwplanner.Planner
}

// PlanSavedMsg is sent when plan is saved successfully.
type PlanSavedMsg struct {
	Count int
}

// WeekSummaryMsg is sent when week summary data is ready.
type WeekSummaryMsg struct {
	Summary *summary.WeekSummary
}

// LoadInitialWeeks loads 3 weeks (prev, current, next).
func LoadInitialWeeks(repo task.Repository, weekStart time.Time) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		prevStart := weekStart.AddDate(0, 0, -7)
		currStart := weekStart
		nextStart := weekStart.AddDate(0, 0, 7)

		prevTasks, err := repo.ListTasksByDateRange(ctx, prevStart, prevStart.AddDate(0, 0, 6))
		if err != nil {
			return ErrMsg{Err: err}
		}

		currTasks, err := repo.ListTasksByDateRange(ctx, currStart, currStart.AddDate(0, 0, 6))
		if err != nil {
			return ErrMsg{Err: err}
		}

		nextTasks, err := repo.ListTasksByDateRange(ctx, nextStart, nextStart.AddDate(0, 0, 6))
		if err != nil {
			return ErrMsg{Err: err}
		}

		return InitialLoadMsg{
			Window: task.NewWeekWindow(
				task.NewWeekFromTasks(prevStart, prevTasks),
				task.NewWeekFromTasks(currStart, currTasks),
				task.NewWeekFromTasks(nextStart, nextTasks),
			),
		}
	}
}

// LoadWeek loads tasks for the current week only (used after mutations).
func LoadWeek(repo task.Repository, weekStart time.Time) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		start := weekStart
		end := start.AddDate(0, 0, 6)

		tasks, err := repo.ListTasksByDateRange(ctx, start, end)
		if err != nil {
			return ErrMsg{Err: err}
		}

		return WeekLoadedMsg{Week: task.NewWeekFromTasks(start, tasks)}
	}
}

// LoadNextWeek loads the next week after shifting forward.
func LoadNextWeek(repo task.Repository, weekStart time.Time) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		nextStart := weekStart.AddDate(0, 0, 7)

		tasks, err := repo.ListTasksByDateRange(ctx, nextStart, nextStart.AddDate(0, 0, 6))
		if err != nil {
			return ErrMsg{Err: err}
		}

		return WeekShiftedMsg{Week: task.NewWeekFromTasks(nextStart, tasks), Forward: true}
	}
}

// LoadPrevWeek loads the previous week after shifting backward.
func LoadPrevWeek(repo task.Repository, weekStart time.Time) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		prevStart := weekStart.AddDate(0, 0, -7)

		tasks, err := repo.ListTasksByDateRange(ctx, prevStart, prevStart.AddDate(0, 0, 6))
		if err != nil {
			return ErrMsg{Err: err}
		}

		return WeekShiftedMsg{Week: task.NewWeekFromTasks(prevStart, tasks), Forward: false}
	}
}

// SavePlan creates a command to save the current plan.
func SavePlan(planner *dwplanner.Planner, result *dwplanner.PlanResult) tea.Cmd {
	return func() tea.Msg {
		if planner == nil || result == nil {
			return ErrMsg{Err: fmt.Errorf("no plan to save")}
		}

		if err := planner.Save(context.Background(), result); err != nil {
			return ErrMsg{Err: fmt.Errorf("saving plan: %w", err)}
		}

		return PlanSavedMsg{Count: result.TotalTasks()}
	}
}

// Plan creates a command that runs the LLM planning.
func Plan(input string, cfg *config.Config, repo task.Repository) tea.Cmd {
	return func() tea.Msg {
		client, err := llm.NewClient(cfg.LLM.Provider, cfg.LLM.Model, cfg.LLM.BaseURL)
		if err != nil {
			return ErrMsg{Err: fmt.Errorf("creating LLM client: %w", err)}
		}

		planner := dwplanner.New(client, cfg, repo)

		result, err := planner.PlanWithRetry(context.Background(), dwplanner.PlanRequest{Input: input}, 3)
		if err != nil {
			return ErrMsg{Err: fmt.Errorf("planning: %w", err)}
		}

		return PlanResultMsg{Result: result, Planner: planner}
	}
}

// WeekSummary builds a week summary for the current week.
func WeekSummary(cfg *config.Config, repo task.Repository, weekStart time.Time) tea.Cmd {
	return func() tea.Msg {
		weekSummary, err := summary.BuildWeekSummary(context.Background(), repo, summary.BuildWeekSummaryOptions{
			WeekStart:      weekStart,
			PeakStart:      cfg.Schedule.PeakHoursStart,
			PeakEnd:        cfg.Schedule.PeakHoursEnd,
			IncludeInsight: true,
			Provider:       cfg.LLM.Provider,
			Model:          cfg.LLM.Model,
			BaseURL:        cfg.LLM.BaseURL,
		})
		if err != nil {
			return ErrMsg{Err: err}
		}
		return WeekSummaryMsg{Summary: weekSummary}
	}
}

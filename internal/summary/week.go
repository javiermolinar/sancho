// Package summary provides shared week summary utilities.
package summary

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/javiermolinar/sancho/internal/dateutil"
	"github.com/javiermolinar/sancho/internal/llm"
	"github.com/javiermolinar/sancho/internal/task"
)

// WeekSummary holds aggregated week data and optional insight.
type WeekSummary struct {
	Start   time.Time
	End     time.Time
	Tasks   []*task.Task
	Stats   task.WeekStats
	Insight string
}

// WeekSummaryOptions configures week summary statistics.
type WeekSummaryOptions struct {
	PeakStart string
	PeakEnd   string
}

// BuildWeekSummaryOptions configures the repository-backed summary builder.
type BuildWeekSummaryOptions struct {
	WeekStart      time.Time
	PeakStart      string
	PeakEnd        string
	IncludeInsight bool
	Provider       string
	Model          string
	BaseURL        string
}

// SummarizeWeek builds week summary data from tasks and a reference date.
func SummarizeWeek(weekStart time.Time, tasks []*task.Task, opts WeekSummaryOptions) *WeekSummary {
	start, end := dateutil.WeekRange(weekStart)
	week := task.NewWeekFromTasks(start, tasks)
	stats := week.Stats()
	if opts.PeakStart != "" && opts.PeakEnd != "" {
		stats = week.StatsWithPeakHours(opts.PeakStart, opts.PeakEnd)
	}

	return &WeekSummary{
		Start: start,
		End:   end,
		Tasks: week.AllTasks(),
		Stats: stats,
	}
}

// BuildWeekSummary loads tasks for the requested week and optionally adds insight.
func BuildWeekSummary(ctx context.Context, repo task.Repository, opts BuildWeekSummaryOptions) (*WeekSummary, error) {
	weekStart := opts.WeekStart
	if weekStart.IsZero() {
		weekStart = time.Now()
	}

	start, end := dateutil.WeekRange(weekStart)
	tasks, err := repo.ListTasksByDateRange(ctx, start, end)
	if err != nil {
		return nil, fmt.Errorf("fetching tasks: %w", err)
	}

	summary := SummarizeWeek(start, tasks, WeekSummaryOptions{
		PeakStart: opts.PeakStart,
		PeakEnd:   opts.PeakEnd,
	})

	if opts.IncludeInsight && len(summary.Tasks) > 0 {
		if opts.Model == "" {
			return nil, errors.New("model is required for insight")
		}
		client, err := llm.NewClient(opts.Provider, opts.Model, opts.BaseURL)
		if err != nil {
			return nil, fmt.Errorf("creating LLM client: %w", err)
		}

		evaluator := llm.NewEvaluatorWithOpts(client, llm.EvalOpts{
			PeakStart: opts.PeakStart,
			PeakEnd:   opts.PeakEnd,
		})
		result, err := evaluator.EvaluateWeek(ctx, start, end, summary.Tasks)
		if err != nil {
			return nil, fmt.Errorf("evaluating week: %w", err)
		}
		summary.Insight = result
	}

	return summary, nil
}

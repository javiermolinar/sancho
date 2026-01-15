// Package dwplanner provides high-level deep work planning orchestration.
// It coordinates the LLM, scheduler, and repository to plan tasks from natural language input.
// Both CLI and TUI can use this package.
package dwplanner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/javiermolinar/sancho/internal/config"
	"github.com/javiermolinar/sancho/internal/llm"
	"github.com/javiermolinar/sancho/internal/scheduler"
	"github.com/javiermolinar/sancho/internal/task"
)

// ErrMaxRetriesExceeded is returned when all retry attempts fail validation.
var ErrMaxRetriesExceeded = errors.New("maximum retries exceeded, validation still failing")

// Planner orchestrates task planning using LLM, scheduler, and repository.
type Planner struct {
	llmClient llm.Client
	scheduler *scheduler.Scheduler
	repo      task.Repository
	config    *config.Config

	// Conversation state for interactive planning
	messages      []llm.Message
	existingTasks []*task.Task
	lastResponse  *llm.PlanResponse
}

func useCompactPrompt(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case llm.ProviderOllama, llm.ProviderLMStudio, "lm-studio", "llmstudio":
		return true
	default:
		return false
	}
}

// New creates a new Planner with the given dependencies.
func New(client llm.Client, cfg *config.Config, repo task.Repository) *Planner {
	sched := scheduler.New(cfg.Schedule.Workdays, cfg.Schedule.DayStart, cfg.Schedule.DayEnd)
	return &Planner{
		llmClient: client,
		scheduler: sched,
		repo:      repo,
		config:    cfg,
	}
}

// PlanRequest contains the input for planning.
type PlanRequest struct {
	Input string // Natural language description of tasks
}

// PlanResult contains the result of a planning operation.
type PlanResult struct {
	// TasksByDate groups planned tasks by their scheduled date (YYYY-MM-DD)
	TasksByDate map[string][]PlannedTask

	// SortedDates contains the dates in chronological order for display
	SortedDates []string

	// LLM metadata
	Warnings    []string
	Suggestions []string

	// Validation info (populated if retries exhausted)
	ValidationErrors []ValidationError

	// Context for display
	EffectiveStart   string
	EffectiveEnd     string
	AvailableMinutes int
	IsNonWorkday     bool
	TodayDate        time.Time
}

// PlannedTask represents a single planned task.
type PlannedTask struct {
	Description    string
	Category       string // "deep" or "shallow"
	ScheduledDate  string // YYYY-MM-DD
	ScheduledStart string // "HH:MM"
	ScheduledEnd   string // "HH:MM"
}

// TotalTasks returns the total number of planned tasks across all days.
func (r *PlanResult) TotalTasks() int {
	total := 0
	for _, tasks := range r.TasksByDate {
		total += len(tasks)
	}
	return total
}

// HasValidationErrors returns true if there are unresolved validation errors.
func (r *PlanResult) HasValidationErrors() bool {
	return len(r.ValidationErrors) > 0
}

// PlanWithRetry creates a schedule from natural language input with validation and retry.
// It fetches existing tasks, calls the LLM, validates the response, and retries on failure.
// If maxRetries are exhausted, returns result with ValidationErrors populated.
func (p *Planner) PlanWithRetry(ctx context.Context, req PlanRequest, maxRetries int) (*PlanResult, error) {
	now := time.Now()

	// Fetch existing tasks for context
	existing, err := p.fetchExistingTasks(ctx, now)
	if err != nil {
		return nil, fmt.Errorf("fetching existing tasks: %w", err)
	}
	p.existingTasks = existing

	recent, err := p.fetchRecentTasks(ctx, now)
	if err != nil {
		return nil, fmt.Errorf("fetching recent tasks: %w", err)
	}

	// Calculate scheduling context
	slot := p.scheduler.NextAvailableStart(now)
	effectiveStart := slot.Start
	effectiveEnd := p.config.Schedule.DayEnd
	availableMinutes := p.scheduler.AvailableMinutes(scheduler.AvailableSlot{
		Start: effectiveStart,
		End:   effectiveEnd,
	})

	// Find next workday
	nextWorkdaySlot := p.scheduler.NextAvailableStart(
		time.Date(slot.Date.Year(), slot.Date.Month(), slot.Date.Day(), 23, 59, 0, 0, time.Local),
	)

	// Build initial LLM request
	llmReq := llm.PlanRequest{
		Input:            req.Input,
		Date:             now,
		DayStart:         effectiveStart,
		DayEnd:           effectiveEnd,
		NextWorkday:      nextWorkdaySlot.Date.Format("Monday, January 2"),
		ExistingTasks:    p.convertToExistingTasks(existing),
		RecentTasks:      p.convertToExistingTasks(recent),
		UseCompactPrompt: useCompactPrompt(p.config.LLM.Provider),
	}

	// Build initial messages
	llmPlanner := llm.NewPlanner(p.llmClient)
	p.messages = llmPlanner.BuildInitialMessages(llmReq)

	// Add user message
	p.messages = append(p.messages, llm.Message{Role: "user", Content: req.Input})

	// Validation loop
	var lastValidation ValidationResult
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Call LLM
		resp, err := llmPlanner.PlanWithMessages(ctx, p.messages)
		if err != nil {
			return nil, fmt.Errorf("LLM planning (attempt %d): %w", attempt+1, err)
		}
		p.lastResponse = resp

		// Validate response
		validator := NewValidator(now, effectiveStart, effectiveEnd, existing)
		lastValidation = validator.Validate(resp.Tasks)

		if lastValidation.Valid {
			// Success - build result
			return p.buildResult(resp, slot.Date, effectiveStart, effectiveEnd, availableMinutes, nil), nil
		}

		// Validation failed - append error feedback for retry
		if attempt < maxRetries {
			// Append assistant response (for context)
			respJSON, _ := json.Marshal(resp)
			p.messages = append(p.messages, llm.Message{
				Role:    "assistant",
				Content: string(respJSON),
			})

			// Append error feedback
			p.messages = append(p.messages, llm.Message{
				Role:    "user",
				Content: lastValidation.FormatErrors(),
			})
		}
	}

	// All retries exhausted - return result with errors
	result := p.buildResult(p.lastResponse, slot.Date, effectiveStart, effectiveEnd, availableMinutes, lastValidation.Errors)
	return result, nil
}

// ContinuePlanning adds context to the conversation and replans.
// Used when user wants to modify the proposal.
func (p *Planner) ContinuePlanning(ctx context.Context, additionalContext string, maxRetries int) (*PlanResult, error) {
	if len(p.messages) == 0 {
		return nil, errors.New("no active planning session")
	}

	now := time.Now()

	// Calculate scheduling context
	slot := p.scheduler.NextAvailableStart(now)
	effectiveStart := slot.Start
	effectiveEnd := p.config.Schedule.DayEnd
	availableMinutes := p.scheduler.AvailableMinutes(scheduler.AvailableSlot{
		Start: effectiveStart,
		End:   effectiveEnd,
	})

	// Add previous response to context if we have one
	if p.lastResponse != nil {
		respJSON, _ := json.Marshal(p.lastResponse)
		p.messages = append(p.messages, llm.Message{
			Role:    "assistant",
			Content: string(respJSON),
		})
	}

	// Add user's additional context
	p.messages = append(p.messages, llm.Message{
		Role:    "user",
		Content: additionalContext,
	})

	// Re-plan with updated messages
	llmPlanner := llm.NewPlanner(p.llmClient)

	var lastValidation ValidationResult
	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := llmPlanner.PlanWithMessages(ctx, p.messages)
		if err != nil {
			return nil, fmt.Errorf("LLM planning (attempt %d): %w", attempt+1, err)
		}
		p.lastResponse = resp

		// Validate response
		validator := NewValidator(now, effectiveStart, effectiveEnd, p.existingTasks)
		lastValidation = validator.Validate(resp.Tasks)

		if lastValidation.Valid {
			return p.buildResult(resp, slot.Date, effectiveStart, effectiveEnd, availableMinutes, nil), nil
		}

		// Retry with error feedback
		if attempt < maxRetries {
			respJSON, _ := json.Marshal(resp)
			p.messages = append(p.messages, llm.Message{
				Role:    "assistant",
				Content: string(respJSON),
			})
			p.messages = append(p.messages, llm.Message{
				Role:    "user",
				Content: lastValidation.FormatErrors(),
			})
		}
	}

	// Return with validation errors
	result := p.buildResult(p.lastResponse, slot.Date, effectiveStart, effectiveEnd, availableMinutes, lastValidation.Errors)
	return result, nil
}

// Save persists the planned tasks to the repository.
func (p *Planner) Save(ctx context.Context, result *PlanResult) error {
	if result.HasValidationErrors() {
		return errors.New("cannot save: result has validation errors")
	}

	var tasks []*task.Task
	for _, dateTasks := range result.TasksByDate {
		for _, pt := range dateTasks {
			t, err := p.toTask(pt)
			if err != nil {
				return fmt.Errorf("converting task: %w", err)
			}
			tasks = append(tasks, t)
		}
	}

	if len(tasks) == 0 {
		return nil
	}

	return p.repo.CreateTasks(ctx, tasks)
}

// fetchExistingTasks retrieves all scheduled tasks from the given date onwards.
func (p *Planner) fetchExistingTasks(ctx context.Context, from time.Time) ([]*task.Task, error) {
	startOfDay := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	endDate := startOfDay.AddDate(0, 1, 0) // One month ahead
	return p.repo.ListTasksByDateRange(ctx, startOfDay, endDate)
}

// fetchRecentTasks retrieves scheduled tasks from the previous 14 days for history context.
func (p *Planner) fetchRecentTasks(ctx context.Context, from time.Time) ([]*task.Task, error) {
	startOfDay := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	historyStart := startOfDay.AddDate(0, 0, -14)
	historyEnd := startOfDay.AddDate(0, 0, -1)
	return p.repo.ListTasksByDateRange(ctx, historyStart, historyEnd)
}

// convertToExistingTasks converts task.Task slice to llm.ExistingTask slice.
func (p *Planner) convertToExistingTasks(tasks []*task.Task) []llm.ExistingTask {
	result := make([]llm.ExistingTask, 0, len(tasks))
	for _, t := range tasks {
		if !t.IsScheduled() {
			continue // Skip cancelled/postponed tasks
		}
		result = append(result, llm.ExistingTask{
			Date:        t.ScheduledDate.Format("2006-01-02"),
			Start:       t.ScheduledStart,
			End:         t.ScheduledEnd,
			Description: t.Description,
			Category:    string(t.Category),
		})
	}
	return result
}

// buildResult creates a PlanResult from an LLM response.
func (p *Planner) buildResult(resp *llm.PlanResponse, todayDate time.Time, effectiveStart, effectiveEnd string, availableMinutes int, validationErrors []ValidationError) *PlanResult {
	result := &PlanResult{
		TasksByDate:      make(map[string][]PlannedTask),
		Warnings:         resp.Warnings,
		Suggestions:      resp.Suggestions,
		ValidationErrors: validationErrors,
		EffectiveStart:   effectiveStart,
		EffectiveEnd:     effectiveEnd,
		AvailableMinutes: availableMinutes,
		IsNonWorkday:     !p.scheduler.IsWorkday(todayDate),
		TodayDate:        todayDate,
	}

	// Group tasks by date
	for _, t := range resp.Tasks {
		pt := PlannedTask{
			Description:    t.Description,
			Category:       t.Category,
			ScheduledDate:  t.ScheduledDate,
			ScheduledStart: t.ScheduledStart,
			ScheduledEnd:   t.ScheduledEnd,
		}
		result.TasksByDate[t.ScheduledDate] = append(result.TasksByDate[t.ScheduledDate], pt)
	}

	// Create sorted date list
	for date := range result.TasksByDate {
		result.SortedDates = append(result.SortedDates, date)
	}
	sort.Strings(result.SortedDates)

	return result
}

// toTask converts a PlannedTask to a domain Task.
func (p *Planner) toTask(pt PlannedTask) (*task.Task, error) {
	category := task.CategoryDeep
	if pt.Category == "shallow" {
		category = task.CategoryShallow
	}

	scheduledDate, err := time.Parse("2006-01-02", pt.ScheduledDate)
	if err != nil {
		return nil, fmt.Errorf("parsing date %q: %w", pt.ScheduledDate, err)
	}

	return &task.Task{
		Description:    pt.Description,
		Category:       category,
		ScheduledDate:  scheduledDate,
		ScheduledStart: pt.ScheduledStart,
		ScheduledEnd:   pt.ScheduledEnd,
		Status:         task.StatusScheduled,
	}, nil
}

// Plan is the legacy method for backwards compatibility.
//
// Deprecated: Use PlanWithRetry for new code.
func (p *Planner) Plan(ctx context.Context, req OldPlanRequest) (*OldPlanResult, error) {
	now := time.Now()

	// Default to today if no date specified
	targetDate := req.Date
	if targetDate.IsZero() {
		targetDate = now
	}

	result := &OldPlanResult{
		TodayDate:    targetDate,
		EffectiveEnd: p.config.Schedule.DayEnd,
	}

	// Check if target date is a workday
	result.IsNonWorkday = !p.scheduler.IsWorkday(targetDate)

	// Determine if we're planning for today
	isToday := targetDate.Year() == now.Year() && targetDate.YearDay() == now.YearDay()

	// Calculate effective start time
	var planTime time.Time
	if isToday {
		slot := p.scheduler.NextAvailableStart(now)
		if slot.Date.YearDay() != now.YearDay() || slot.Date.Year() != now.Year() {
			result.DateChanged = true
			result.TodayDate = slot.Date
			result.IsNonWorkday = !p.scheduler.IsWorkday(slot.Date)
			result.EffectiveStart = slot.Start
			planTime = time.Date(result.TodayDate.Year(), result.TodayDate.Month(), result.TodayDate.Day(),
				9, 0, 0, 0, now.Location())
		} else {
			result.EffectiveStart = slot.Start
			planTime = time.Date(result.TodayDate.Year(), result.TodayDate.Month(), result.TodayDate.Day(),
				now.Hour(), now.Minute(), 0, 0, now.Location())
		}
	} else {
		result.EffectiveStart = p.config.Schedule.DayStart
		planTime = time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(),
			9, 0, 0, 0, time.Local)
	}

	result.AvailableMinutes = p.scheduler.AvailableMinutes(scheduler.AvailableSlot{
		Start: result.EffectiveStart,
		End:   result.EffectiveEnd,
	})

	nextWorkdaySlot := p.scheduler.NextAvailableStart(
		time.Date(result.TodayDate.Year(), result.TodayDate.Month(), result.TodayDate.Day(),
			23, 59, 0, 0, time.Local),
	)
	result.NextWorkdayDate = nextWorkdaySlot.Date

	llmPlanner := llm.NewPlanner(p.llmClient)
	resp, err := llmPlanner.Plan(ctx, llm.PlanRequest{
		Input:            req.Input,
		Date:             planTime,
		DayStart:         result.EffectiveStart,
		DayEnd:           result.EffectiveEnd,
		NextWorkday:      result.NextWorkdayDate.Format("Monday, January 2"),
		UseCompactPrompt: useCompactPrompt(p.config.LLM.Provider),
	})
	if err != nil {
		return nil, err
	}

	result.Warnings = resp.Warnings
	result.Suggestions = resp.Suggestions

	for _, t := range resp.Tasks {
		pt := OldPlannedTask{
			Description:    t.Description,
			Category:       t.Category,
			ScheduledStart: t.ScheduledStart,
			ScheduledEnd:   t.ScheduledEnd,
		}

		if t.ScheduledDate == "next_workday" {
			result.NextWorkdayTasks = append(result.NextWorkdayTasks, pt)
		} else {
			result.TodayTasks = append(result.TodayTasks, pt)
		}
	}

	return result, nil
}

// OldPlanRequest is the legacy request type.
//
// Deprecated: Use PlanRequest for new code.
type OldPlanRequest struct {
	Input string
	Date  time.Time
}

// OldPlanResult is the legacy result type.
//
// Deprecated: Use PlanResult for new code.
type OldPlanResult struct {
	TodayDate        time.Time
	NextWorkdayDate  time.Time
	EffectiveStart   string
	EffectiveEnd     string
	AvailableMinutes int
	TodayTasks       []OldPlannedTask
	NextWorkdayTasks []OldPlannedTask
	Warnings         []string
	Suggestions      []string
	DateChanged      bool
	IsNonWorkday     bool
}

// TotalTasks returns the total number of tasks.
func (r *OldPlanResult) TotalTasks() int {
	return len(r.TodayTasks) + len(r.NextWorkdayTasks)
}

// OldPlannedTask is the legacy task type.
//
// Deprecated: Use PlannedTask for new code.
type OldPlannedTask struct {
	Description    string
	Category       string
	ScheduledStart string
	ScheduledEnd   string
}

// LegacySave persists the old plan result format.
//
// Deprecated: Use Save with PlanResult for new code.
func (p *Planner) LegacySave(ctx context.Context, result *OldPlanResult) error {
	var tasks []*task.Task

	for _, pt := range result.TodayTasks {
		tasks = append(tasks, p.legacyToTask(pt, result.TodayDate))
	}

	for _, pt := range result.NextWorkdayTasks {
		tasks = append(tasks, p.legacyToTask(pt, result.NextWorkdayDate))
	}

	if len(tasks) == 0 {
		return nil
	}

	return p.repo.CreateTasks(ctx, tasks)
}

func (p *Planner) legacyToTask(pt OldPlannedTask, date time.Time) *task.Task {
	category := task.CategoryDeep
	if pt.Category == "shallow" {
		category = task.CategoryShallow
	}

	return &task.Task{
		Description:    pt.Description,
		Category:       category,
		ScheduledDate:  date,
		ScheduledStart: pt.ScheduledStart,
		ScheduledEnd:   pt.ScheduledEnd,
		Status:         task.StatusScheduled,
	}
}

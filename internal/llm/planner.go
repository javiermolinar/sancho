package llm

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/javiermolinar/sancho/internal/task"
)

const systemPromptWithContext = `You are a productivity assistant implementing Cal Newport's deep work methodology.

Context:
- Current date and time: %s, %s %s (format: DayOfWeek, YYYY-MM-DD HH:MM)
- Today: %s (%s, %s)
- Tomorrow: %s (%s, %s)
- Configured workday hours: %s to %s (Mon-Fri typically)
- Next workday: %s (%s)

%s

%s

%s

User request: "%s"

CRITICAL DATE RULES:
1. "today" ALWAYS means %s, even if it's a weekend
2. "tomorrow" ALWAYS means the next calendar day (%s), even if it's a weekend
3. If the user explicitly asks for a weekend date, SCHEDULE IT on that weekend
4. Only suggest workdays as alternatives, never silently change "today" or "tomorrow"
5. For weekends, use the same workday hours (e.g., %s to %s) unless user specifies otherwise

Date parsing:
- "today" → %s (current day, even if weekend)
- "tomorrow" → %s (next calendar day, even if weekend)  
- "monday", "next monday", "next M" → next occurrence of Monday
- "saturday", "next saturday" → next Saturday (schedule it!)
- "in X days" → add X days to today
- "next week" → add 7 days to today
- Explicit "YYYY-MM-DD" → use that exact date

Other rules:
1. Resolve ALL dates to YYYY-MM-DD format in scheduled_date
2. Never schedule before current time (%s) if scheduling for today
3. Never overlap with existing tasks listed above
4. Use 24-hour time format (HH:MM) for scheduled_start and scheduled_end
5. Round durations to 15-minute increments (minimum 15 minutes)
6. Categorize as "deep" (focused, cognitively demanding) or "shallow" (admin, meetings, email)
7. Schedule deep work in longer blocks, prefer earlier in the day
8. Batch shallow tasks together when possible
9. Add a warning if scheduling on a weekend (but still schedule it!)
10. Warn if tasks don't fit in available time
11. If a task lacks a specific time, infer a likely placement using the recent schedule history above
12. If a task matches a suggested time window, prefer that time unless the user specifies otherwise

Respond ONLY with valid JSON (no markdown, no explanation):
{
  "tasks": [
    {
      "description": "string",
      "category": "deep" or "shallow",
      "scheduled_date": "YYYY-MM-DD",
      "scheduled_start": "HH:MM",
      "scheduled_end": "HH:MM"
    }
  ],
  "warnings": ["string"],
  "suggestions": ["string"]
}`

const systemPromptCompact = `You are a scheduling assistant. Use the context and return JSON only.

Today: %s (%s)
Tomorrow: %s (%s)
Current time: %s
Workday hours: %s to %s
Next workday: %s (%s)

%s

User request: "%s"

Rules:
- Return JSON only (no markdown).
- Use scheduled_date YYYY-MM-DD and time HH:MM (24-hour).
- Schedule tasks within workday hours unless user explicitly requests otherwise.
- Do not overlap with existing tasks above.
- Do not schedule before current time if scheduling today.
- Use 15-minute increments (minimum 15 minutes).
- Category must be "deep" or "shallow".
- "warnings" and "suggestions" must be arrays of strings (no objects).

JSON schema:
{
  "tasks": [
    {
      "description": "string",
      "category": "deep" or "shallow",
      "scheduled_date": "YYYY-MM-DD",
      "scheduled_start": "HH:MM",
      "scheduled_end": "HH:MM"
    }
  ],
  "warnings": ["string"],
  "suggestions": ["string"]
}`

// ExistingTask represents a task already in the schedule for LLM context.
type ExistingTask struct {
	Date        string // YYYY-MM-DD
	Start       string // HH:MM
	End         string // HH:MM
	Description string
	Category    string // "deep" or "shallow"
}

// PlanRequest contains the input for the planner.
type PlanRequest struct {
	Input            string
	Date             time.Time
	DayStart         string         // "HH:MM"
	DayEnd           string         // "HH:MM"
	NextWorkday      string         // e.g., "Monday, January 13"
	ExistingTasks    []ExistingTask // Tasks already scheduled (for overlap avoidance)
	RecentTasks      []ExistingTask // Recent history for schedule pattern inference
	UseCompactPrompt bool           // Use a shorter prompt for local models
}

// PlanResponse contains the parsed LLM response.
type PlanResponse struct {
	Tasks       []PlannedTask `json:"tasks"`
	Warnings    []string      `json:"warnings"`
	Suggestions []string      `json:"suggestions"`
}

// PlannedTask represents a task planned by the LLM.
type PlannedTask struct {
	Description    string `json:"description"`
	Category       string `json:"category"`
	ScheduledDate  string `json:"scheduled_date"` // YYYY-MM-DD format
	ScheduledStart string `json:"scheduled_start"`
	ScheduledEnd   string `json:"scheduled_end"`
}

// Planner uses an LLM to plan tasks from natural language input.
type Planner struct {
	client Client
}

// NewPlanner creates a new Planner with the given LLM client.
func NewPlanner(client Client) *Planner {
	return &Planner{client: client}
}

// Plan converts natural language input into scheduled tasks.
func (p *Planner) Plan(ctx context.Context, req PlanRequest) (*PlanResponse, error) {
	messages := p.buildInitialMessages(req)
	return p.planWithMessages(ctx, messages)
}

// PlanWithMessages allows planning with a pre-built message history.
// This is used for retry logic where we need to append error feedback.
func (p *Planner) PlanWithMessages(ctx context.Context, messages []Message) (*PlanResponse, error) {
	return p.planWithMessages(ctx, messages)
}

// BuildInitialMessages creates the initial message list for a planning request.
// Exported so dwplanner can build and modify messages for retries.
func (p *Planner) BuildInitialMessages(req PlanRequest) []Message {
	return p.buildInitialMessages(req)
}

func (p *Planner) buildInitialMessages(req PlanRequest) []Message {
	dayOfWeek := req.Date.Format("Monday")
	currentDate := req.Date.Format("2006-01-02")
	currentTime := req.Date.Format("15:04")
	tomorrow := req.Date.AddDate(0, 0, 1)
	tomorrowDate := tomorrow.Format("2006-01-02")
	tomorrowDay := tomorrow.Format("Monday")
	todayKind := dayKind(req.Date)
	tomorrowKind := dayKind(tomorrow)

	dayStart := req.DayStart
	if dayStart == "" {
		dayStart = "09:00"
	}
	dayEnd := req.DayEnd
	if dayEnd == "" {
		dayEnd = "18:00" // Updated fallback to match current common usage, but it should ideally come from config
	}
	nextWorkday := req.NextWorkday
	if nextWorkday == "" {
		nextWorkday = "tomorrow"
	}
	nextWorkdayDate := req.Date.AddDate(0, 0, 1).Format("2006-01-02")

	existingSection := p.formatExistingTasks(req.ExistingTasks)
	recentSection := p.formatRecentTasks(req.RecentTasks)
	suggestedSection := p.formatSuggestedTimes(req.RecentTasks)

	var prompt string
	if req.UseCompactPrompt {
		prompt = fmt.Sprintf(systemPromptCompact,
			dayOfWeek,       // Today's day of week
			currentDate,     // Today's date
			tomorrowDay,     // Tomorrow's day of week
			tomorrowDate,    // Tomorrow's date
			currentTime,     // Current time
			dayStart,        // Workday start
			dayEnd,          // Workday end
			nextWorkday,     // Next workday name
			nextWorkdayDate, // Next workday date
			existingSection, // Existing tasks
			req.Input,       // User request
		)
	} else {
		prompt = fmt.Sprintf(systemPromptWithContext,
			dayOfWeek,        // DayOfWeek in header
			currentDate,      // YYYY-MM-DD in header
			currentTime,      // HH:MM in header
			dayOfWeek,        // Today's day of week
			currentDate,      // Today's date
			todayKind,        // Today's day kind
			tomorrowDay,      // Tomorrow's day of week
			tomorrowDate,     // Tomorrow's date
			tomorrowKind,     // Tomorrow's day kind
			dayStart,         // Workday start
			dayEnd,           // Workday end
			nextWorkday,      // Next workday name
			nextWorkdayDate,  // Next workday date
			existingSection,  // Existing tasks
			recentSection,    // Recent history
			suggestedSection, // Suggested time windows
			req.Input,        // User request
			currentDate,      // "today" means this
			tomorrowDate,     // "tomorrow" means this
			dayStart,         // Weekend hours start
			dayEnd,           // Weekend hours end
			currentDate,      // Date parsing: today
			tomorrowDate,     // Date parsing: tomorrow
			currentTime,      // Current time for "don't schedule before"
		)
	}

	return []Message{
		{Role: "system", Content: prompt},
	}
}

func dayKind(t time.Time) string {
	switch t.Weekday() {
	case time.Saturday, time.Sunday:
		return "weekend"
	default:
		return "weekday"
	}
}

func (p *Planner) formatExistingTasks(tasks []ExistingTask) string {
	if len(tasks) == 0 {
		return "Existing scheduled tasks: None"
	}

	var sb strings.Builder
	sb.WriteString("Existing scheduled tasks (avoid overlaps):\n")
	for _, t := range tasks {
		sb.WriteString(fmt.Sprintf("- %s %s-%s: %s [%s]\n",
			t.Date, t.Start, t.End, t.Description, t.Category))
	}
	return sb.String()
}

func (p *Planner) formatRecentTasks(tasks []ExistingTask) string {
	if len(tasks) == 0 {
		return "Recent schedule history (last 14 days): None"
	}

	tasks = sortedExistingTasks(tasks)

	var sb strings.Builder
	sb.WriteString("Recent schedule history (last 14 days):\n")
	for _, t := range tasks {
		sb.WriteString(fmt.Sprintf("- %s %s-%s: %s [%s]\n",
			t.Date, t.Start, t.End, t.Description, t.Category))
	}
	return sb.String()
}

func (p *Planner) formatSuggestedTimes(tasks []ExistingTask) string {
	if len(tasks) == 0 {
		return "Suggested time windows from recent history: None"
	}

	suggestions := suggestedTimeWindows(tasks)
	if len(suggestions) == 0 {
		return "Suggested time windows from recent history: None"
	}

	var sb strings.Builder
	sb.WriteString("Suggested time windows from recent history (median):\n")
	for _, suggestion := range suggestions {
		sb.WriteString(fmt.Sprintf("- %s\n", suggestion))
	}
	return sb.String()
}

func suggestedTimeWindows(tasks []ExistingTask) []string {
	type timeSummary struct {
		starts []int
		ends   []int
		count  int
	}

	summaries := make(map[string]*timeSummary)
	for _, t := range tasks {
		if t.Description == "" {
			continue
		}
		start, ok := minutesFromHHMM(t.Start)
		if !ok {
			continue
		}
		end, ok := minutesFromHHMM(t.End)
		if !ok {
			continue
		}
		summary := summaries[t.Description]
		if summary == nil {
			summary = &timeSummary{}
			summaries[t.Description] = summary
		}
		summary.starts = append(summary.starts, start)
		summary.ends = append(summary.ends, end)
		summary.count++
	}

	if len(summaries) == 0 {
		return nil
	}

	keys := make([]string, 0, len(summaries))
	for key := range summaries {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	suggestions := make([]string, 0, len(keys))
	for _, key := range keys {
		summary := summaries[key]
		if len(summary.starts) == 0 || len(summary.ends) == 0 {
			continue
		}
		sort.Ints(summary.starts)
		sort.Ints(summary.ends)
		startMedian := roundToQuarterHour(medianMinutes(summary.starts))
		endMedian := roundToQuarterHour(medianMinutes(summary.ends))
		suggestions = append(suggestions, fmt.Sprintf("%s: ~%s-%s (n=%d)",
			key, minutesToHHMM(startMedian), minutesToHHMM(endMedian), summary.count))
	}
	return suggestions
}

func sortedExistingTasks(tasks []ExistingTask) []ExistingTask {
	sorted := append([]ExistingTask(nil), tasks...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Date != sorted[j].Date {
			return sorted[i].Date < sorted[j].Date
		}
		if sorted[i].Start != sorted[j].Start {
			return sorted[i].Start < sorted[j].Start
		}
		if sorted[i].End != sorted[j].End {
			return sorted[i].End < sorted[j].End
		}
		return sorted[i].Description < sorted[j].Description
	})
	return sorted
}

func minutesFromHHMM(value string) (int, bool) {
	parsed, err := time.Parse("15:04", value)
	if err != nil {
		return 0, false
	}
	return parsed.Hour()*60 + parsed.Minute(), true
}

func medianMinutes(values []int) int {
	if len(values) == 0 {
		return 0
	}
	mid := len(values) / 2
	if len(values)%2 == 1 {
		return values[mid]
	}
	return (values[mid-1] + values[mid]) / 2
}

func roundToQuarterHour(minutes int) int {
	if minutes < 0 {
		return 0
	}
	rounded := ((minutes + 7) / 15) * 15
	if rounded > 23*60+59 {
		return 23*60 + 59
	}
	return rounded
}

func minutesToHHMM(minutes int) string {
	if minutes < 0 {
		minutes = 0
	}
	if minutes > 23*60+59 {
		minutes = 23*60 + 59
	}
	hours := minutes / 60
	mins := minutes % 60
	return fmt.Sprintf("%02d:%02d", hours, mins)
}

func (p *Planner) planWithMessages(ctx context.Context, messages []Message) (*PlanResponse, error) {
	var resp PlanResponse
	if err := p.client.ChatJSON(ctx, messages, &resp); err != nil {
		return nil, fmt.Errorf("getting plan from LLM: %w", err)
	}
	return &resp, nil
}

// ToTasks converts planned tasks to domain Task objects.
// It parses the YYYY-MM-DD date from each task.
func (pr *PlanResponse) ToTasks() ([]*task.Task, error) {
	tasks := make([]*task.Task, 0, len(pr.Tasks))

	for _, pt := range pr.Tasks {
		category := task.CategoryDeep
		if pt.Category == "shallow" {
			category = task.CategoryShallow
		}

		// Parse the scheduled date
		scheduledDate, err := time.Parse("2006-01-02", pt.ScheduledDate)
		if err != nil {
			return nil, fmt.Errorf("parsing scheduled date %q: %w", pt.ScheduledDate, err)
		}

		t := &task.Task{
			Description:    pt.Description,
			Category:       category,
			ScheduledDate:  scheduledDate,
			ScheduledStart: pt.ScheduledStart,
			ScheduledEnd:   pt.ScheduledEnd,
			Status:         task.StatusScheduled,
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}

// ToTasksWithFallback converts planned tasks to domain Task objects.
// It uses todayDate and nextWorkdayDate as fallbacks for legacy "today"/"next_workday" values.
//
// Deprecated: Use ToTasks() instead once all LLM responses use YYYY-MM-DD format.
func (pr *PlanResponse) ToTasksWithFallback(todayDate, nextWorkdayDate time.Time) ([]*task.Task, error) {
	tasks := make([]*task.Task, 0, len(pr.Tasks))

	for _, pt := range pr.Tasks {
		category := task.CategoryDeep
		if pt.Category == "shallow" {
			category = task.CategoryShallow
		}

		// Determine the scheduled date
		var scheduledDate time.Time
		switch pt.ScheduledDate {
		case "today":
			scheduledDate = todayDate
		case "next_workday":
			scheduledDate = nextWorkdayDate
		default:
			// Try to parse as YYYY-MM-DD
			parsed, err := time.Parse("2006-01-02", pt.ScheduledDate)
			if err != nil {
				// Fallback to today if parsing fails
				scheduledDate = todayDate
			} else {
				scheduledDate = parsed
			}
		}

		t := &task.Task{
			Description:    pt.Description,
			Category:       category,
			ScheduledDate:  scheduledDate,
			ScheduledStart: pt.ScheduledStart,
			ScheduledEnd:   pt.ScheduledEnd,
			Status:         task.StatusScheduled,
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}

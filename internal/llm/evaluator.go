// Package llm provides LLM client and evaluation functionality.
package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/javiermolinar/sancho/internal/task"
)

const evaluatorSystemPrompt = `You are a minimalist productivity analyst. Output ONLY the exact format shown - no markdown, no extra text. Be extremely concise.`

const userPromptTemplate = `Analyze this week's work log and output EXACTLY this format (no markdown, no code blocks):

THEME: [ 2-4 word theme ]

⚠️  PEAK LEAKAGE: Xh of ⚡ energy spent on [S] (specific activities).
📉 ENERGY DECAY: One sentence about how deep work duration changed Mon→Fri.
🧠 RESIDUE RISK: Mention if any [D] block followed [S] with <15m gap.

NEXT WEEK:
➜  First specific action to protect peak hours.
➜  Second specific scheduling change.

Data Format:
- [D] = Deep Work, [S] = Shallow Work  
- ⚡ = Peak Energy Window (%s-%s)

Weekly Data:
%s

Rules:
- Use the exact emoji prefixes shown (⚠️, 📉, 🧠, ➜)
- Keep each line under 70 characters
- Be specific with times and durations from the data
- If no issue exists for a category, omit that line
- Output plain text only, no markdown formatting`

// EvalOpts configures the evaluation behavior.
type EvalOpts struct {
	PeakStart string // Peak hours start (HH:MM), empty if not configured
	PeakEnd   string // Peak hours end (HH:MM), empty if not configured
}

// Evaluator provides LLM-based task evaluation using deep work methodology.
type Evaluator struct {
	client Client
	opts   EvalOpts
}

// NewEvaluator creates a new Evaluator with the given LLM client.
func NewEvaluator(client Client) *Evaluator {
	return &Evaluator{client: client}
}

// NewEvaluatorWithOpts creates a new Evaluator with options.
func NewEvaluatorWithOpts(client Client, opts EvalOpts) *Evaluator {
	return &Evaluator{client: client, opts: opts}
}

// EvaluateWeek sends the week's tasks to the LLM for deep work analysis.
func (e *Evaluator) EvaluateWeek(ctx context.Context, start, end time.Time, tasks []*task.Task) (string, error) {
	weekData := e.formatWeekData(start, end, tasks)

	// Format peak hours for prompt (default to common values if not set)
	peakStart := e.opts.PeakStart
	peakEnd := e.opts.PeakEnd
	if peakStart == "" {
		peakStart = "09:00"
	}
	if peakEnd == "" {
		peakEnd = "12:00"
	}

	prompt := fmt.Sprintf(userPromptTemplate, peakStart, peakEnd, weekData)

	return e.client.Chat(ctx, []Message{
		{Role: "system", Content: evaluatorSystemPrompt},
		{Role: "user", Content: prompt},
	})
}

// formatWeekData formats tasks in the CLI-style format for LLM consumption.
func (e *Evaluator) formatWeekData(start, end time.Time, tasks []*task.Task) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Week: %s - %s\n\n",
		start.Format("Mon Jan 2"),
		end.Format("Mon Jan 2, 2006")))

	var currentDate string
	for _, t := range tasks {
		date := t.ScheduledDate.Format("2006-01-02")
		dayName := t.ScheduledDate.Format("Mon Jan 2")

		// Print day header if new day
		if date != currentDate {
			if currentDate != "" {
				sb.WriteString("\n")
			}
			sb.WriteString(fmt.Sprintf("%s\n", dayName))
			currentDate = date
		}

		// Peak indicator
		peakIndicator := "  "
		if e.opts.PeakStart != "" && e.opts.PeakEnd != "" {
			if overlapMinutes(t.ScheduledStart, t.ScheduledEnd, e.opts.PeakStart, e.opts.PeakEnd) > 0 {
				peakIndicator = "⚡"
			}
		}

		// Category
		cat := "[S]"
		if t.Category == task.CategoryDeep {
			cat = "[D]"
		}

		// Duration
		duration := formatDuration(taskDurationMinutes(t))

		sb.WriteString(fmt.Sprintf("  %s %s-%s  %s  %s  %s\n",
			peakIndicator,
			t.ScheduledStart,
			t.ScheduledEnd,
			cat,
			t.Description,
			duration))
	}

	return sb.String()
}

// taskDurationMinutes calculates the duration of a task in minutes.
func taskDurationMinutes(t *task.Task) int {
	start, err1 := time.Parse("15:04", t.ScheduledStart)
	end, err2 := time.Parse("15:04", t.ScheduledEnd)
	if err1 != nil || err2 != nil {
		return 0
	}
	return int(end.Sub(start).Minutes())
}

// formatDuration formats minutes as a human-readable duration.
func formatDuration(minutes int) string {
	if minutes == 0 {
		return "0m"
	}
	hours := minutes / 60
	mins := minutes % 60
	if hours == 0 {
		return fmt.Sprintf("%dm", mins)
	}
	if mins == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh%dm", hours, mins)
}

// overlapMinutes calculates overlapping minutes between two time ranges.
func overlapMinutes(start1, end1, start2, end2 string) int {
	s1 := timeToMinutes(start1)
	e1 := timeToMinutes(end1)
	s2 := timeToMinutes(start2)
	e2 := timeToMinutes(end2)

	overlapStart := max(s1, s2)
	overlapEnd := min(e1, e2)

	if overlapEnd <= overlapStart {
		return 0
	}
	return overlapEnd - overlapStart
}

// timeToMinutes converts "HH:MM" to minutes since midnight.
func timeToMinutes(t string) int {
	if len(t) < 5 {
		return 0
	}
	hours := int(t[0]-'0')*10 + int(t[1]-'0')
	mins := int(t[3]-'0')*10 + int(t[4]-'0')
	return hours*60 + mins
}

package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/javiermolinar/sancho/internal/task"
)

// DebugLogger logs TUI state, keystrokes, and events to a file.
type DebugLogger struct {
	mu      sync.Mutex
	file    *os.File
	enabled bool
	seq     int
}

// Global debug logger instance
var debugLog *DebugLogger

// DebugLogPath is the fixed path for debug logs
const DebugLogPath = "sancho-debug.log"

// InitDebugLogger initializes the debug logger if debug mode is enabled.
func InitDebugLogger(enabled bool) error {
	if !enabled {
		debugLog = &DebugLogger{enabled: false}
		return nil
	}

	// Create log file in current directory with fixed name (easy to find)
	logPath := DebugLogPath
	f, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("creating debug log: %w", err)
	}

	debugLog = &DebugLogger{
		file:    f,
		enabled: true,
	}

	debugLog.log("DEBUG_START", map[string]any{
		"log_file": logPath,
		"time":     time.Now().Format(time.RFC3339),
	})

	return nil
}

// CloseDebugLogger closes the debug log file.
func CloseDebugLogger() {
	if debugLog != nil && debugLog.file != nil {
		debugLog.log("DEBUG_END", map[string]any{
			"time": time.Now().Format(time.RFC3339),
		})
		_ = debugLog.file.Close()
	}
}

// log writes a structured log entry.
func (d *DebugLogger) log(event string, data map[string]any) {
	if d == nil || !d.enabled || d.file == nil {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.seq++
	entry := map[string]any{
		"seq":   d.seq,
		"ts":    time.Now().Format("15:04:05.000"),
		"event": event,
	}
	for k, v := range data {
		entry[k] = v
	}

	b, _ := json.Marshal(entry)
	_, _ = fmt.Fprintf(d.file, "%s\n", b)
}

// LogKeyPress logs a key press event.
func LogKeyPress(msg tea.KeyMsg) {
	if debugLog == nil || !debugLog.enabled {
		return
	}
	debugLog.log("KEY_PRESS", map[string]any{
		"key":  msg.String(),
		"type": fmt.Sprintf("%T", msg.Type),
	})
}

// LogModeChange logs a mode change.
func LogModeChange(from, to Mode, reason string) {
	if debugLog == nil || !debugLog.enabled {
		return
	}
	debugLog.log("MODE_CHANGE", map[string]any{
		"from":   modeString(from),
		"to":     modeString(to),
		"reason": reason,
	})
}

// LogCursorMove logs cursor movement.
func LogCursorMove(day, slot int, reason string) {
	if debugLog == nil || !debugLog.enabled {
		return
	}
	debugLog.log("CURSOR_MOVE", map[string]any{
		"day":    day,
		"slot":   slot,
		"reason": reason,
	})
}

// LogSlotState logs the current slot state.
func LogSlotState(sm *SlotStateManager, action string) {
	if debugLog == nil || !debugLog.enabled {
		return
	}

	data := map[string]any{
		"action":     action,
		"is_editing": sm.IsEditing(),
		"is_moving":  sm.IsMoving(),
	}

	if sm.IsMoving() {
		if ms := sm.MoveState(); ms != nil {
			data["move_state"] = map[string]any{
				"source_day":  ms.SourceDay,
				"source_slot": ms.SourceSlot,
				"target_day":  ms.TargetDay,
				"target_slot": ms.TargetSlot,
				"task_id":     ms.MovingTask.ID,
				"task_desc":   truncateStr(ms.MovingTask.Description, 30),
			}
		}
		if mt := sm.MovingTask(); mt != nil {
			data["moving_task"] = map[string]any{
				"id":    mt.ID,
				"desc":  truncateStr(mt.Description, 30),
				"start": mt.ScheduledStart,
				"end":   mt.ScheduledEnd,
			}
		}
	}

	// Log task positions in the grid
	grid := sm.Grid()
	if grid != nil {
		tasks := grid.AllTasks()
		taskPositions := make([]map[string]any, 0, len(tasks))
		seen := make(map[int64]bool)
		for _, t := range tasks {
			if t != nil && !seen[t.ID] {
				seen[t.ID] = true
				day, start, end, found := grid.FindTask(t)
				if found {
					taskPositions = append(taskPositions, map[string]any{
						"id":         t.ID,
						"desc":       truncateStr(t.Description, 20),
						"day":        day,
						"start_slot": start,
						"end_slot":   end,
					})
				}
			}
		}
		data["grid_tasks"] = taskPositions
	}

	debugLog.log("SLOT_STATE", data)
}

// LogWeekWindow logs the current WeekWindow state.
func LogWeekWindow(ww *task.WeekWindow, action string) {
	if debugLog == nil || !debugLog.enabled {
		return
	}
	if ww == nil {
		debugLog.log("WEEK_WINDOW", map[string]any{
			"action": action,
			"status": "nil",
		})
		return
	}

	data := map[string]any{
		"action": action,
	}

	if curr := ww.Current(); curr != nil {
		var tasks []map[string]any
		for dayIdx := 0; dayIdx < 7; dayIdx++ {
			d := curr.Day(dayIdx)
			if d == nil {
				continue
			}
			for _, t := range d.ScheduledTasks() {
				tasks = append(tasks, map[string]any{
					"day":   dayIdx,
					"id":    t.ID,
					"desc":  truncateStr(t.Description, 20),
					"start": t.ScheduledStart,
					"end":   t.ScheduledEnd,
				})
			}
		}
		data["current_week_tasks"] = tasks
	}

	debugLog.log("WEEK_WINDOW", data)
}

// LogRenderCell logs cell rendering info (only for cursor or moving task cells).
func LogRenderCell(day, slot int, timeLabel string, taskDesc string, isCursor bool, isMovingTask bool) {
	if debugLog == nil || !debugLog.enabled {
		return
	}
	// Only log cells that are relevant (cursor or moving task)
	if !isCursor && !isMovingTask {
		return
	}
	debugLog.log("RENDER_CELL", map[string]any{
		"day":            day,
		"slot":           slot,
		"time_label":     timeLabel,
		"task_desc":      taskDesc,
		"is_cursor":      isCursor,
		"is_moving_task": isMovingTask,
	})
}

// LogTaskLookup logs task lookup results.
func LogTaskLookup(day int, timeLabel string, found bool, taskID int64, taskDesc string) {
	if debugLog == nil || !debugLog.enabled {
		return
	}
	debugLog.log("TASK_LOOKUP", map[string]any{
		"day":        day,
		"time_label": timeLabel,
		"found":      found,
		"task_id":    taskID,
		"task_desc":  taskDesc,
	})
}

// LogError logs an error.
func LogError(context string, err error) {
	if debugLog == nil || !debugLog.enabled {
		return
	}
	debugLog.log("ERROR", map[string]any{
		"context": context,
		"error":   err.Error(),
	})
}

// modeString returns a string representation of a Mode.
func modeString(m Mode) string {
	switch m {
	case ModeNormal:
		return "Normal"
	case ModeEdit:
		return "Edit"
	case ModeMove:
		return "Move"
	case ModePrompt:
		return "Prompt"
	case ModeModal:
		return "Modal"
	default:
		return fmt.Sprintf("Unknown(%d)", m)
	}
}

// truncateStr truncates a string to max length.
func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// LogChromeBreakdown logs detailed breakdown of chrome line calculation.
func LogChromeBreakdown(breakdown map[string]int) {
	if debugLog == nil || !debugLog.enabled {
		return
	}
	data := make(map[string]any, len(breakdown))
	for k, v := range breakdown {
		data[k] = v
	}
	debugLog.log("CHROME_BREAKDOWN", data)
}

package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/javiermolinar/sancho/internal/task"
	"github.com/javiermolinar/sancho/internal/tui/commands"
	"github.com/javiermolinar/sancho/internal/tui/input"
)

// handleKeyMsg handles keyboard input.
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Log keystroke
	LogKeyPress(msg)

	// Global keys (work in all modes)
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}
	// Mode-specific handling
	switch m.mode {
	case ModePrompt:
		return m.handlePromptKeys(msg)
	case ModeMove:
		return m.handleMoveKeys(msg)
	case ModeModal:
		return m.handleModalKeys(msg)
	case ModeEdit:
		return m.handleEditKeys(msg)
	default:
		return m.handleNormalKeys(msg)
	}
}

// handleNormalKeys handles keys in normal mode.
func (m Model) handleNormalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	ww := m.slotState.WeekWindow()

	switch msg.String() {
	case "q":
		return m, tea.Quit

	// Navigation
	case "h", "left":
		if m.cursor.Day > 0 {
			m.cursor.Day--
		} else {
			// Move to previous week, Sunday - use cached week if available
			if ww != nil && ww.HasPrevious() {
				m.weekStart = m.weekStart.AddDate(0, 0, -7)
				m.cursor.Day = 6
				m.loading = true
				return m, commands.LoadPrevWeek(m.repo, m.weekStart)
			}
			// Fallback: full reload
			m.weekStart = m.weekStart.AddDate(0, 0, -7)
			m.cursor.Day = 6
			return m, commands.LoadInitialWeeks(m.repo, m.weekStart)
		}
	case "l", "right":
		if m.cursor.Day < 6 {
			m.cursor.Day++
		} else {
			// Move to next week, Monday - use cached week if available
			if ww != nil && ww.HasNext() {
				m.weekStart = m.weekStart.AddDate(0, 0, 7)
				m.cursor.Day = 0
				m.loading = true
				return m, commands.LoadNextWeek(m.repo, m.weekStart)
			}
			// Fallback: full reload
			m.weekStart = m.weekStart.AddDate(0, 0, 7)
			m.cursor.Day = 0
			return m, commands.LoadInitialWeeks(m.repo, m.weekStart)
		}
	case "j", "down":
		m.cursor.Slot = m.nextSlotDown()
		m.ensureCursorVisible()
	case "k", "up":
		m.cursor.Slot = m.nextSlotUp()
		m.ensureCursorVisible()

	// Page navigation
	case "pgdown", "ctrl+d":
		visible := m.visibleRows()
		maxSlot := m.maxSlots() - 1
		m.cursor.Slot = min(maxSlot, m.cursor.Slot+visible)
		m.ensureCursorVisible()
	case "pgup", "ctrl+u":
		visible := m.visibleRows()
		m.cursor.Slot = max(0, m.cursor.Slot-visible)
		m.ensureCursorVisible()

	// Week navigation (jump to prev/next week)
	case "H", "shift+left":
		if ww != nil && ww.HasPrevious() {
			m.weekStart = m.weekStart.AddDate(0, 0, -7)
			m.loading = true
			return m, commands.LoadPrevWeek(m.repo, m.weekStart)
		}
		// Fallback: full reload if no cached prev
		m.weekStart = m.weekStart.AddDate(0, 0, -7)
		return m, commands.LoadInitialWeeks(m.repo, m.weekStart)
	case "L", "shift+right":
		if ww != nil && ww.HasNext() {
			m.weekStart = m.weekStart.AddDate(0, 0, 7)
			m.loading = true
			return m, commands.LoadNextWeek(m.repo, m.weekStart)
		}
		// Fallback: full reload if no cached next
		m.weekStart = m.weekStart.AddDate(0, 0, 7)
		return m, commands.LoadInitialWeeks(m.repo, m.weekStart)

	// Actions
	case "/":
		m.mode = ModePrompt
		m.prompt.SetValue("/")
		m.prompt.Focus()
		m.calculateLayout()
		m.layoutCache = m.buildLayoutCache(m.width, m.height)
		return m, textinput.Blink
	case "p":
		m.mode = ModePrompt
		m.prompt.SetValue("/plan ")
		m.prompt.Focus()
		m.calculateLayout()
		m.layoutCache = m.buildLayoutCache(m.width, m.height)
		return m, textinput.Blink

	case "enter":
		return m.handleEnter()

	case "d":
		return m.handleQuickPostpone()

	// Edit mode entry
	case "i":
		m.slotState.EnterEditMode()
		m.mode = ModeEdit
		m.statusMsg = "Edit mode: g/s/Space/x to modify, y to move, u to undo, Enter to save, Esc to cancel"
		return m, nil

	// These operations require edit mode
	case "y":
		m.statusMsg = "Press i to enter edit mode first"
		return m, nil

	case "g":
		m.statusMsg = "Press i to enter edit mode first"
		return m, nil

	case "s":
		m.statusMsg = "Press i to enter edit mode first"
		return m, nil

	case " ":
		m.statusMsg = "Press i to enter edit mode first"
		return m, nil

	case "x":
		m.statusMsg = "Press i to enter edit mode first"
		return m, nil
	}

	return m, nil
}

// handleEditKeys handles keys in edit mode.
// In edit mode, changes are made in-memory and can be undone.
// Press Enter to save all changes to DB, Esc to discard.
func (m Model) handleEditKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	ww := m.slotState.WeekWindow()

	switch msg.String() {
	case "q":
		// Don't allow quit without confirming changes
		if m.slotState.HasChanges() {
			m.statusMsg = "Unsaved changes! Press Enter to save or Esc to discard"
			return m, nil
		}
		return m, tea.Quit

	// Save changes
	case "enter":
		ctx := context.Background()
		if err := m.slotState.SaveChanges(ctx, m.repo); err != nil {
			m.statusMsg = fmt.Sprintf("Error saving: %v", err)
			return m, nil
		}
		m.mode = ModeNormal
		m.statusMsg = "Changes saved"
		return m, commands.LoadWeek(m.repo, m.weekStart) // Reload to sync with DB

	// Discard changes
	case "esc":
		m.slotState.DiscardChanges()
		m.mode = ModeNormal
		m.statusMsg = "Changes discarded"
		return m, nil

	// Undo last operation
	case "u":
		if !m.slotState.CanUndo() {
			m.statusMsg = "Nothing to undo"
			return m, nil
		}
		if err := m.slotState.Undo(); err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", err)
			return m, nil
		}
		remaining := m.slotState.UndoCount()
		if remaining > 0 {
			m.statusMsg = fmt.Sprintf("Undone (%d more available)", remaining)
		} else {
			m.statusMsg = "Undone (no more changes)"
		}
		return m, nil

	// Navigation
	case "h", "left":
		if m.cursor.Day > 0 {
			m.cursor.Day--
		} else if ww != nil && ww.HasPrevious() {
			m.weekStart = m.weekStart.AddDate(0, 0, -7)
			m.cursor.Day = 6
			m.loading = true
			return m, commands.LoadPrevWeek(m.repo, m.weekStart)
		}
	case "l", "right":
		if m.cursor.Day < 6 {
			m.cursor.Day++
		} else if ww != nil && ww.HasNext() {
			m.weekStart = m.weekStart.AddDate(0, 0, 7)
			m.cursor.Day = 0
			m.loading = true
			return m, commands.LoadNextWeek(m.repo, m.weekStart)
		}
	case "j", "down":
		m.cursor.Slot = m.nextSlotDown()
		m.ensureCursorVisible()
	case "k", "up":
		m.cursor.Slot = m.nextSlotUp()
		m.ensureCursorVisible()

	// Page navigation
	case "pgdown", "ctrl+d":
		visible := m.visibleRows()
		maxSlot := m.maxSlots() - 1
		m.cursor.Slot = min(maxSlot, m.cursor.Slot+visible)
		m.ensureCursorVisible()
	case "pgup", "ctrl+u":
		visible := m.visibleRows()
		m.cursor.Slot = max(0, m.cursor.Slot-visible)
		m.ensureCursorVisible()

	// Edit operations
	case "g":
		return m.handleGrow()

	case "s":
		return m.handleShrink()

	case " ":
		return m.handleSpace()

	case "y":
		return m.handleYank()

	case "x":
		return m.handleRemoveSpace()
	}

	return m, nil
}

// handlePromptKeys handles keys in prompt mode.
func (m Model) handlePromptKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		m.prompt.Blur()
		m.prompt.SetValue("")
		m.calculateLayout()
		m.layoutCache = m.buildLayoutCache(m.width, m.height)
		return m, nil

	case "enter":
		value := m.prompt.Value()
		m.mode = ModeNormal
		m.prompt.Blur()
		m.prompt.SetValue("")
		m.calculateLayout()
		m.layoutCache = m.buildLayoutCache(m.width, m.height)
		return m.handlePromptSubmit(value)

	case "tab":
		if completion, ok := input.PromptAutocomplete(m.prompt.Value(), promptCommands); ok {
			m.prompt.SetValue(completion)
			m.prompt.CursorEnd()
			m.calculateLayout()
			m.layoutCache = m.buildLayoutCache(m.width, m.height)
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.prompt, cmd = m.prompt.Update(msg)
	m.calculateLayout()
	m.layoutCache = m.buildLayoutCache(m.width, m.height)
	return m, cmd
}

// handleMoveKeys handles keys in move mode.
// Uses the new direction-based SlotStateManager for moves.
func (m Model) handleMoveKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		LogModeChange(m.mode, ModeEdit, "move_cancelled")
		m.slotState.CancelMove()
		m.mode = ModeEdit // Return to edit mode, not normal mode
		m.statusMsg = "Move cancelled"
		m.markCacheDirty()
		return m, nil

	case "enter":
		return m.confirmMove()

	// Direction-based moves using SlotStateManager
	case "k", "up":
		if err := m.slotState.MoveUp(); err != nil {
			LogError("MoveUp", err)
			return m, nil
		}
		m.updateCursorToMovingTask()
		LogSlotState(m.slotState, "after_move_up")
		LogCursorMove(m.cursor.Day, m.cursor.Slot, "move_up")
		m.markCacheDirty()
		return m, nil

	case "j", "down":
		if err := m.slotState.MoveDown(); err != nil {
			LogError("MoveDown", err)
			return m, nil
		}
		m.updateCursorToMovingTask()
		LogSlotState(m.slotState, "after_move_down")
		LogCursorMove(m.cursor.Day, m.cursor.Slot, "move_down")
		m.markCacheDirty()
		return m, nil

	case "l", "right":
		if err := m.slotState.MoveRight(); err != nil {
			LogError("MoveRight", err)
			return m, nil
		}
		m.updateCursorToMovingTask()
		LogSlotState(m.slotState, "after_move_right")
		LogCursorMove(m.cursor.Day, m.cursor.Slot, "move_right")
		m.markCacheDirty()
		return m, nil

	case "h", "left":
		if err := m.slotState.MoveLeft(); err != nil {
			LogError("MoveLeft", err)
			return m, nil
		}
		m.updateCursorToMovingTask()
		LogSlotState(m.slotState, "after_move_left")
		LogCursorMove(m.cursor.Day, m.cursor.Slot, "move_left")
		m.markCacheDirty()
		return m, nil
	}

	return m, nil
}

// handleModalKeys handles keys in modal mode.
func (m Model) handleModalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.modalType {
	case ModalTaskForm:
		return m.handleTaskFormKeys(msg)
	case ModalTaskDetail:
		return m.handleTaskDetailKeys(msg)
	case ModalConfirmDelete:
		return m.handleConfirmDeleteKeys(msg)
	case ModalPlanResult:
		return m.handlePlanResultKeys(msg)
	case ModalWeekSummary:
		return m.handleWeekSummaryKeys(msg)
	default:
		if msg.String() == "esc" {
			m.mode = ModeNormal
			m.modalType = ModalNone
			return m, nil
		}
	}
	return m, nil
}

// handleTaskFormKeys handles keys in task form modal.
func (m Model) handleTaskFormKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		m.modalType = ModalNone
		m.modalTask = nil
		m.formDesc.Blur()
		m.formDesc.SetValue("")
		return m, nil

	case "tab":
		if m.modalTask != nil {
			return m, nil
		}
		m.formFocus = (m.formFocus + 1) % 2
		if m.formFocus == 0 {
			m.formDesc.Focus()
		} else {
			m.formDesc.Blur()
		}
		return m, nil

	case "shift+tab":
		if m.modalTask != nil {
			return m, nil
		}
		m.formFocus = (m.formFocus + 1) % 2 // +1 is same as -1 mod 2
		if m.formFocus == 0 {
			m.formDesc.Focus()
		} else {
			m.formDesc.Blur()
		}
		return m, nil

	case "enter":
		if m.formFocus == 0 {
			if m.modalTask != nil {
				return m.saveTaskFromForm()
			}
			// Move to next field
			m.formFocus = 1
			m.formDesc.Blur()
			return m, nil
		}
		// Save the task
		return m.saveTaskFromForm()

	case "left", "h":
		if m.formFocus == 1 {
			if m.formDuration > 0 {
				m.formDuration--
			}
			return m, nil
		}

	case "right", "l":
		if m.formFocus == 1 {
			if m.formDuration < len(durationOptions)-1 {
				m.formDuration++
			}
			return m, nil
		}
	}

	// Handle text input for description field
	if m.formFocus == 0 {
		var cmd tea.Cmd
		m.formDesc, cmd = m.formDesc.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleTaskDetailKeys handles keys in task detail modal.
func (m Model) handleTaskDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter":
		m.mode = ModeNormal
		m.modalType = ModalNone
		m.modalTask = nil
		return m, nil

	case "o":
		// Cycle outcome
		if m.modalTask != nil {
			return m.cycleOutcome()
		}

	case "e":
		if m.modalTask != nil {
			if m.modalTask.IsPast() {
				m.statusMsg = "Cannot edit past tasks"
				return m, nil
			}
			m.modalType = ModalTaskForm
			m.formDesc.SetValue(m.modalTask.Description)
			m.formDesc.Focus()
			m.formFocus = 0
			return m, textinput.Blink
		}

	case "x":
		// Open delete confirmation
		if m.modalTask != nil && !m.modalTask.IsPast() {
			m.modalType = ModalConfirmDelete
			m.confirmMessage = fmt.Sprintf("Cancel task: %s?", m.modalTask.Description)
			return m, nil
		}
		if m.modalTask != nil && m.modalTask.IsPast() {
			m.statusMsg = "Cannot cancel past tasks"
		}
	}
	return m, nil
}

// handleConfirmDeleteKeys handles keys in confirm delete modal.
func (m Model) handleConfirmDeleteKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n":
		// Go back to detail modal if we came from there, otherwise close
		if m.modalTask != nil {
			m.modalType = ModalTaskDetail
		} else {
			m.mode = ModeNormal
			m.modalType = ModalNone
		}
		return m, nil

	case "enter", "y":
		if m.modalTask != nil {
			// Delete the task
			ctx := context.Background()
			if err := m.repo.CancelTask(ctx, m.modalTask.ID); err != nil {
				m.statusMsg = fmt.Sprintf("Error: %v", err)
			} else {
				m.statusMsg = fmt.Sprintf("Cancelled: %s", m.modalTask.Description)
			}
			m.modalTask = nil
			m.mode = ModeNormal
			m.modalType = ModalNone
			return m, commands.LoadWeek(m.repo, m.weekStart)
		}
	}
	return m, nil
}

// handlePlanResultKeys handles keys in plan result modal.
func (m Model) handlePlanResultKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "c":
		// Cancel planning
		m.mode = ModeNormal
		m.modalType = ModalNone
		m.planResult = nil
		m.planner = nil
		m.statusMsg = "Planning cancelled"
		return m, nil

	case "enter", "a":
		// Accept and save
		if m.planResult == nil {
			return m, nil
		}
		if m.planResult.HasValidationErrors() {
			m.statusMsg = "Cannot save: validation errors present"
			return m, nil
		}
		return m, commands.SavePlan(m.planner, m.planResult)

	case "m":
		// Amend - ask for feedback
		m.mode = ModePrompt
		m.modalType = ModalNone
		m.prompt.SetValue("")
		m.prompt.Focus()
		m.statusMsg = "What would you like to amend?"
		return m, textinput.Blink
	}
	return m, nil
}

// saveTaskFromForm creates a new task from the form data.
func (m Model) saveTaskFromForm() (tea.Model, tea.Cmd) {
	desc := strings.TrimSpace(m.formDesc.Value())
	if desc == "" {
		m.statusMsg = "Description is required"
		return m, nil
	}

	if m.modalTask != nil {
		if m.modalTask.IsPast() {
			m.statusMsg = "Cannot edit past tasks"
			return m, nil
		}

		ctx := context.Background()
		if err := m.repo.UpdateTaskDescription(ctx, m.modalTask.ID, desc); err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", err)
			return m, nil
		}

		m.modalTask.Description = desc
		m.formDesc.SetValue("")
		m.formDesc.Blur()
		m.formFocus = 0
		m.modalTask = nil
		m.mode = ModeNormal
		m.modalType = ModalNone
		m.statusMsg = fmt.Sprintf("Updated: %s", desc)
		return m, commands.LoadWeek(m.repo, m.weekStart)
	}

	// Calculate task times
	taskDate := m.weekStart.AddDate(0, 0, m.cursor.Day)
	startTime := m.slotToTime(m.cursor.Slot)
	duration := durationOptions[m.formDuration]
	endTime := addMinutesToTime(startTime, duration)

	// Determine category
	category := task.CategoryDeep
	if m.formCategory == 1 {
		category = task.CategoryShallow
	}

	// Create the task
	newTask := &task.Task{
		Description:    desc,
		Category:       category,
		ScheduledDate:  taskDate,
		ScheduledStart: startTime,
		ScheduledEnd:   endTime,
		Status:         task.StatusScheduled,
	}

	ctx := context.Background()
	if err := m.repo.CreateTask(ctx, newTask); err != nil {
		m.statusMsg = fmt.Sprintf("Error: %v", err)
		return m, nil
	}

	// Clear form and close modal
	m.formDesc.SetValue("")
	m.formDesc.Blur()
	m.formCategory = 0
	m.formDuration = 1
	m.formFocus = 0
	m.mode = ModeNormal
	m.modalType = ModalNone
	m.statusMsg = fmt.Sprintf("Created: %s", desc)

	return m, commands.LoadWeek(m.repo, m.weekStart)
}

// cycleOutcome cycles through task outcomes.
func (m Model) cycleOutcome() (tea.Model, tea.Cmd) {
	if m.modalTask == nil {
		return m, nil
	}

	// Cycle: nil -> on_time -> over -> under -> nil
	var newOutcome task.Outcome
	if m.modalTask.Outcome == nil {
		newOutcome = task.OutcomeOnTime
	} else {
		switch *m.modalTask.Outcome {
		case task.OutcomeOnTime:
			newOutcome = task.OutcomeOver
		case task.OutcomeOver:
			newOutcome = task.OutcomeUnder
		default:
			// Reset to no outcome - for now just cycle back to on_time
			newOutcome = task.OutcomeOnTime
		}
	}

	ctx := context.Background()
	if err := m.repo.SetTaskOutcome(ctx, m.modalTask.ID, newOutcome); err != nil {
		m.statusMsg = fmt.Sprintf("Error: %v", err)
		return m, nil
	}

	m.modalTask.Outcome = &newOutcome
	m.statusMsg = fmt.Sprintf("Outcome: %s", newOutcome)
	return m, commands.LoadWeek(m.repo, m.weekStart)
}

// handleEnter handles Enter key press.
func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	t := m.taskAtCursor()
	if t == nil {
		// Empty slot - open new task modal
		m.mode = ModeModal
		m.modalType = ModalTaskForm
		m.modalTask = nil
		m.formDesc.SetValue("")
		m.formDesc.Focus()
		m.formCategory = 0 // deep
		m.formDuration = 1 // 30 min
		m.formFocus = 0
		return m, textinput.Blink
	}

	// Task exists - open detail popup
	m.mode = ModeModal
	m.modalType = ModalTaskDetail
	m.modalTask = t
	return m, nil
}

// handleYank enters move mode.
func (m Model) handleYank() (tea.Model, tea.Cmd) {
	t := m.taskAtCursor()
	if t == nil {
		m.statusMsg = "No task to move"
		return m, nil
	}

	if t.IsPast() {
		m.statusMsg = "Cannot move past tasks"
		return m, nil
	}

	// Start move in SlotStateManager
	if err := m.slotState.StartMove(t); err != nil {
		m.statusMsg = fmt.Sprintf("Error: %v", err)
		LogError("StartMove", err)
		return m, nil
	}

	LogModeChange(m.mode, ModeMove, "yank_task")
	LogSlotState(m.slotState, "start_move")
	LogCursorMove(m.cursor.Day, m.cursor.Slot, "yank_start")

	m.mode = ModeMove
	m.moveOriginalDay = m.cursor.Day
	m.moveOriginalSlot = m.cursor.Slot
	m.statusMsg = fmt.Sprintf("Moving: %s (jk to move up/down, l to next day, Enter to confirm, Esc to cancel)", t.Description)
	m.markCacheDirty()
	return m, nil
}

// confirmMove confirms the move operation.
func (m Model) confirmMove() (tea.Model, tea.Cmd) {
	if !m.slotState.IsMoving() {
		m.mode = ModeEdit
		return m, nil
	}

	// Get the moving task description before confirming
	movingTask := m.slotState.MovingTask()
	description := ""
	if movingTask != nil {
		description = movingTask.Description
	}

	// Confirm the move in SlotStateManager
	if err := m.slotState.ConfirmMove(); err != nil {
		m.statusMsg = fmt.Sprintf("Error: %v", err)
		m.mode = ModeEdit
		return m, nil
	}

	m.mode = ModeEdit
	m.statusMsg = "Moved: " + description
	m.markCacheDirty()
	return m, nil
}

// handleQuickPostpone postpones to next working day.
func (m Model) handleQuickPostpone() (tea.Model, tea.Cmd) {
	t := m.taskAtCursor()
	if t == nil {
		m.statusMsg = "No task to postpone"
		return m, nil
	}

	if t.IsPast() {
		m.statusMsg = "Cannot postpone past tasks"
		return m, nil
	}

	// Find next working day
	nextDay := t.ScheduledDate.AddDate(0, 0, 1)
	for !m.isWorkday(nextDay) {
		nextDay = nextDay.AddDate(0, 0, 1)
	}

	ctx := context.Background()
	_, err := m.repo.PostponeTask(ctx, t.ID, nextDay, t.ScheduledStart, t.ScheduledEnd)
	if err != nil {
		return m, func() tea.Msg { return commands.ErrMsg{Err: err} }
	}

	m.statusMsg = fmt.Sprintf("Postponed to %s", nextDay.Format("Mon Jan 2"))
	return m, commands.LoadWeek(m.repo, m.weekStart)
}

// handleGrow grows task by 15 minutes (edit mode only).
func (m Model) handleGrow() (tea.Model, tea.Cmd) {
	t := m.taskAtCursor()
	if t == nil {
		m.statusMsg = "No task to grow"
		return m, nil
	}

	if t.IsPast() {
		m.statusMsg = "Cannot modify past tasks"
		return m, nil
	}

	// Use SlotStateManager for grow
	if err := m.slotState.Grow(t); err != nil {
		m.statusMsg = fmt.Sprintf("Error: %v", err)
		return m, nil
	}

	m.statusMsg = fmt.Sprintf("Grew: %s", t.Description)
	m.markCacheDirty()
	return m, nil
}

// handleShrink shrinks task by 15 minutes (edit mode only).
func (m Model) handleShrink() (tea.Model, tea.Cmd) {
	t := m.taskAtCursor()
	if t == nil {
		m.statusMsg = "No task to shrink"
		return m, nil
	}

	if t.IsPast() {
		m.statusMsg = "Cannot modify past tasks"
		return m, nil
	}

	// Use SlotStateManager for shrink
	if err := m.slotState.Shrink(t); err != nil {
		if errors.Is(err, ErrMinimumSlotsDuration) {
			m.statusMsg = "Cannot shrink below 15 minutes"
		} else {
			m.statusMsg = fmt.Sprintf("Error: %v", err)
		}
		return m, nil
	}

	m.statusMsg = fmt.Sprintf("Shrunk: %s", t.Description)
	m.focusCursorOnTaskEnd(t)
	m.markCacheDirty()
	return m, nil
}

// handleSpace adds a 15-minute gap after the current task (edit mode only).
func (m Model) handleSpace() (tea.Model, tea.Cmd) {
	t := m.taskAtCursor()
	if t == nil {
		dayIndex := WeekAndDayToDayIndex(1, m.cursor.Day)
		slot := m.displaySlotToSlot(m.cursor.Slot)
		if err := m.slotState.AddSpaceAt(dayIndex, slot); err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", err)
			return m, nil
		}
		m.statusMsg = "Added space"
		m.markCacheDirty()
		return m, nil
	}

	// Use SlotStateManager for add space
	if err := m.slotState.AddSpaceAfter(t); err != nil {
		m.statusMsg = fmt.Sprintf("Error: %v", err)
		return m, nil
	}

	m.statusMsg = fmt.Sprintf("Added space: %s", t.Description)
	m.markCacheDirty()
	return m, nil
}

// handleRemoveSpace removes a 15-minute gap at the cursor (edit mode only).
func (m Model) handleRemoveSpace() (tea.Model, tea.Cmd) {
	t := m.taskAtCursor()
	if t != nil {
		m.statusMsg = "No gap to remove"
		return m, nil
	}

	dayIndex := WeekAndDayToDayIndex(1, m.cursor.Day)
	slot := m.displaySlotToSlot(m.cursor.Slot)
	if err := m.slotState.RemoveSpaceAt(dayIndex, slot); err != nil {
		if errors.Is(err, ErrNoGapToRemove) {
			m.statusMsg = "No gap to remove"
			return m, nil
		}
		if errors.Is(err, ErrTaskAlreadyStarted) {
			m.statusMsg = "Cannot modify past tasks"
			return m, nil
		}
		m.statusMsg = fmt.Sprintf("Error: %v", err)
		return m, nil
	}

	m.statusMsg = "Removed space"
	m.markCacheDirty()
	return m, nil
}

// handlePromptSubmit processes the submitted prompt.
func (m Model) handlePromptSubmit(value string) (tea.Model, tea.Cmd) {
	if value == "" {
		return m, nil
	}

	if strings.HasPrefix(value, "/") {
		fields := strings.Fields(value)
		if len(fields) == 0 {
			return m, nil
		}
		switch fields[0] {
		case "/plan":
			input := strings.TrimSpace(strings.TrimPrefix(value, "/plan"))
			if input == "" {
				m.statusMsg = "Plan requires input"
				return m, nil
			}
			m.planInput = input
			m.statusMsg = "Planning..."
			return m, commands.Plan(input, m.config, m.repo)
		case "/help":
			m.statusMsg = "Commands: /plan, /week, /help, /reflect"
			return m, nil
		case "/reflect":
			m.statusMsg = "Reflect is not implemented yet"
			return m, nil
		case "/week":
			m.statusMsg = "Summarizing..."
			return m, commands.WeekSummary(m.config, m.repo, m.weekStart)
		default:
			m.statusMsg = fmt.Sprintf("Unknown command: %s", fields[0])
			return m, nil
		}
	}

	m.planInput = value
	m.statusMsg = "Planning..."
	return m, commands.Plan(value, m.config, m.repo)
}

func (m Model) handleWeekSummaryKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "w":
		m.weekSummaryView = weekSummaryViewTasks
		return m, nil
	case "s":
		m.weekSummaryView = weekSummaryViewSummary
		return m, nil
	case "y":
		if m.weekSummary == nil || len(m.weekSummary.Tasks) == 0 {
			m.statusMsg = "No tasks to copy"
			return m, nil
		}
		if err := clipboard.WriteAll(m.weekSummaryCopyText); err != nil {
			m.statusMsg = fmt.Sprintf("Copy failed: %v", err)
			return m, nil
		}
		m.statusMsg = "Copied week tasks"
		return m, nil
	case "esc", "enter":
		m.mode = ModeNormal
		m.modalType = ModalNone
		m.weekSummary = nil
		m.weekSummarySummaryText = nil
		m.weekSummaryTasksText = nil
		m.weekSummaryCopyText = ""
		return m, nil
	}
	return m, nil
}

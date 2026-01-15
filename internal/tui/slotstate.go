package tui

import (
	"context"
	"errors"
	"time"

	"github.com/javiermolinar/sancho/internal/task"
)

// SlotStateManager errors.
var (
	ErrSlotNotInEditMode = errors.New("not in edit mode")
	ErrSlotNothingToUndo = errors.New("nothing to undo")
	ErrSlotNotMoving     = errors.New("not in move mode")
	ErrSlotAlreadyMoving = errors.New("already moving a task")
)

const (
	slotDefaultMaxHistory = 50
)

// SlotHistoryEntry represents a single undo-able operation.
type SlotHistoryEntry struct {
	Description string    // e.g., "Move: task name"
	Grid        *SlotGrid // The grid state before the operation
}

// SlotMoveState contains information needed to render move mode.
type SlotMoveState struct {
	MovingTask   *task.Task   // The task being moved
	SourceDay    int          // Original day of the moving task
	SourceSlot   int          // Original start slot
	TargetDay    int          // Current target day
	TargetSlot   int          // Current target slot
	ShiftedTasks []*task.Task // Tasks that have been shifted
}

// SlotStateManager manages SlotGrid state with history and edit mode support.
// It provides an immutable approach where each operation creates a new grid.
type SlotStateManager struct {
	config SlotConfig

	// Saved state (synced with DB)
	savedGrid *SlotGrid

	// Working state (in-memory edits during edit mode)
	workingGrid *SlotGrid

	// Edit mode flag
	editing bool

	// Undo history (only used during edit mode)
	// Stores references to immutable grids - very efficient!
	history    []SlotHistoryEntry
	maxHistory int

	// Track which days have been modified (for efficient persistence)
	dirtyDays map[int]bool

	// Move session state
	isMoving         bool
	moveSourceDay    int
	moveSourceSlot   int
	moveNumSlots     int
	movingTask       *task.Task
	beforeMoveGrid   *SlotGrid // Grid state before move started
	currentMoveState *SlotMoveState
}

// NewSlotStateManager creates a new slot state manager.
func NewSlotStateManager(config SlotConfig) *SlotStateManager {
	return &SlotStateManager{
		config:     config,
		maxHistory: slotDefaultMaxHistory,
		dirtyDays:  make(map[int]bool),
	}
}

// Config returns the slot configuration.
func (sm *SlotStateManager) Config() SlotConfig {
	return sm.config
}

// UpdateConfig updates the slot configuration.
// This is needed when navigating between weeks to update the FirstDate.
func (sm *SlotStateManager) UpdateConfig(config SlotConfig) {
	sm.config = config
}

// Grid returns the current grid for rendering.
// Returns workingGrid if editing, savedGrid otherwise.
func (sm *SlotStateManager) Grid() *SlotGrid {
	if sm.editing && sm.workingGrid != nil {
		return sm.workingGrid
	}
	return sm.savedGrid
}

// SavedGrid returns the saved (immutable) grid.
func (sm *SlotStateManager) SavedGrid() *SlotGrid {
	return sm.savedGrid
}

// SetGrid sets the saved grid (after loading from DB).
func (sm *SlotStateManager) SetGrid(grid *SlotGrid) {
	sm.savedGrid = grid
	if sm.editing {
		sm.workingGrid = grid.clone()
	}
}

// IsEditing returns true if in edit mode.
func (sm *SlotStateManager) IsEditing() bool {
	return sm.editing
}

// EnterEditMode starts an edit session.
func (sm *SlotStateManager) EnterEditMode() {
	if sm.editing {
		return
	}
	sm.editing = true
	sm.workingGrid = sm.savedGrid.clone()
	sm.history = nil
	sm.dirtyDays = make(map[int]bool)
}

// DiscardChanges exits edit mode and reverts all changes.
func (sm *SlotStateManager) DiscardChanges() {
	sm.editing = false
	sm.workingGrid = nil
	sm.history = nil
	sm.dirtyDays = make(map[int]bool)
	sm.clearMoveState()
}

// HasChanges returns true if there are unsaved modifications.
func (sm *SlotStateManager) HasChanges() bool {
	return len(sm.dirtyDays) > 0
}

// DirtyDays returns the set of day indices that have been modified.
func (sm *SlotStateManager) DirtyDays() map[int]bool {
	result := make(map[int]bool, len(sm.dirtyDays))
	for k, v := range sm.dirtyDays {
		result[k] = v
	}
	return result
}

// CanUndo returns true if there are operations to undo.
func (sm *SlotStateManager) CanUndo() bool {
	return sm.editing && len(sm.history) > 0
}

// UndoCount returns the number of operations that can be undone.
func (sm *SlotStateManager) UndoCount() int {
	return len(sm.history)
}

// Undo reverts the last operation.
func (sm *SlotStateManager) Undo() error {
	if !sm.editing {
		return ErrSlotNotInEditMode
	}
	if len(sm.history) == 0 {
		return ErrSlotNothingToUndo
	}

	// Pop the last entry
	entry := sm.history[len(sm.history)-1]
	sm.history = sm.history[:len(sm.history)-1]

	// Restore the grid from the snapshot (just use the reference!)
	sm.workingGrid = entry.Grid

	// Recalculate dirty days
	if len(sm.history) == 0 {
		sm.dirtyDays = make(map[int]bool)
	}

	return nil
}

// pushHistory saves the current state before a modification.
func (sm *SlotStateManager) pushHistory(description string) {
	if len(sm.history) >= sm.maxHistory {
		// Remove oldest entry
		sm.history = sm.history[1:]
	}

	// Store reference to current grid (immutable, so this is safe and cheap!)
	sm.history = append(sm.history, SlotHistoryEntry{
		Description: description,
		Grid:        sm.workingGrid,
	})
}

// markDayDirty marks a day as modified.
func (sm *SlotStateManager) markDayDirty(dayIndex int) {
	sm.dirtyDays[dayIndex] = true
}

// clearMoveState clears all move-related state.
func (sm *SlotStateManager) clearMoveState() {
	sm.isMoving = false
	sm.moveSourceDay = 0
	sm.moveSourceSlot = 0
	sm.moveNumSlots = 0
	sm.movingTask = nil
	sm.beforeMoveGrid = nil
	sm.currentMoveState = nil
}

// IsMoving returns true if currently in move mode.
func (sm *SlotStateManager) IsMoving() bool {
	return sm.isMoving
}

// MoveState returns the current move state for rendering.
func (sm *SlotStateManager) MoveState() *SlotMoveState {
	return sm.currentMoveState
}

// StartMove begins moving a task.
func (sm *SlotStateManager) StartMove(t *task.Task) error {
	if !sm.editing {
		return ErrSlotNotInEditMode
	}
	if sm.isMoving {
		return ErrSlotAlreadyMoving
	}
	if t == nil {
		return ErrSlotTaskNotFound
	}

	// Find the task in the grid
	day, startSlot, endSlot, found := sm.workingGrid.FindTask(t)
	if !found {
		return ErrSlotTaskNotFound
	}

	// Check if task can be modified (not in past)
	if err := sm.workingGrid.canModifyTask(t); err != nil {
		return err
	}

	// Store move session state
	sm.isMoving = true
	sm.moveSourceDay = day
	sm.moveSourceSlot = startSlot
	sm.moveNumSlots = endSlot - startSlot
	sm.movingTask = t
	sm.beforeMoveGrid = sm.workingGrid // Store reference (immutable!)

	// Initialize move state for rendering
	sm.currentMoveState = &SlotMoveState{
		MovingTask:   t,
		SourceDay:    day,
		SourceSlot:   startSlot,
		TargetDay:    day,
		TargetSlot:   startSlot,
		ShiftedTasks: nil,
	}

	return nil
}

// MoveUp moves the task to earlier time on the same day.
// Each call accumulates on workingGrid. Use CancelMove to revert all moves.
func (sm *SlotStateManager) MoveUp() error {
	if !sm.isMoving {
		return ErrSlotNotMoving
	}

	newGrid, err := sm.workingGrid.MoveUp(sm.movingTask)
	if err != nil {
		return err
	}

	sm.workingGrid = newGrid
	sm.updateMoveState()
	return nil
}

// MoveDown moves the task to later time on the same day.
// Each call accumulates on workingGrid. Use CancelMove to revert all moves.
func (sm *SlotStateManager) MoveDown() error {
	if !sm.isMoving {
		return ErrSlotNotMoving
	}

	newGrid, err := sm.workingGrid.MoveDown(sm.movingTask)
	if err != nil {
		return err
	}

	sm.workingGrid = newGrid
	sm.updateMoveState()
	return nil
}

// MoveRight moves the task to the next day at the same slot.
// Each call accumulates on workingGrid. Use CancelMove to revert all moves.
func (sm *SlotStateManager) MoveRight() error {
	if !sm.isMoving {
		return ErrSlotNotMoving
	}

	newGrid, err := sm.workingGrid.MoveRight(sm.movingTask)
	if err != nil {
		return err
	}

	sm.workingGrid = newGrid
	sm.updateMoveState()
	return nil
}

// MoveLeft moves the task to the previous day at the same slot.
// Each call accumulates on workingGrid. Use CancelMove to revert all moves.
func (sm *SlotStateManager) MoveLeft() error {
	if !sm.isMoving {
		return ErrSlotNotMoving
	}

	newGrid, err := sm.workingGrid.MoveLeft(sm.movingTask)
	if err != nil {
		return err
	}

	sm.workingGrid = newGrid
	sm.updateMoveState()
	return nil
}

// updateMoveState updates the move state after a direction-based move.
func (sm *SlotStateManager) updateMoveState() {
	// Find current position of the moving task
	targetDay, targetSlot, _, found := sm.workingGrid.FindTask(sm.movingTask)
	if !found {
		return
	}

	// Find shifted tasks (tasks that moved from their original position)
	shiftedTasks := sm.findShiftedTasks()

	// Update move state for rendering
	sm.currentMoveState = &SlotMoveState{
		MovingTask:   sm.movingTask,
		SourceDay:    sm.moveSourceDay,
		SourceSlot:   sm.moveSourceSlot,
		TargetDay:    targetDay,
		TargetSlot:   targetSlot,
		ShiftedTasks: shiftedTasks,
	}
}

// findShiftedTasks returns tasks whose positions differ between beforeMoveGrid and workingGrid.
func (sm *SlotStateManager) findShiftedTasks() []*task.Task {
	if sm.beforeMoveGrid == nil || sm.workingGrid == nil {
		return nil
	}

	var shifted []*task.Task
	seen := make(map[int64]bool)

	// Compare each task's position
	for _, t := range sm.workingGrid.AllTasks() {
		if t.ID == sm.movingTask.ID {
			continue // Skip the moving task itself
		}
		if seen[t.ID] {
			continue
		}
		seen[t.ID] = true

		// Find position in both grids
		beforeDay, beforeSlot, _, foundBefore := sm.beforeMoveGrid.FindTask(t)
		afterDay, afterSlot, _, foundAfter := sm.workingGrid.FindTask(t)

		if foundBefore && foundAfter {
			if beforeDay != afterDay || beforeSlot != afterSlot {
				shifted = append(shifted, t)
			}
		}
	}

	return shifted
}

// ConfirmMove commits the move operation.
func (sm *SlotStateManager) ConfirmMove() error {
	if !sm.isMoving {
		return ErrSlotNotMoving
	}

	// Push history (before the move)
	sm.pushHistory("Move: " + sm.movingTask.Description)

	// Mark affected days as dirty
	sm.markDayDirty(sm.moveSourceDay)
	if sm.currentMoveState != nil && sm.currentMoveState.TargetDay != sm.moveSourceDay {
		sm.markDayDirty(sm.currentMoveState.TargetDay)
	}

	// Clear move state (workingGrid already has the new state)
	sm.clearMoveState()

	return nil
}

// CancelMove aborts the move operation and restores the original state.
func (sm *SlotStateManager) CancelMove() {
	if !sm.isMoving {
		return
	}

	// Restore grid from before move
	sm.workingGrid = sm.beforeMoveGrid

	// Clear move state
	sm.clearMoveState()
}

// Grow extends a task by one slot, shifting subsequent tasks if needed.
func (sm *SlotStateManager) Grow(t *task.Task) error {
	if !sm.editing {
		return ErrSlotNotInEditMode
	}
	if t == nil {
		return ErrSlotTaskNotFound
	}

	// Check if task can be modified
	if err := sm.workingGrid.canModifyTask(t); err != nil {
		return err
	}

	// Push history before modification
	sm.pushHistory("Grow: " + t.Description)

	// Perform grow
	newGrid, err := sm.workingGrid.Grow(t)
	if err != nil {
		// Pop history since operation failed
		sm.history = sm.history[:len(sm.history)-1]
		return err
	}

	// Find which day was affected
	day, _, _, found := sm.workingGrid.FindTask(t)
	if found {
		sm.markDayDirty(day)
	}

	sm.workingGrid = newGrid
	return nil
}

// Shrink reduces a task by one slot.
func (sm *SlotStateManager) Shrink(t *task.Task) error {
	if !sm.editing {
		return ErrSlotNotInEditMode
	}
	if t == nil {
		return ErrSlotTaskNotFound
	}

	// Check if task can be modified
	if err := sm.workingGrid.canModifyTask(t); err != nil {
		return err
	}

	// Push history before modification
	sm.pushHistory("Shrink: " + t.Description)

	// Perform shrink
	newGrid, err := sm.workingGrid.Shrink(t)
	if err != nil {
		// Pop history since operation failed
		sm.history = sm.history[:len(sm.history)-1]
		return err
	}

	// Find which day was affected
	day, _, _, found := sm.workingGrid.FindTask(t)
	if found {
		sm.markDayDirty(day)
	}

	sm.workingGrid = newGrid
	return nil
}

// AddSpace inserts one empty slot after a task, shifting subsequent tasks.
func (sm *SlotStateManager) AddSpace(t *task.Task) error {
	return sm.AddSpaceAfter(t)
}

// AddSpaceAfter inserts one empty slot after a task, shifting subsequent tasks.
func (sm *SlotStateManager) AddSpaceAfter(t *task.Task) error {
	if !sm.editing {
		return ErrSlotNotInEditMode
	}
	if t == nil {
		return ErrSlotTaskNotFound
	}

	// Push history before modification
	sm.pushHistory("AddSpace: " + t.Description)

	// Perform add space
	newGrid, err := sm.workingGrid.AddSpace(t)
	if err != nil {
		// Pop history since operation failed
		sm.history = sm.history[:len(sm.history)-1]
		return err
	}

	// Find which day was affected
	day, _, _, found := sm.workingGrid.FindTask(t)
	if found {
		sm.markDayDirty(day)
	}

	sm.workingGrid = newGrid
	return nil
}

// AddSpaceAt inserts one empty slot at the given day/slot, shifting subsequent slots.
func (sm *SlotStateManager) AddSpaceAt(day, slot int) error {
	if !sm.editing {
		return ErrSlotNotInEditMode
	}

	// Push history before modification
	sm.pushHistory("AddSpace: empty slot")

	newGrid, err := sm.workingGrid.AddSpaceAt(day, slot)
	if err != nil {
		sm.history = sm.history[:len(sm.history)-1]
		return err
	}

	sm.markDayDirty(day)
	sm.workingGrid = newGrid
	return nil
}

// RemoveSpaceAt removes one empty slot at the given day/slot, shifting subsequent slots.
func (sm *SlotStateManager) RemoveSpaceAt(day, slot int) error {
	if !sm.editing {
		return ErrSlotNotInEditMode
	}

	newGrid, err := sm.workingGrid.RemoveSpaceAt(day, slot)
	if err != nil {
		return err
	}

	sm.pushHistory("RemoveSpace: empty slot")
	sm.markDayDirty(day)
	sm.workingGrid = newGrid
	return nil
}

// Delete removes a task from the grid, shifting subsequent tasks left.
func (sm *SlotStateManager) Delete(t *task.Task) error {
	if !sm.editing {
		return ErrSlotNotInEditMode
	}
	if t == nil {
		return ErrSlotTaskNotFound
	}

	// Find which day will be affected (before deletion)
	day, _, _, found := sm.workingGrid.FindTask(t)
	if !found {
		return ErrSlotTaskNotFound
	}

	// Push history before modification
	sm.pushHistory("Delete: " + t.Description)

	// Perform delete
	newGrid, err := sm.workingGrid.Delete(t)
	if err != nil {
		// Pop history since operation failed
		sm.history = sm.history[:len(sm.history)-1]
		return err
	}

	sm.markDayDirty(day)
	sm.workingGrid = newGrid
	return nil
}

// IsTaskShifted returns true if a task has been shifted from its original position.
// Used for rendering shifted tasks differently in move mode.
func (sm *SlotStateManager) IsTaskShifted(taskID int64) bool {
	if sm.currentMoveState == nil {
		return false
	}
	for _, t := range sm.currentMoveState.ShiftedTasks {
		if t.ID == taskID {
			return true
		}
	}
	return false
}

// CommitChanges should be called after saving to DB.
// Updates savedGrid to match workingGrid and exits edit mode.
func (sm *SlotStateManager) CommitChanges() {
	if !sm.editing {
		return
	}

	sm.savedGrid = sm.workingGrid
	sm.editing = false
	sm.workingGrid = nil
	sm.history = nil
	sm.dirtyDays = make(map[int]bool)
	sm.clearMoveState()
}

// ============================================================================
// Repository Integration
// ============================================================================

// SaveChanges persists all modifications to the database using the provided repository.
// It extracts changed tasks from the grid and updates them in the database.
// After saving, it exits edit mode and updates the saved grid.
func (sm *SlotStateManager) SaveChanges(ctx context.Context, repo task.Repository) error {
	if !sm.editing {
		return nil
	}

	// Get the changes between saved and working grids
	changes := GetChangedTasks(sm.savedGrid, sm.workingGrid)

	// Group updates by the NEW date (after move)
	updatesByDate := make(map[string][]task.TaskTimeUpdate)
	dateMap := make(map[string]time.Time)

	for _, t := range changes.UpdatedTasks {
		dateKey := t.ScheduledDate.Format("2006-01-02")
		updatesByDate[dateKey] = append(updatesByDate[dateKey], task.TaskTimeUpdate{
			ID:       t.ID,
			NewStart: t.ScheduledStart,
			NewEnd:   t.ScheduledEnd,
		})
		dateMap[dateKey] = t.ScheduledDate
	}

	// Persist each day's updates
	for dateKey, updates := range updatesByDate {
		if len(updates) == 0 {
			continue
		}
		date := dateMap[dateKey]
		if err := repo.BatchUpdateTaskTimes(ctx, date, updates); err != nil {
			return err
		}
	}

	// Update saved state and exit edit mode
	sm.savedGrid = sm.workingGrid
	sm.editing = false
	sm.workingGrid = nil
	sm.history = nil
	sm.dirtyDays = make(map[int]bool)
	sm.clearMoveState()

	return nil
}

// MovingTask returns the task currently being moved.
// Returns nil if not in move mode.
func (sm *SlotStateManager) MovingTask() *task.Task {
	if !sm.isMoving || sm.movingTask == nil {
		return nil
	}
	return sm.movingTask
}

// TaskAt returns the task at the given day and slot position.
// Returns nil if the position is empty or out of bounds.
func (sm *SlotStateManager) TaskAt(day, slot int) *task.Task {
	grid := sm.Grid()
	if grid == nil {
		return nil
	}
	return grid.TaskAt(day, slot)
}

// TasksOnDay returns all tasks on a specific day.
func (sm *SlotStateManager) TasksOnDay(day int) []*task.Task {
	grid := sm.Grid()
	if grid == nil {
		return nil
	}
	return grid.TasksOnDay(day)
}

// AllTasks returns all tasks in the current grid.
func (sm *SlotStateManager) AllTasks() []*task.Task {
	grid := sm.Grid()
	if grid == nil {
		return nil
	}
	return grid.AllTasks()
}

// FindTask returns the position of a task in the grid.
// Returns day, startSlot, endSlot (exclusive), and found.
func (sm *SlotStateManager) FindTask(t *task.Task) (day, startSlot, endSlot int, found bool) {
	grid := sm.Grid()
	if grid == nil {
		return 0, 0, 0, false
	}
	return grid.FindTask(t)
}

// FindTaskByID returns the position of a task by ID.
// Returns day, startSlot, endSlot (exclusive), and found.
func (sm *SlotStateManager) FindTaskByID(id int64) (day, startSlot, endSlot int, found bool) {
	grid := sm.Grid()
	if grid == nil {
		return 0, 0, 0, false
	}
	return grid.FindTaskByID(id)
}

// WeekWindow converts the current grid to a WeekWindow for rendering.
// This provides compatibility with views that expect WeekWindow format.
func (sm *SlotStateManager) WeekWindow() *task.WeekWindow {
	return SlotGridToWeekWindow(sm.Grid())
}

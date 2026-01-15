package tui

import (
	"testing"
	"time"

	"github.com/javiermolinar/sancho/internal/task"
)

// stateTestConfig creates a config suitable for state manager tests.
func stateTestConfig() SlotConfig {
	futureDate := time.Now().AddDate(0, 0, 7).Truncate(24 * time.Hour)
	return SlotConfig{
		FirstDate:         futureDate,
		NumDays:           7,
		SlotDuration:      15,
		WorkingHoursStart: 9 * 60,  // 09:00
		WorkingHoursEnd:   17 * 60, // 17:00
		Now: func() time.Time {
			// Return a time before the grid starts, so nothing is "past"
			return futureDate.Add(-24 * time.Hour)
		},
	}
}

// getTaskByID returns the task with the given ID from the grid.
func getTaskByID(grid *SlotGrid, id int64) *task.Task {
	for _, t := range grid.AllTasks() {
		if t.ID == id {
			return t
		}
	}
	return nil
}

func TestSlotStateManager_NewAndBasics(t *testing.T) {
	cfg := stateTestConfig()
	sm := NewSlotStateManager(cfg)

	if sm.IsEditing() {
		t.Error("new state manager should not be in edit mode")
	}
	if sm.IsMoving() {
		t.Error("new state manager should not be in move mode")
	}
	if sm.Grid() != nil {
		t.Error("new state manager should have nil grid")
	}
	if sm.HasChanges() {
		t.Error("new state manager should have no changes")
	}
}

func TestSlotStateManager_SetGrid(t *testing.T) {
	cfg := stateTestConfig()
	sm := NewSlotStateManager(cfg)

	grid := gridFromString("AAAA----", cfg)
	sm.SetGrid(grid)

	if sm.Grid() != grid {
		t.Error("SetGrid should set the grid")
	}
	if sm.SavedGrid() != grid {
		t.Error("SetGrid should set savedGrid")
	}
}

func TestSlotStateManager_EnterEditMode(t *testing.T) {
	cfg := stateTestConfig()
	sm := NewSlotStateManager(cfg)

	grid := gridFromString("AAAA----", cfg)
	sm.SetGrid(grid)

	sm.EnterEditMode()

	if !sm.IsEditing() {
		t.Error("should be in edit mode")
	}
	if sm.Grid() == sm.SavedGrid() {
		t.Error("working grid should be a clone, not same reference")
	}

	// Verify working grid has same content
	taskA := sm.Grid().TaskAt(0, 0)
	if taskA == nil {
		t.Error("working grid should have task A")
	}
}

func TestSlotStateManager_DiscardChanges(t *testing.T) {
	cfg := stateTestConfig()
	sm := NewSlotStateManager(cfg)

	grid := gridFromString("AAAA----", cfg)
	sm.SetGrid(grid)
	sm.EnterEditMode()

	taskA := getTaskByID(sm.Grid(), 0)

	// Make a change
	err := sm.Grow(taskA)
	if err != nil {
		t.Fatalf("Grow failed: %v", err)
	}

	if !sm.HasChanges() {
		t.Error("should have changes after Grow")
	}

	sm.DiscardChanges()

	if sm.IsEditing() {
		t.Error("should not be in edit mode after discard")
	}
	if sm.HasChanges() {
		t.Error("should not have changes after discard")
	}
}

func TestSlotStateManager_Undo(t *testing.T) {
	cfg := stateTestConfig()
	sm := NewSlotStateManager(cfg)

	grid := gridFromString("AAAA----", cfg)
	sm.SetGrid(grid)
	sm.EnterEditMode()

	taskA := getTaskByID(sm.Grid(), 0)

	// Record original state
	originalEnd := 4 // Task A occupies slots 0-3

	// Grow the task
	err := sm.Grow(taskA)
	if err != nil {
		t.Fatalf("Grow failed: %v", err)
	}

	// Verify task grew
	_, _, endSlot, _ := sm.Grid().FindTask(taskA)
	if endSlot != originalEnd+1 {
		t.Errorf("task should have grown: got end %d, want %d", endSlot, originalEnd+1)
	}

	if !sm.CanUndo() {
		t.Error("should be able to undo")
	}
	if sm.UndoCount() != 1 {
		t.Errorf("undo count should be 1, got %d", sm.UndoCount())
	}

	// Undo
	err = sm.Undo()
	if err != nil {
		t.Fatalf("Undo failed: %v", err)
	}

	// Verify task is back to original size
	_, _, endSlot, _ = sm.Grid().FindTask(taskA)
	if endSlot != originalEnd {
		t.Errorf("task should be back to original: got end %d, want %d", endSlot, originalEnd)
	}

	if sm.CanUndo() {
		t.Error("should not be able to undo after undoing")
	}
}

func TestSlotStateManager_UndoErrors(t *testing.T) {
	cfg := stateTestConfig()
	sm := NewSlotStateManager(cfg)

	grid := gridFromString("AAAA----", cfg)
	sm.SetGrid(grid)

	// Undo without edit mode
	err := sm.Undo()
	if err != ErrSlotNotInEditMode {
		t.Errorf("expected ErrSlotNotInEditMode, got %v", err)
	}

	sm.EnterEditMode()

	// Undo with no history
	err = sm.Undo()
	if err != ErrSlotNothingToUndo {
		t.Errorf("expected ErrSlotNothingToUndo, got %v", err)
	}
}

func TestSlotStateManager_Grow(t *testing.T) {
	cfg := stateTestConfig()
	sm := NewSlotStateManager(cfg)

	grid := gridFromString("AAAA----", cfg)
	sm.SetGrid(grid)
	sm.EnterEditMode()

	taskA := getTaskByID(sm.Grid(), 0)

	err := sm.Grow(taskA)
	if err != nil {
		t.Fatalf("Grow failed: %v", err)
	}

	// Verify task grew
	day, startSlot, endSlot, found := sm.Grid().FindTask(taskA)
	if !found {
		t.Fatal("task not found after grow")
	}
	if day != 0 || startSlot != 0 || endSlot != 5 {
		t.Errorf("wrong position after grow: day=%d, start=%d, end=%d", day, startSlot, endSlot)
	}

	// Verify dirty tracking
	if !sm.HasChanges() {
		t.Error("should have changes")
	}
	dirtyDays := sm.DirtyDays()
	if !dirtyDays[0] {
		t.Error("day 0 should be dirty")
	}
}

func TestSlotStateManager_Shrink(t *testing.T) {
	cfg := stateTestConfig()
	sm := NewSlotStateManager(cfg)

	grid := gridFromString("AAAA----", cfg)
	sm.SetGrid(grid)
	sm.EnterEditMode()

	taskA := getTaskByID(sm.Grid(), 0)

	err := sm.Shrink(taskA)
	if err != nil {
		t.Fatalf("Shrink failed: %v", err)
	}

	// Verify task shrunk
	_, startSlot, endSlot, found := sm.Grid().FindTask(taskA)
	if !found {
		t.Fatal("task not found after shrink")
	}
	if startSlot != 0 || endSlot != 3 {
		t.Errorf("wrong position after shrink: start=%d, end=%d", startSlot, endSlot)
	}
}

func TestSlotStateManager_ShrinkMinimum(t *testing.T) {
	cfg := stateTestConfig()
	sm := NewSlotStateManager(cfg)

	grid := gridFromString("A-------", cfg) // 1 slot task
	sm.SetGrid(grid)
	sm.EnterEditMode()

	taskA := getTaskByID(sm.Grid(), 0)

	err := sm.Shrink(taskA)
	if err != ErrMinimumSlotsDuration {
		t.Errorf("expected ErrMinimumSlotsDuration, got %v", err)
	}
}

func TestSlotStateManager_AddSpace(t *testing.T) {
	cfg := stateTestConfig()
	sm := NewSlotStateManager(cfg)

	// Use "AAAABBBB" so there's something to shift
	grid := gridFromString("AAAABBBB", cfg)
	sm.SetGrid(grid)
	sm.EnterEditMode()

	taskA := getTaskByID(sm.Grid(), 0)
	taskB := getTaskByID(sm.Grid(), 1)

	err := sm.AddSpace(taskA)
	if err != nil {
		t.Fatalf("AddSpace failed: %v", err)
	}

	// Task A should still be at 0-3 (4 slots)
	_, startSlotA, endSlotA, foundA := sm.Grid().FindTask(taskA)
	if !foundA {
		t.Fatal("task A not found after add space")
	}
	if startSlotA != 0 || endSlotA != 4 {
		t.Errorf("task A wrong position: start=%d, end=%d, want start=0, end=4", startSlotA, endSlotA)
	}

	// Task B should have shifted right by 1 (start=5, end=9)
	_, startSlotB, endSlotB, foundB := sm.Grid().FindTask(taskB)
	if !foundB {
		t.Fatal("task B not found after add space")
	}
	if startSlotB != 5 || endSlotB != 9 {
		t.Errorf("task B wrong position after add space: start=%d, end=%d, want start=5, end=9", startSlotB, endSlotB)
	}

	// Verify slot 4 is now empty (space between A and B)
	if sm.Grid().TaskAt(0, 4) != nil {
		t.Error("slot 4 should be empty after add space")
	}
}

func TestSlotStateManager_RemoveSpaceAt(t *testing.T) {
	cfg := stateTestConfig()
	sm := NewSlotStateManager(cfg)

	grid := gridFromString("AAAA-BBBB", cfg)
	sm.SetGrid(grid)
	sm.EnterEditMode()

	taskB := getTaskByID(sm.Grid(), 1)

	err := sm.RemoveSpaceAt(0, 4)
	if err != nil {
		t.Fatalf("RemoveSpaceAt failed: %v", err)
	}

	_, startSlotB, endSlotB, foundB := sm.Grid().FindTask(taskB)
	if !foundB {
		t.Fatal("task B not found after remove space")
	}
	if startSlotB != 4 || endSlotB != 8 {
		t.Errorf("task B wrong position after remove space: start=%d, end=%d, want start=4, end=8", startSlotB, endSlotB)
	}
}

func TestSlotStateManager_Delete(t *testing.T) {
	cfg := stateTestConfig()
	sm := NewSlotStateManager(cfg)

	grid := gridFromString("AAAABBBB", cfg)
	sm.SetGrid(grid)
	sm.EnterEditMode()

	taskA := getTaskByID(sm.Grid(), 0)
	taskB := getTaskByID(sm.Grid(), 1)

	err := sm.Delete(taskA)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify task A is gone
	_, _, _, found := sm.Grid().FindTask(taskA)
	if found {
		t.Error("task A should be deleted")
	}

	// Verify task B shifted left
	_, startSlot, endSlot, found := sm.Grid().FindTask(taskB)
	if !found {
		t.Fatal("task B not found after delete")
	}
	if startSlot != 0 || endSlot != 4 {
		t.Errorf("wrong position for B after delete: start=%d, end=%d", startSlot, endSlot)
	}
}

func TestSlotStateManager_OperationsRequireEditMode(t *testing.T) {
	cfg := stateTestConfig()
	sm := NewSlotStateManager(cfg)

	grid := gridFromString("AAAA----", cfg)
	sm.SetGrid(grid)
	// NOT entering edit mode

	taskA := getTaskByID(grid, 0)

	tests := []struct {
		name string
		fn   func() error
	}{
		{"Grow", func() error { return sm.Grow(taskA) }},
		{"Shrink", func() error { return sm.Shrink(taskA) }},
		{"AddSpace", func() error { return sm.AddSpace(taskA) }},
		{"Delete", func() error { return sm.Delete(taskA) }},
		{"StartMove", func() error { return sm.StartMove(taskA) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if err != ErrSlotNotInEditMode {
				t.Errorf("expected ErrSlotNotInEditMode, got %v", err)
			}
		})
	}
}

func TestSlotStateManager_StartMove(t *testing.T) {
	cfg := stateTestConfig()
	sm := NewSlotStateManager(cfg)

	grid := gridFromString("AAAA----", cfg)
	sm.SetGrid(grid)
	sm.EnterEditMode()

	taskA := getTaskByID(sm.Grid(), 0)

	err := sm.StartMove(taskA)
	if err != nil {
		t.Fatalf("StartMove failed: %v", err)
	}

	if !sm.IsMoving() {
		t.Error("should be in move mode")
	}

	moveState := sm.MoveState()
	if moveState == nil {
		t.Fatal("move state should not be nil")
	}
	if moveState.MovingTask != taskA {
		t.Error("wrong moving task")
	}
	if moveState.SourceDay != 0 || moveState.SourceSlot != 0 {
		t.Errorf("wrong source position: day=%d, slot=%d", moveState.SourceDay, moveState.SourceSlot)
	}
}

func TestSlotStateManager_StartMoveErrors(t *testing.T) {
	cfg := stateTestConfig()
	sm := NewSlotStateManager(cfg)

	grid := gridFromString("AAAA----", cfg)
	sm.SetGrid(grid)
	sm.EnterEditMode()

	taskA := getTaskByID(sm.Grid(), 0)

	// Already moving
	err := sm.StartMove(taskA)
	if err != nil {
		t.Fatalf("StartMove failed: %v", err)
	}

	err = sm.StartMove(taskA)
	if err != ErrSlotAlreadyMoving {
		t.Errorf("expected ErrSlotAlreadyMoving, got %v", err)
	}
}

func TestSlotStateManager_MoveDown(t *testing.T) {
	cfg := stateTestConfig()
	sm := NewSlotStateManager(cfg)

	grid := gridFromString("AAAA----", cfg)
	sm.SetGrid(grid)
	sm.EnterEditMode()

	taskA := getTaskByID(sm.Grid(), 0)

	err := sm.StartMove(taskA)
	if err != nil {
		t.Fatalf("StartMove failed: %v", err)
	}

	// Move down (into the gap)
	err = sm.MoveDown()
	if err != nil {
		t.Fatalf("MoveDown failed: %v", err)
	}

	// Verify task moved
	_, startSlot, _, found := sm.Grid().FindTask(taskA)
	if !found {
		t.Fatal("task not found after move")
	}
	// Task should have moved to end of gap (slot 4)
	if startSlot != 4 {
		t.Errorf("task should be at slot 4, got %d", startSlot)
	}

	// Verify move state updated
	moveState := sm.MoveState()
	if moveState.TargetSlot != 4 {
		t.Errorf("wrong target slot in move state: %d", moveState.TargetSlot)
	}
}

func TestSlotStateManager_MoveUp(t *testing.T) {
	cfg := stateTestConfig()
	sm := NewSlotStateManager(cfg)

	grid := gridFromString("----AAAA", cfg)
	sm.SetGrid(grid)
	sm.EnterEditMode()

	taskA := getTaskByID(sm.Grid(), 0)

	err := sm.StartMove(taskA)
	if err != nil {
		t.Fatalf("StartMove failed: %v", err)
	}

	// Move up (into the gap)
	err = sm.MoveUp()
	if err != nil {
		t.Fatalf("MoveUp failed: %v", err)
	}

	// Verify task moved
	_, startSlot, _, found := sm.Grid().FindTask(taskA)
	if !found {
		t.Fatal("task not found after move")
	}
	// Task should have moved to start of gap (slot 0)
	if startSlot != 0 {
		t.Errorf("task should be at slot 0, got %d", startSlot)
	}
}

func TestSlotStateManager_MoveRight(t *testing.T) {
	cfg := stateTestConfig()
	sm := NewSlotStateManager(cfg)

	grid := gridFromString("AAAA----|--------", cfg)
	sm.SetGrid(grid)
	sm.EnterEditMode()

	taskA := getTaskByID(sm.Grid(), 0)

	err := sm.StartMove(taskA)
	if err != nil {
		t.Fatalf("StartMove failed: %v", err)
	}

	// Move to next day
	err = sm.MoveRight()
	if err != nil {
		t.Fatalf("MoveRight failed: %v", err)
	}

	// Verify task is on day 1
	day, startSlot, _, found := sm.Grid().FindTask(taskA)
	if !found {
		t.Fatal("task not found after cross-day move")
	}
	if day != 1 || startSlot != 0 {
		t.Errorf("wrong position: day=%d, slot=%d", day, startSlot)
	}

	// Verify move state updated
	moveState := sm.MoveState()
	if moveState.TargetDay != 1 {
		t.Errorf("wrong target day in move state: %d", moveState.TargetDay)
	}
}

func TestSlotStateManager_MoveErrors(t *testing.T) {
	cfg := stateTestConfig()
	sm := NewSlotStateManager(cfg)

	grid := gridFromString("AAAA----", cfg)
	sm.SetGrid(grid)
	sm.EnterEditMode()

	// MoveDown without StartMove
	err := sm.MoveDown()
	if err != ErrSlotNotMoving {
		t.Errorf("expected ErrSlotNotMoving, got %v", err)
	}

	// MoveUp without StartMove
	err = sm.MoveUp()
	if err != ErrSlotNotMoving {
		t.Errorf("expected ErrSlotNotMoving, got %v", err)
	}

	// MoveRight without StartMove
	err = sm.MoveRight()
	if err != ErrSlotNotMoving {
		t.Errorf("expected ErrSlotNotMoving, got %v", err)
	}
}

func TestSlotStateManager_ConfirmMove(t *testing.T) {
	cfg := stateTestConfig()
	sm := NewSlotStateManager(cfg)

	grid := gridFromString("AAAA----", cfg)
	sm.SetGrid(grid)
	sm.EnterEditMode()

	taskA := getTaskByID(sm.Grid(), 0)

	err := sm.StartMove(taskA)
	if err != nil {
		t.Fatalf("StartMove failed: %v", err)
	}

	err = sm.MoveDown()
	if err != nil {
		t.Fatalf("MoveDown failed: %v", err)
	}

	err = sm.ConfirmMove()
	if err != nil {
		t.Fatalf("ConfirmMove failed: %v", err)
	}

	if sm.IsMoving() {
		t.Error("should not be in move mode after confirm")
	}

	// Verify task stayed at new position
	_, startSlot, _, found := sm.Grid().FindTask(taskA)
	if !found {
		t.Fatal("task not found after confirm")
	}
	if startSlot != 4 {
		t.Errorf("task should be at slot 4, got %d", startSlot)
	}

	// Verify we can undo
	if !sm.CanUndo() {
		t.Error("should be able to undo after confirm")
	}

	// Verify dirty tracking
	if !sm.HasChanges() {
		t.Error("should have changes after confirm")
	}
}

func TestSlotStateManager_CancelMove(t *testing.T) {
	cfg := stateTestConfig()
	sm := NewSlotStateManager(cfg)

	grid := gridFromString("AAAA----", cfg)
	sm.SetGrid(grid)
	sm.EnterEditMode()

	taskA := getTaskByID(sm.Grid(), 0)

	err := sm.StartMove(taskA)
	if err != nil {
		t.Fatalf("StartMove failed: %v", err)
	}

	err = sm.MoveDown()
	if err != nil {
		t.Fatalf("MoveDown failed: %v", err)
	}

	sm.CancelMove()

	if sm.IsMoving() {
		t.Error("should not be in move mode after cancel")
	}

	// Verify task is back at original position
	_, startSlot, _, found := sm.Grid().FindTask(taskA)
	if !found {
		t.Fatal("task not found after cancel")
	}
	if startSlot != 0 {
		t.Errorf("task should be at slot 0, got %d", startSlot)
	}
}

func TestSlotStateManager_MoveWithShiftedTasks(t *testing.T) {
	cfg := stateTestConfig()
	sm := NewSlotStateManager(cfg)

	grid := gridFromString("AAAABBBB", cfg)
	sm.SetGrid(grid)
	sm.EnterEditMode()

	taskA := getTaskByID(sm.Grid(), 0)
	taskB := getTaskByID(sm.Grid(), 1)

	err := sm.StartMove(taskA)
	if err != nil {
		t.Fatalf("StartMove failed: %v", err)
	}

	// Move A down (should swap with B)
	err = sm.MoveDown()
	if err != nil {
		t.Fatalf("MoveDown failed: %v", err)
	}

	moveState := sm.MoveState()
	if len(moveState.ShiftedTasks) != 1 {
		t.Errorf("expected 1 shifted task, got %d", len(moveState.ShiftedTasks))
	}

	// Verify B is marked as shifted
	if !sm.IsTaskShifted(taskB.ID) {
		t.Error("task B should be marked as shifted")
	}
}

func TestSlotStateManager_CommitChanges(t *testing.T) {
	cfg := stateTestConfig()
	sm := NewSlotStateManager(cfg)

	grid := gridFromString("AAAA----", cfg)
	sm.SetGrid(grid)
	sm.EnterEditMode()

	taskA := getTaskByID(sm.Grid(), 0)

	err := sm.Grow(taskA)
	if err != nil {
		t.Fatalf("Grow failed: %v", err)
	}

	sm.CommitChanges()

	if sm.IsEditing() {
		t.Error("should not be in edit mode after commit")
	}
	if sm.HasChanges() {
		t.Error("should not have changes after commit")
	}

	// Verify saved grid has the changes
	_, _, endSlot, found := sm.SavedGrid().FindTask(taskA)
	if !found {
		t.Fatal("task not found in saved grid")
	}
	if endSlot != 5 {
		t.Errorf("saved grid should have grown task: end=%d", endSlot)
	}
}

func TestSlotStateManager_MultipleOperationsAndUndo(t *testing.T) {
	cfg := stateTestConfig()
	sm := NewSlotStateManager(cfg)

	grid := gridFromString("AAAA----", cfg)
	sm.SetGrid(grid)
	sm.EnterEditMode()

	taskA := getTaskByID(sm.Grid(), 0)

	// Grow twice
	err := sm.Grow(taskA)
	if err != nil {
		t.Fatalf("Grow 1 failed: %v", err)
	}
	err = sm.Grow(taskA)
	if err != nil {
		t.Fatalf("Grow 2 failed: %v", err)
	}

	if sm.UndoCount() != 2 {
		t.Errorf("expected 2 undo entries, got %d", sm.UndoCount())
	}

	// Undo once
	err = sm.Undo()
	if err != nil {
		t.Fatalf("Undo 1 failed: %v", err)
	}

	_, _, endSlot, _ := sm.Grid().FindTask(taskA)
	if endSlot != 5 {
		t.Errorf("after first undo, end should be 5, got %d", endSlot)
	}

	// Undo again
	err = sm.Undo()
	if err != nil {
		t.Fatalf("Undo 2 failed: %v", err)
	}

	_, _, endSlot, _ = sm.Grid().FindTask(taskA)
	if endSlot != 4 {
		t.Errorf("after second undo, end should be 4, got %d", endSlot)
	}
}

func TestSlotStateManager_PastValidation(t *testing.T) {
	// Create config where first slots are in the past
	today := time.Now()
	midnight := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	now := midnight.Add(10 * time.Hour) // 10:00 today = slot 40

	cfg := SlotConfig{
		FirstDate:         midnight,
		NumDays:           7,
		SlotDuration:      15,
		WorkingHoursStart: 9 * 60,  // 09:00
		WorkingHoursEnd:   17 * 60, // 17:00
		Now:               func() time.Time { return now },
	}

	// Create a grid with tasks
	// At 10:00, slot 40 is current (10:00-10:15)
	// Slots 0-40 are past or current, slots 41+ are future
	grid := NewSlotGrid(cfg)
	taskA := &task.Task{ID: 0, Description: "Past Task", Category: task.CategoryDeep, Status: task.StatusScheduled}
	taskB := &task.Task{ID: 1, Description: "Future Task", Category: task.CategoryDeep, Status: task.StatusScheduled}

	// Place task A at slots 0-1 (00:00-00:30) - in the past
	grid, _ = grid.Place(taskA, 0, 0, 2)
	// Place task B at slots 50-53 (12:30-13:30) - in the future
	grid, _ = grid.Place(taskB, 0, 50, 4)

	sm := NewSlotStateManager(cfg)
	sm.SetGrid(grid)
	sm.EnterEditMode()

	// Cannot start move for past task
	err := sm.StartMove(taskA)
	if err != ErrTaskAlreadyStarted {
		t.Errorf("expected ErrTaskAlreadyStarted, got %v", err)
	}

	// Can start move for future task
	err = sm.StartMove(taskB)
	if err != nil {
		t.Fatalf("StartMove for future task failed: %v", err)
	}
}

func TestSlotStateManager_AccumulatedMoves(t *testing.T) {
	cfg := stateTestConfig()
	sm := NewSlotStateManager(cfg)

	// Create grid with A followed by gap, then B
	// A is 4 slots, gap is 8 slots, B is 4 slots
	grid := gridFromString("AAAA--------BBBB", cfg)
	sm.SetGrid(grid)
	sm.EnterEditMode()

	taskA := getTaskByID(sm.Grid(), 0)

	err := sm.StartMove(taskA)
	if err != nil {
		t.Fatalf("StartMove failed: %v", err)
	}

	// Move down multiple times - with one-step-at-a-time behavior in gaps,
	// each MoveDown moves A by its own size (4 slots) when in a gap
	err = sm.MoveDown() // Move into gap: A moves from 0 to 4
	if err != nil {
		t.Fatalf("MoveDown 1 failed: %v", err)
	}

	err = sm.MoveDown() // Continue in gap: A moves from 4 to 8
	if err != nil {
		t.Fatalf("MoveDown 2 failed: %v", err)
	}

	// After two moves in gap, A should be at slot 8 (still in gap, one step before B)
	_, startSlot, _, found := sm.Grid().FindTask(taskA)
	if !found {
		t.Fatal("task A not found")
	}

	// Task A should be at slot 8 after two moves (4 slots each move)
	if startSlot != 8 {
		t.Errorf("task A should be at slot 8 after 2 moves, got %d", startSlot)
	}

	// One more move should swap A with B
	err = sm.MoveDown()
	if err != nil {
		t.Fatalf("MoveDown 3 failed: %v", err)
	}

	_, startSlot, _, found = sm.Grid().FindTask(taskA)
	if !found {
		t.Fatal("task A not found after swap")
	}

	// Now A should have swapped with B (B was at 12-15, A takes that position)
	if startSlot != 12 {
		t.Errorf("task A should be at slot 12 after swap with B, got %d", startSlot)
	}

	// Cancel should restore original position
	sm.CancelMove()

	_, startSlot, _, _ = sm.Grid().FindTask(taskA)
	if startSlot != 0 {
		t.Errorf("after cancel, task A should be at slot 0, got %d", startSlot)
	}
}

func TestSlotStateManager_UpdateConfig(t *testing.T) {
	cfg := stateTestConfig()
	sm := NewSlotStateManager(cfg)

	// Verify initial config
	initialFirstDate := sm.Config().FirstDate
	if !initialFirstDate.Equal(cfg.FirstDate) {
		t.Errorf("initial config FirstDate mismatch: got %v, want %v", sm.Config().FirstDate, cfg.FirstDate)
	}

	// Create new config with different FirstDate (shifted forward by 1 week)
	newFirstDate := cfg.FirstDate.AddDate(0, 0, 7)
	newCfg := SlotConfig{
		FirstDate:         newFirstDate,
		NumDays:           cfg.NumDays,
		SlotDuration:      cfg.SlotDuration,
		WorkingHoursStart: cfg.WorkingHoursStart,
		WorkingHoursEnd:   cfg.WorkingHoursEnd,
		Now:               cfg.Now,
	}

	// Update config
	sm.UpdateConfig(newCfg)

	// Verify config was updated
	if !sm.Config().FirstDate.Equal(newFirstDate) {
		t.Errorf("config FirstDate not updated: got %v, want %v", sm.Config().FirstDate, newFirstDate)
	}
	if sm.Config().NumDays != newCfg.NumDays {
		t.Errorf("config NumDays mismatch: got %d, want %d", sm.Config().NumDays, newCfg.NumDays)
	}
}

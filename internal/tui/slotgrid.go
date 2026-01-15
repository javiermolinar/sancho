package tui

import (
	"errors"
	"strings"
	"time"

	"github.com/javiermolinar/sancho/internal/task"
)

// SlotGrid errors.
var (
	ErrTaskAlreadyStarted   = errors.New("task has already started and cannot be moved")
	ErrSlotOccupied         = errors.New("target slot is already occupied")
	ErrInvalidSlotPosition  = errors.New("invalid slot position")
	ErrSlotTaskNotFound     = errors.New("task not found in grid")
	ErrMinimumSlotsDuration = errors.New("cannot shrink below 1 slot (15 minutes)")
	ErrNoGapToRemove        = errors.New("no gap to remove")
)

const (
	// DefaultSlotDuration is the internal slot duration in minutes.
	DefaultSlotDuration = 15
	// DefaultNumWeeks is the number of weeks in the sliding window.
	DefaultNumWeeks = 3
	// DaysPerWeek is the number of days per week.
	DaysPerWeek = 7
	// SlotsPerDay is 24 hours * 4 slots per hour = 96 slots.
	SlotsPerDay = 96
	// MinutesPerDay is 24 hours * 60 minutes.
	MinutesPerDay = 1440
)

// SlotConfig holds grid configuration.
type SlotConfig struct {
	SlotDuration int              // Always 15 minutes internally
	NumDays      int              // 21 (3 weeks * 7 days)
	FirstDate    time.Time        // Date of day index 0
	Now          func() time.Time // Injectable for testing

	// For UI display only (not used in grid logic):
	WorkingHoursStart int // e.g., 480 (8:00) - minutes from midnight
	WorkingHoursEnd   int // e.g., 1080 (18:00) - minutes from midnight

	// DisplaySlotSize is the number of 15-min slots that make up one screen block.
	// Used for visual movement: when moving into a gap, move by this many slots.
	// Examples: 1 (15-min blocks), 2 (30-min blocks), 4 (60-min blocks).
	// Defaults to 4 (60-min blocks) if not set.
	DisplaySlotSize int
}

// SlotsPerDay returns the number of slots per day (always 96 for 24h grid).
func (c SlotConfig) SlotsPerDay() int {
	return SlotsPerDay
}

// TotalSlots returns the total number of slots in the grid.
func (c SlotConfig) TotalSlots() int {
	return SlotsPerDay * c.NumDays
}

// SlotToTime converts a slot index to a time string "HH:MM".
func (c SlotConfig) SlotToTime(slot int) string {
	mins := slot * c.SlotDuration
	return task.MinutesToTime(mins)
}

// TimeToSlot converts a time string "HH:MM" to a slot index.
func (c SlotConfig) TimeToSlot(timeStr string) int {
	mins := task.TimeToMinutes(timeStr)
	return mins / c.SlotDuration
}

// SlotToMinutes converts a slot index to minutes from midnight.
func (c SlotConfig) SlotToMinutes(slot int) int {
	return slot * c.SlotDuration
}

// MinutesToSlot converts minutes from midnight to a slot index.
func (c SlotConfig) MinutesToSlot(mins int) int {
	return mins / c.SlotDuration
}

// IsWorkingHours returns true if the slot is within configured working hours.
func (c SlotConfig) IsWorkingHours(slot int) bool {
	mins := slot * c.SlotDuration
	return mins >= c.WorkingHoursStart && mins < c.WorkingHoursEnd
}

// WorkingHoursStartSlot returns the first slot of working hours.
func (c SlotConfig) WorkingHoursStartSlot() int {
	return c.WorkingHoursStart / c.SlotDuration
}

// WorkingHoursEndSlot returns the slot after working hours end.
func (c SlotConfig) WorkingHoursEndSlot() int {
	return c.WorkingHoursEnd / c.SlotDuration
}

// WorkingSlotsPerDay returns the number of slots in working hours.
func (c SlotConfig) WorkingSlotsPerDay() int {
	return (c.WorkingHoursEnd - c.WorkingHoursStart) / c.SlotDuration
}

// GetDisplaySlotSize returns the display slot size, defaulting to 4 (60 min) if not set.
func (c SlotConfig) GetDisplaySlotSize() int {
	if c.DisplaySlotSize <= 0 {
		return 4 // Default: 60 minutes = 4 x 15-min slots
	}
	return c.DisplaySlotSize
}

// DayIndexToDate converts a day index to a date.
func (c SlotConfig) DayIndexToDate(dayIndex int) time.Time {
	return c.FirstDate.AddDate(0, 0, dayIndex)
}

// DateToDayIndex converts a date to a day index.
// Returns -1 if the date is before FirstDate or after the grid ends.
func (c SlotConfig) DateToDayIndex(date time.Time) int {
	// Truncate both to start of day for comparison
	first := time.Date(c.FirstDate.Year(), c.FirstDate.Month(), c.FirstDate.Day(), 0, 0, 0, 0, c.FirstDate.Location())
	d := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	days := int(d.Sub(first).Hours() / 24)
	if days < 0 || days >= c.NumDays {
		return -1
	}
	return days
}

// NewSlotConfig creates a SlotConfig with defaults.
func NewSlotConfig(workStart, workEnd string, firstDate time.Time) SlotConfig {
	return SlotConfig{
		SlotDuration:      DefaultSlotDuration,
		NumDays:           DefaultNumWeeks * DaysPerWeek,
		FirstDate:         firstDate,
		Now:               time.Now,
		WorkingHoursStart: task.TimeToMinutes(workStart),
		WorkingHoursEnd:   task.TimeToMinutes(workEnd),
	}
}

// SlotGrid is an immutable data structure representing task positions.
// Each slot is 15 minutes. A task spanning multiple slots has its pointer
// in each slot it occupies. The grid uses a 24-hour day (96 slots).
type SlotGrid struct {
	slots  []*task.Task // Length = SlotsPerDay * NumDays
	config SlotConfig
}

// NewSlotGrid creates an empty SlotGrid.
func NewSlotGrid(config SlotConfig) *SlotGrid {
	return &SlotGrid{
		slots:  make([]*task.Task, config.TotalSlots()),
		config: config,
	}
}

// Config returns the grid configuration.
func (g *SlotGrid) Config() SlotConfig {
	return g.config
}

// slotIndex calculates the flat array index for a day and slot.
func (g *SlotGrid) slotIndex(day, slot int) int {
	return day*SlotsPerDay + slot
}

// isValidPosition checks if day and slot are within bounds.
func (g *SlotGrid) isValidPosition(day, slot int) bool {
	if day < 0 || day >= g.config.NumDays {
		return false
	}
	if slot < 0 || slot >= SlotsPerDay {
		return false
	}
	return true
}

// TaskAt returns the task at the given position, or nil if empty.
func (g *SlotGrid) TaskAt(day, slot int) *task.Task {
	if !g.isValidPosition(day, slot) {
		return nil
	}
	return g.slots[g.slotIndex(day, slot)]
}

// IsEmpty returns true if the given position is empty.
func (g *SlotGrid) IsEmpty(day, slot int) bool {
	return g.TaskAt(day, slot) == nil
}

// FindTask returns the position and size of a task in the grid.
// Returns day, startSlot, endSlot (exclusive), and found.
func (g *SlotGrid) FindTask(t *task.Task) (day, startSlot, endSlot int, found bool) {
	if t == nil {
		return 0, 0, 0, false
	}
	return g.FindTaskByID(t.ID)
}

// FindTaskByID returns the position and size of a task by ID.
// Returns day, startSlot, endSlot (exclusive), and found.
func (g *SlotGrid) FindTaskByID(id int64) (day, startSlot, endSlot int, found bool) {
	for d := 0; d < g.config.NumDays; d++ {
		for s := 0; s < SlotsPerDay; s++ {
			idx := g.slotIndex(d, s)
			if g.slots[idx] != nil && g.slots[idx].ID == id {
				// Found the start of the task, count consecutive slots
				startSlot = s
				day = d
				endSlot = s + 1
				for endSlot < SlotsPerDay && g.slots[g.slotIndex(d, endSlot)] != nil && g.slots[g.slotIndex(d, endSlot)].ID == id {
					endSlot++
				}
				return day, startSlot, endSlot, true
			}
		}
	}
	return 0, 0, 0, false
}

// AllTasks returns all unique tasks in the grid.
func (g *SlotGrid) AllTasks() []*task.Task {
	seen := make(map[int64]bool)
	var result []*task.Task

	for _, t := range g.slots {
		if t != nil && !seen[t.ID] {
			seen[t.ID] = true
			result = append(result, t)
		}
	}
	return result
}

// TasksOnDay returns all unique tasks on a specific day.
func (g *SlotGrid) TasksOnDay(day int) []*task.Task {
	if day < 0 || day >= g.config.NumDays {
		return nil
	}

	seen := make(map[int64]bool)
	var result []*task.Task

	for s := 0; s < SlotsPerDay; s++ {
		t := g.slots[g.slotIndex(day, s)]
		if t != nil && !seen[t.ID] {
			seen[t.ID] = true
			result = append(result, t)
		}
	}
	return result
}

// clone creates a deep copy of the grid.
func (g *SlotGrid) clone() *SlotGrid {
	newSlots := make([]*task.Task, len(g.slots))
	copy(newSlots, g.slots)
	return &SlotGrid{
		slots:  newSlots,
		config: g.config,
	}
}

// currentTimePosition returns the current day index and slot based on Now().
// Returns (-1, -1) if Now is before the grid starts (nothing is past).
// Returns (NumDays, 0) if Now is after the grid ends (everything is past).
func (g *SlotGrid) currentTimePosition() (day, slot int) {
	now := g.config.Now()

	// Find day index
	day = g.config.DateToDayIndex(now)
	if day < 0 {
		// Before grid start - nothing in the grid is past
		return -1, -1
	}
	if day >= g.config.NumDays {
		// After grid end - everything in the grid is past
		return g.config.NumDays, 0
	}

	// Find slot from current time (24-hour grid, so always valid)
	mins := now.Hour()*60 + now.Minute()
	slot = mins / g.config.SlotDuration

	return day, slot
}

// isPastPosition returns true if the given position is in the past.
func (g *SlotGrid) isPastPosition(day, slot int) bool {
	nowDay, nowSlot := g.currentTimePosition()

	// If now is before the grid, nothing is past
	if nowDay < 0 {
		return false
	}

	// Compare positions
	nowPos := nowDay*SlotsPerDay + nowSlot
	targetPos := day*SlotsPerDay + slot
	return targetPos <= nowPos
}

// canModifyTask checks if a task can be modified (not started yet).
func (g *SlotGrid) canModifyTask(t *task.Task) error {
	day, startSlot, _, found := g.FindTask(t)
	if !found {
		return ErrSlotTaskNotFound
	}

	if g.isPastPosition(day, startSlot) {
		return ErrTaskAlreadyStarted
	}

	return nil
}

// Place adds a task to the grid at the specified position.
// This is used during initial load and does not check for past positions.
// Returns a new grid with the task placed.
func (g *SlotGrid) Place(t *task.Task, day, startSlot, numSlots int) (*SlotGrid, error) {
	if t == nil {
		return g, nil
	}

	if !g.isValidPosition(day, startSlot) {
		return nil, ErrInvalidSlotPosition
	}

	endSlot := startSlot + numSlots
	if endSlot > SlotsPerDay {
		// Truncate to day boundary
		endSlot = SlotsPerDay
	}

	// Check if slots are available
	for s := startSlot; s < endSlot; s++ {
		existing := g.TaskAt(day, s)
		if existing != nil && existing.ID != t.ID {
			return nil, ErrSlotOccupied
		}
	}

	// Clone and place
	newGrid := g.clone()
	for s := startSlot; s < endSlot; s++ {
		newGrid.slots[newGrid.slotIndex(day, s)] = t
	}

	return newGrid, nil
}

// ============================================================================
// Direction-based Move Operations
// ============================================================================

// MoveUp moves a task to earlier time on the same day.
// Returns the same grid if already at slot 0 or no-op.
// The task swaps with the previous task or moves one step into a gap.
//
// Behavior:
// - If adjacent to another task: swap positions (other task takes our old position)
// - If adjacent to a gap: move by task's size into the gap (one "step")
func (g *SlotGrid) MoveUp(t *task.Task) (*SlotGrid, error) {
	if t == nil {
		return nil, ErrSlotTaskNotFound
	}

	if err := g.canModifyTask(t); err != nil {
		return nil, err
	}

	day, startSlot, endSlot, found := g.FindTask(t)
	if !found {
		return nil, ErrSlotTaskNotFound
	}

	numSlots := endSlot - startSlot

	// Already at first slot - no-op
	if startSlot == 0 {
		return g, nil
	}

	// What's immediately before us?
	prevTask := g.TaskAt(day, startSlot-1)

	if prevTask == nil {
		// Moving into a gap - move by displaySlotSize (one visual "step")
		// But don't go past the previous task or start of day
		gapStart := startSlot - 1
		for gapStart > 0 && g.TaskAt(day, gapStart-1) == nil {
			gapStart--
		}

		// Move by display slot size (visual block), not task size
		stepSize := g.config.GetDisplaySlotSize()
		landing := startSlot - stepSize
		if landing < gapStart {
			landing = gapStart
		}
		if landing >= startSlot {
			// Can't move
			return g, nil
		}

		// Clone and perform move
		newGrid := g.clone()

		// Clear current position
		for s := startSlot; s < endSlot; s++ {
			newGrid.slots[newGrid.slotIndex(day, s)] = nil
		}

		// Place at new position
		for s := landing; s < landing+numSlots; s++ {
			newGrid.slots[newGrid.slotIndex(day, s)] = t
		}

		return newGrid, nil
	}

	// Adjacent to another task - swap positions
	// Find the start of the previous task
	prevStart := startSlot - 1
	for prevStart > 0 && g.TaskAt(day, prevStart-1) != nil && g.TaskAt(day, prevStart-1).ID == prevTask.ID {
		prevStart--
	}
	prevSlots := startSlot - prevStart

	// After swap: our task at [prevStart, prevStart+numSlots), prevTask at [prevStart+numSlots, prevStart+numSlots+prevSlots)

	// Clone and perform swap
	newGrid := g.clone()

	// Clear both tasks
	for s := prevStart; s < endSlot; s++ {
		newGrid.slots[newGrid.slotIndex(day, s)] = nil
	}

	// Place our task first (at previous task's old position)
	for s := prevStart; s < prevStart+numSlots; s++ {
		newGrid.slots[newGrid.slotIndex(day, s)] = t
	}

	// Place prevTask after
	newStart := prevStart + numSlots
	for s := newStart; s < newStart+prevSlots; s++ {
		newGrid.slots[newGrid.slotIndex(day, s)] = prevTask
	}

	return newGrid, nil
}

// MoveDown moves a task to later time on the same day.
// Returns the same grid if at day end or no-op.
// The task swaps with the next task or moves one step into a gap.
//
// Behavior:
// - If adjacent to another task: swap positions (other task takes our old position)
// - If adjacent to a gap: move by task's size into the gap (one "step")
func (g *SlotGrid) MoveDown(t *task.Task) (*SlotGrid, error) {
	if t == nil {
		return nil, ErrSlotTaskNotFound
	}

	if err := g.canModifyTask(t); err != nil {
		return nil, err
	}

	day, startSlot, endSlot, found := g.FindTask(t)
	if !found {
		return nil, ErrSlotTaskNotFound
	}

	numSlots := endSlot - startSlot

	// Already at last position - no-op
	if endSlot >= SlotsPerDay {
		return g, nil
	}

	// What's immediately after us?
	nextTask := g.TaskAt(day, endSlot)

	if nextTask == nil {
		// Moving into a gap - move by displaySlotSize (one visual "step")
		// But don't go past the next task or end of day
		gapEnd := endSlot
		for gapEnd < SlotsPerDay && g.TaskAt(day, gapEnd) == nil {
			gapEnd++
		}

		// Move by display slot size (visual block), not task size
		stepSize := g.config.GetDisplaySlotSize()
		landing := startSlot + stepSize
		maxLanding := gapEnd - numSlots
		if landing > maxLanding {
			landing = maxLanding
		}
		if landing <= startSlot {
			// Can't move (gap too small or already at position)
			return g, nil
		}

		// Clone and perform move
		newGrid := g.clone()

		// Clear current position
		for s := startSlot; s < endSlot; s++ {
			newGrid.slots[newGrid.slotIndex(day, s)] = nil
		}

		// Place at new position
		for s := landing; s < landing+numSlots; s++ {
			newGrid.slots[newGrid.slotIndex(day, s)] = t
		}

		return newGrid, nil
	}

	// Adjacent to another task - swap positions
	// Find the end of the next task
	nextEnd := endSlot + 1
	for nextEnd < SlotsPerDay && g.TaskAt(day, nextEnd) != nil && g.TaskAt(day, nextEnd).ID == nextTask.ID {
		nextEnd++
	}
	nextSlots := nextEnd - endSlot

	// After swap: nextTask at [startSlot, startSlot+nextSlots), our task at [startSlot+nextSlots, startSlot+nextSlots+numSlots)
	newStart := startSlot + nextSlots

	// Clone and perform swap
	newGrid := g.clone()

	// Clear both tasks
	for s := startSlot; s < nextEnd; s++ {
		newGrid.slots[newGrid.slotIndex(day, s)] = nil
	}

	// Place nextTask first (at our old position)
	for s := startSlot; s < startSlot+nextSlots; s++ {
		newGrid.slots[newGrid.slotIndex(day, s)] = nextTask
	}

	// Place our task after
	for s := newStart; s < newStart+numSlots; s++ {
		newGrid.slots[newGrid.slotIndex(day, s)] = t
	}

	return newGrid, nil
}

// MoveRight moves a task to the next day at the same slot.
// Returns the same grid if the move would cause overflow on target day.
// Returns the same grid if target position would be in the past.
// If the target slot is occupied, inserts after the existing task.
func (g *SlotGrid) MoveRight(t *task.Task) (*SlotGrid, error) {
	if t == nil {
		return nil, ErrSlotTaskNotFound
	}

	if err := g.canModifyTask(t); err != nil {
		return nil, err
	}

	sourceDay, startSlot, endSlot, found := g.FindTask(t)
	if !found {
		return nil, ErrSlotTaskNotFound
	}

	numSlots := endSlot - startSlot
	targetDay := sourceDay + 1

	// Check if target day is valid
	if targetDay >= g.config.NumDays {
		// No-op: at last day of grid
		return g, nil
	}

	// Check if target position is in the past
	if g.isPastPosition(targetDay, startSlot) {
		// No-op: can't move to a past slot
		return g, nil
	}

	// Find where to insert the task on target day.
	// If startSlot is in the middle of a task, insert after that task ends.
	insertSlot := g.findInsertSlot(targetDay, startSlot)

	// Check if move would cause overflow on target day
	if g.wouldOverflow(targetDay, insertSlot, numSlots) {
		// No-op: would push tasks past midnight
		return g, nil
	}

	// Clone and perform move
	newGrid := g.clone()

	// === SOURCE DAY: Remove task and shift left (preserving gaps) ===
	// Shift all slots after the removed task left by numSlots positions
	for s := startSlot; s < SlotsPerDay-numSlots; s++ {
		newGrid.slots[newGrid.slotIndex(sourceDay, s)] = newGrid.slots[newGrid.slotIndex(sourceDay, s+numSlots)]
	}
	// Clear the slots at the end that were vacated
	for s := SlotsPerDay - numSlots; s < SlotsPerDay; s++ {
		newGrid.slots[newGrid.slotIndex(sourceDay, s)] = nil
	}

	// === TARGET DAY: Shift right and place task ===
	// Shift slots from insertSlot onwards by numSlots to make room
	for s := SlotsPerDay - 1; s >= insertSlot+numSlots; s-- {
		newGrid.slots[newGrid.slotIndex(targetDay, s)] = newGrid.slots[newGrid.slotIndex(targetDay, s-numSlots)]
	}
	// Clear the slots where we'll place the task
	for s := insertSlot; s < insertSlot+numSlots; s++ {
		newGrid.slots[newGrid.slotIndex(targetDay, s)] = nil
	}

	// Place task at insert position
	for s := insertSlot; s < insertSlot+numSlots; s++ {
		newGrid.slots[newGrid.slotIndex(targetDay, s)] = t
	}

	return newGrid, nil
}

// MoveLeft moves a task to the previous day at the same slot.
// Returns the same grid if already at day 0.
// Only allows moving to days that are not in the past.
// If the target slot is occupied, inserts after the existing task.
func (g *SlotGrid) MoveLeft(t *task.Task) (*SlotGrid, error) {
	if t == nil {
		return nil, ErrSlotTaskNotFound
	}

	if err := g.canModifyTask(t); err != nil {
		return nil, err
	}

	sourceDay, startSlot, endSlot, found := g.FindTask(t)
	if !found {
		return nil, ErrSlotTaskNotFound
	}

	numSlots := endSlot - startSlot
	targetDay := sourceDay - 1

	// Check if target day is valid
	if targetDay < 0 {
		// No-op: at first day of grid
		return g, nil
	}

	// Check if target day is in the past
	if g.isPastPosition(targetDay, startSlot) {
		// No-op: can't move to a past slot
		return g, nil
	}

	// Find where to insert the task on target day.
	// If startSlot is in the middle of a task, insert after that task ends.
	insertSlot := g.findInsertSlot(targetDay, startSlot)

	// Check if move would cause overflow on target day
	if g.wouldOverflow(targetDay, insertSlot, numSlots) {
		// No-op: would push tasks past midnight
		return g, nil
	}

	// Clone and perform move
	newGrid := g.clone()

	// === SOURCE DAY: Remove task and shift left (preserving gaps) ===
	// Shift all slots after the removed task left by numSlots positions
	for s := startSlot; s < SlotsPerDay-numSlots; s++ {
		newGrid.slots[newGrid.slotIndex(sourceDay, s)] = newGrid.slots[newGrid.slotIndex(sourceDay, s+numSlots)]
	}
	// Clear the slots at the end that were vacated
	for s := SlotsPerDay - numSlots; s < SlotsPerDay; s++ {
		newGrid.slots[newGrid.slotIndex(sourceDay, s)] = nil
	}

	// === TARGET DAY: Shift right and place task ===
	// Shift slots from insertSlot onwards by numSlots to make room
	for s := SlotsPerDay - 1; s >= insertSlot+numSlots; s-- {
		newGrid.slots[newGrid.slotIndex(targetDay, s)] = newGrid.slots[newGrid.slotIndex(targetDay, s-numSlots)]
	}
	// Clear the slots where we'll place the task
	for s := insertSlot; s < insertSlot+numSlots; s++ {
		newGrid.slots[newGrid.slotIndex(targetDay, s)] = nil
	}

	// Place task at insert position
	for s := insertSlot; s < insertSlot+numSlots; s++ {
		newGrid.slots[newGrid.slotIndex(targetDay, s)] = t
	}

	return newGrid, nil
}

// findInsertSlot finds the slot where a task should be inserted.
// If targetSlot is empty, returns targetSlot.
// If targetSlot is at the START of a task, returns targetSlot (will shift that task).
// If targetSlot is in the MIDDLE of a task, returns the slot after that task ends.
func (g *SlotGrid) findInsertSlot(day, targetSlot int) int {
	t := g.TaskAt(day, targetSlot)
	if t == nil {
		// Empty slot - insert here
		return targetSlot
	}

	// Check if targetSlot is at the start of this task
	if targetSlot == 0 || g.TaskAt(day, targetSlot-1) == nil || g.TaskAt(day, targetSlot-1).ID != t.ID {
		// At the start of the task - can shift and insert here
		return targetSlot
	}

	// In the middle of a task - insert after it ends
	end := targetSlot + 1
	for end < SlotsPerDay && g.TaskAt(day, end) != nil && g.TaskAt(day, end).ID == t.ID {
		end++
	}
	return end
}

// wouldOverflow checks if shifting right on target day would push any task past slot 95.
func (g *SlotGrid) wouldOverflow(day, fromSlot, amount int) bool {
	// Find the last occupied slot on the target day
	lastOccupied := -1
	for s := SlotsPerDay - 1; s >= fromSlot; s-- {
		if g.TaskAt(day, s) != nil {
			lastOccupied = s
			break
		}
	}

	if lastOccupied < 0 {
		// No tasks to shift
		return false
	}

	// Would shifting push past boundary?
	return lastOccupied+amount >= SlotsPerDay
}

// ============================================================================
// Other Operations (Grow, Shrink, AddSpace, Delete)
// ============================================================================

// Grow extends a task by one slot (15 minutes).
// Shifts following tasks right if necessary.
// Returns same grid if task end is at slot 95 (no-op).
func (g *SlotGrid) Grow(t *task.Task) (*SlotGrid, error) {
	if t == nil {
		return nil, ErrSlotTaskNotFound
	}

	if err := g.canModifyTask(t); err != nil {
		return nil, err
	}

	day, _, endSlot, found := g.FindTask(t)
	if !found {
		return nil, ErrSlotTaskNotFound
	}

	// Already at day end - no-op
	if endSlot >= SlotsPerDay {
		return g, nil
	}

	growSlot := endSlot // The slot we're growing into

	newGrid := g.clone()

	// Check if there's a task in the grow slot
	existingTask := newGrid.TaskAt(day, growSlot)
	if existingTask != nil && existingTask.ID != t.ID {
		// Need to shift following tasks right
		// Check if there's room (would overflow?)
		lastOccupied := -1
		for s := SlotsPerDay - 1; s >= growSlot; s-- {
			if newGrid.TaskAt(day, s) != nil {
				lastOccupied = s
				break
			}
		}
		if lastOccupied >= 0 && lastOccupied+1 >= SlotsPerDay {
			// Would overflow - no-op
			return g, nil
		}

		// Shift from right to left
		for s := SlotsPerDay - 1; s > growSlot; s-- {
			newGrid.slots[newGrid.slotIndex(day, s)] = newGrid.slots[newGrid.slotIndex(day, s-1)]
		}
	}

	// Add the new slot to the task
	newGrid.slots[newGrid.slotIndex(day, growSlot)] = t

	return newGrid, nil
}

// Shrink reduces a task by one slot (15 minutes).
// Returns error if task is already at minimum (1 slot).
func (g *SlotGrid) Shrink(t *task.Task) (*SlotGrid, error) {
	if t == nil {
		return nil, ErrSlotTaskNotFound
	}

	if err := g.canModifyTask(t); err != nil {
		return nil, err
	}

	day, startSlot, endSlot, found := g.FindTask(t)
	if !found {
		return nil, ErrSlotTaskNotFound
	}

	numSlots := endSlot - startSlot
	if numSlots <= 1 {
		return nil, ErrMinimumSlotsDuration
	}

	newGrid := g.clone()
	// Remove the last slot of the task
	lastSlot := endSlot - 1
	newGrid.slots[newGrid.slotIndex(day, lastSlot)] = nil

	return newGrid, nil
}

// AddSpace adds one empty slot (15 minutes) after a task by shifting following tasks right.
// Returns same grid if at day end or would overflow (no-op).
func (g *SlotGrid) AddSpace(t *task.Task) (*SlotGrid, error) {
	if t == nil {
		return nil, ErrSlotTaskNotFound
	}

	day, _, endSlot, found := g.FindTask(t)
	if !found {
		return nil, ErrSlotTaskNotFound
	}

	if g.isPastPosition(day, endSlot) {
		return nil, ErrTaskAlreadyStarted
	}

	return g.AddSpaceAt(day, endSlot)
}

// AddSpaceAt inserts one empty slot at the given day/slot by shifting following slots right.
// Returns same grid if at day end or would overflow (no-op).
func (g *SlotGrid) AddSpaceAt(day, insertSlot int) (*SlotGrid, error) {
	if day < 0 || day >= g.config.NumDays || insertSlot < 0 || insertSlot > SlotsPerDay {
		return nil, ErrInvalidSlotPosition
	}

	if g.isPastPosition(day, insertSlot) {
		return nil, ErrTaskAlreadyStarted
	}

	if insertSlot >= SlotsPerDay {
		// Already at day end - no-op
		return g, nil
	}

	// Check if there's anything to shift
	hasTasksAfter := false
	for s := insertSlot; s < SlotsPerDay; s++ {
		if g.TaskAt(day, s) != nil {
			hasTasksAfter = true
			break
		}
	}

	if !hasTasksAfter {
		// Nothing to shift, already have space - no-op
		return g, nil
	}

	// Check if there's room to shift (would overflow?)
	lastOccupied := -1
	for s := SlotsPerDay - 1; s >= insertSlot; s-- {
		if g.TaskAt(day, s) != nil {
			lastOccupied = s
			break
		}
	}
	if lastOccupied >= 0 && lastOccupied+1 >= SlotsPerDay {
		// Would overflow - no-op
		return g, nil
	}

	newGrid := g.clone()

	// Shift everything from insertSlot onward right by 1
	for s := SlotsPerDay - 1; s > insertSlot; s-- {
		newGrid.slots[newGrid.slotIndex(day, s)] = newGrid.slots[newGrid.slotIndex(day, s-1)]
	}
	// Clear the insert slot (now empty space)
	newGrid.slots[newGrid.slotIndex(day, insertSlot)] = nil

	return newGrid, nil
}

// RemoveSpaceAt removes one empty slot at the given day/slot by shifting following slots left.
// Returns ErrNoGapToRemove if the slot is not empty or there are no tasks after it.
func (g *SlotGrid) RemoveSpaceAt(day, removeSlot int) (*SlotGrid, error) {
	if day < 0 || day >= g.config.NumDays || removeSlot < 0 || removeSlot >= SlotsPerDay {
		return nil, ErrInvalidSlotPosition
	}

	if g.isPastPosition(day, removeSlot) {
		return nil, ErrTaskAlreadyStarted
	}

	if g.TaskAt(day, removeSlot) != nil {
		return nil, ErrNoGapToRemove
	}

	hasTasksAfter := false
	for s := removeSlot + 1; s < SlotsPerDay; s++ {
		if g.TaskAt(day, s) != nil {
			hasTasksAfter = true
			break
		}
	}

	if !hasTasksAfter {
		return nil, ErrNoGapToRemove
	}

	newGrid := g.clone()

	for s := removeSlot + 1; s < SlotsPerDay; s++ {
		newGrid.slots[newGrid.slotIndex(day, s-1)] = newGrid.slots[newGrid.slotIndex(day, s)]
	}
	newGrid.slots[newGrid.slotIndex(day, SlotsPerDay-1)] = nil

	return newGrid, nil
}

// Delete removes a task from the grid and shifts following tasks left.
// Returns error if task has started.
func (g *SlotGrid) Delete(t *task.Task) (*SlotGrid, error) {
	if t == nil {
		return nil, ErrSlotTaskNotFound
	}

	if err := g.canModifyTask(t); err != nil {
		return nil, err
	}

	day, startSlot, endSlot, found := g.FindTask(t)
	if !found {
		return nil, ErrSlotTaskNotFound
	}

	newGrid := g.clone()

	// Clear the task slots
	for s := startSlot; s < endSlot; s++ {
		newGrid.slots[newGrid.slotIndex(day, s)] = nil
	}

	// Shift following tasks left
	gapStart := startSlot
	for s := endSlot; s < SlotsPerDay; s++ {
		slotTask := newGrid.TaskAt(day, s)
		if slotTask != nil {
			newGrid.slots[newGrid.slotIndex(day, gapStart)] = slotTask
			newGrid.slots[newGrid.slotIndex(day, s)] = nil
			gapStart++
		}
	}

	return newGrid, nil
}

// ============================================================================
// Debug/Print Helpers
// ============================================================================

// Print returns a string visualization of the grid for debugging.
func (g *SlotGrid) Print() string {
	var sb strings.Builder

	for d := 0; d < g.config.NumDays; d++ {
		sb.WriteString("Day ")
		sb.WriteString(string(rune('0' + d)))
		sb.WriteString(": ")

		var lastID int64 = -1
		for s := 0; s < SlotsPerDay; s++ {
			t := g.TaskAt(d, s)
			switch {
			case t == nil:
				sb.WriteRune('-')
				lastID = -1
			case t.ID != lastID:
				letter := rune('A' + (t.ID % 26))
				sb.WriteRune(letter)
				lastID = t.ID
			default:
				letter := rune('A' + (t.ID % 26))
				sb.WriteRune(letter)
			}
		}
		sb.WriteRune('\n')
	}

	return sb.String()
}

// PrintDay returns a compact string for a single day.
func (g *SlotGrid) PrintDay(day int) string {
	if day < 0 || day >= g.config.NumDays {
		return ""
	}

	var sb strings.Builder

	for s := 0; s < SlotsPerDay; s++ {
		t := g.TaskAt(day, s)
		if t == nil {
			sb.WriteRune('-')
		} else {
			letter := rune('A' + (t.ID % 26))
			sb.WriteRune(letter)
		}
	}

	return sb.String()
}

// String returns a compact multi-line representation.
func (g *SlotGrid) String() string {
	var parts []string
	for d := 0; d < g.config.NumDays; d++ {
		parts = append(parts, g.PrintDay(d))
	}
	return strings.Join(parts, "|")
}

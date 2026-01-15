package task

// WeekWindow provides a sliding window of weeks for efficient TUI navigation.
// It maintains 3 consecutive weeks: previous, current, and next.
// This allows smooth week-to-week navigation without loading on every change.
type WeekWindow struct {
	weeks [3]*Week // [0]=prev, [1]=current, [2]=next
}

// NewWeekWindow creates a window with three consecutive weeks.
// The weeks should be consecutive (prev, current, next) for proper navigation.
func NewWeekWindow(prev, current, next *Week) *WeekWindow {
	return &WeekWindow{
		weeks: [3]*Week{prev, current, next},
	}
}

// Current returns the focused (center) week.
func (w *WeekWindow) Current() *Week {
	return w.weeks[1]
}

// Previous returns the week before current.
func (w *WeekWindow) Previous() *Week {
	return w.weeks[0]
}

// Next returns the week after current.
func (w *WeekWindow) Next() *Week {
	return w.weeks[2]
}

// ShiftForward moves the window forward by one week.
// The current week becomes the previous week, the next week becomes current,
// and the provided newNext becomes the new next week.
func (w *WeekWindow) ShiftForward(newNext *Week) {
	w.weeks[0] = w.weeks[1] // old current becomes prev
	w.weeks[1] = w.weeks[2] // old next becomes current
	w.weeks[2] = newNext    // new week becomes next
}

// ShiftBackward moves the window backward by one week.
// The current week becomes the next week, the previous week becomes current,
// and the provided newPrev becomes the new previous week.
func (w *WeekWindow) ShiftBackward(newPrev *Week) {
	w.weeks[2] = w.weeks[1] // old current becomes next
	w.weeks[1] = w.weeks[0] // old prev becomes current
	w.weeks[0] = newPrev    // new week becomes prev
}

// SetCurrent replaces the current week.
// This is useful for cache invalidation after task mutations.
func (w *WeekWindow) SetCurrent(week *Week) {
	w.weeks[1] = week
}

// SetNext replaces the next week after it's been loaded.
func (w *WeekWindow) SetNext(week *Week) {
	w.weeks[2] = week
}

// SetPrevious replaces the previous week after it's been loaded.
func (w *WeekWindow) SetPrevious(week *Week) {
	w.weeks[0] = week
}

// HasNext returns true if the next week is loaded (not nil).
func (w *WeekWindow) HasNext() bool {
	return w.weeks[2] != nil
}

// HasPrevious returns true if the previous week is loaded (not nil).
func (w *WeekWindow) HasPrevious() bool {
	return w.weeks[0] != nil
}

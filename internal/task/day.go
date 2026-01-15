package task

import (
	"fmt"
	"slices"
	"time"
)

// Day holds all tasks for a single day.
type Day struct {
	Date  time.Time
	tasks []*Task // sorted by ScheduledStart
}

// NewDay creates a Day for the given date.
func NewDay(date time.Time) *Day {
	return &Day{
		Date:  truncateToDay(date),
		tasks: make([]*Task, 0),
	}
}

// NewDayWithTasks creates a Day from a slice of tasks.
// Tasks must be for the same date. Returns error if tasks overlap.
func NewDayWithTasks(date time.Time, tasks []*Task) (*Day, error) {
	d := NewDay(date)
	for _, t := range tasks {
		if err := d.AddTask(t); err != nil {
			return nil, err
		}
	}
	return d, nil
}

// Tasks returns a copy of the task slice.
func (d *Day) Tasks() []*Task {
	result := make([]*Task, len(d.tasks))
	copy(result, d.tasks)
	return result
}

// AddTask adds a task to the day, maintaining sorted order by start time.
// Returns ErrTimeBlockOverlap if the task overlaps with an existing scheduled task.
// Only checks overlap against tasks with scheduled status.
func (d *Day) AddTask(t *Task) error {
	if t == nil {
		return nil
	}

	// Only check overlap for scheduled tasks
	if t.IsScheduled() {
		if overlap := d.FindOverlappingTask(t.ScheduledStart, t.ScheduledEnd); overlap != nil {
			return fmt.Errorf("%w: %q (%s-%s) conflicts with %q (%s-%s)",
				ErrTimeBlockOverlap,
				t.Description, t.ScheduledStart, t.ScheduledEnd,
				overlap.Description, overlap.ScheduledStart, overlap.ScheduledEnd,
			)
		}
	}

	// Insert in sorted order by start time
	d.tasks = append(d.tasks, t)
	slices.SortFunc(d.tasks, func(a, b *Task) int {
		if a.ScheduledStart < b.ScheduledStart {
			return -1
		}
		if a.ScheduledStart > b.ScheduledStart {
			return 1
		}
		return 0
	})

	return nil
}

// FindOverlappingTask returns the first scheduled task that overlaps with the given time slot.
// Returns nil if no overlap is found.
func (d *Day) FindOverlappingTask(start, end string) *Task {
	for _, t := range d.tasks {
		if !t.IsScheduled() {
			continue
		}
		if TimesOverlap(start, end, t.ScheduledStart, t.ScheduledEnd) {
			return t
		}
	}
	return nil
}

// HasOverlap returns true if any scheduled task overlaps with the given time slot.
func (d *Day) HasOverlap(start, end string) bool {
	return d.FindOverlappingTask(start, end) != nil
}

// ScheduledTasks returns only tasks with scheduled status.
func (d *Day) ScheduledTasks() []*Task {
	var result []*Task
	for _, t := range d.tasks {
		if t.IsScheduled() {
			result = append(result, t)
		}
	}
	return result
}

// RemoveTask removes a task from the day by ID.
// Returns the removed task, or nil if not found.
func (d *Day) RemoveTask(taskID int64) *Task {
	for i, t := range d.tasks {
		if t.ID == taskID {
			d.tasks = append(d.tasks[:i], d.tasks[i+1:]...)
			return t
		}
	}
	return nil
}

// Len returns the number of tasks in the day.
func (d *Day) Len() int {
	return len(d.tasks)
}

// DayStats holds statistics for a single day.
type DayStats struct {
	DeepMinutes     int
	ShallowMinutes  int
	TotalBlocks     int
	CancelledBlocks int
	PostponedBlocks int
}

// TotalMinutes returns the sum of deep and shallow minutes.
func (s DayStats) TotalMinutes() int {
	return s.DeepMinutes + s.ShallowMinutes
}

// DeepPercent returns the percentage of time spent on deep work.
func (s DayStats) DeepPercent() int {
	if s.TotalMinutes() == 0 {
		return 0
	}
	return (s.DeepMinutes * 100) / s.TotalMinutes()
}

// Stats calculates statistics for the day.
func (d *Day) Stats() DayStats {
	var stats DayStats
	for _, t := range d.tasks {
		stats.TotalBlocks++
		switch t.Status {
		case StatusCancelled:
			stats.CancelledBlocks++
		case StatusPostponed:
			stats.PostponedBlocks++
		default:
			if t.IsDeep() {
				stats.DeepMinutes += t.Duration()
			} else {
				stats.ShallowMinutes += t.Duration()
			}
		}
	}
	return stats
}

// StatsWithPeakHours calculates statistics including peak hour alignment.
func (d *Day) StatsWithPeakHours(peakStart, peakEnd string) DayStatsWithPeak {
	base := d.Stats()
	var peakDeep int

	for _, t := range d.tasks {
		if t.IsScheduled() && t.IsDeep() {
			peakDeep += OverlapMinutes(t.ScheduledStart, t.ScheduledEnd, peakStart, peakEnd)
		}
	}

	return DayStatsWithPeak{
		DayStats:        base,
		PeakDeepMinutes: peakDeep,
	}
}

// DayStatsWithPeak extends DayStats with peak hour tracking.
type DayStatsWithPeak struct {
	DayStats
	PeakDeepMinutes int
}

// truncateToDay removes the time component from a time.Time.
func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

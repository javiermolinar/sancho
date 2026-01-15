package task

import (
	"fmt"
	"time"
)

// Week holds 7 days starting from Monday.
type Week struct {
	StartDate time.Time // Monday of the week
	Days      [7]*Day   // Monday (0) through Sunday (6)
}

// NewWeek creates a Week starting from the Monday of the given date.
func NewWeek(date time.Time) *Week {
	monday := startOfWeek(date)
	w := &Week{StartDate: monday}

	for i := 0; i < 7; i++ {
		dayDate := monday.AddDate(0, 0, i)
		w.Days[i] = NewDay(dayDate)
	}

	return w
}

// NewWeekFromTasks creates a Week and distributes tasks to their respective days.
// Tasks outside the week's date range are ignored.
func NewWeekFromTasks(date time.Time, tasks []*Task) *Week {
	w := NewWeek(date)

	for _, t := range tasks {
		if day := w.DayByDate(t.ScheduledDate); day != nil {
			// Ignore errors - just add what we can
			_ = day.AddTask(t)
		}
	}

	return w
}

// Day returns the Day for the given weekday (0=Monday, 6=Sunday).
// Returns nil if weekday is out of range.
func (w *Week) Day(weekday int) *Day {
	if weekday < 0 || weekday > 6 {
		return nil
	}
	return w.Days[weekday]
}

// DayByDate returns the Day for the given date, nil if not in this week.
func (w *Week) DayByDate(date time.Time) *Day {
	truncated := truncateToDay(date)
	for _, day := range w.Days {
		if day.Date.Equal(truncated) {
			return day
		}
	}
	return nil
}

// AllTasks returns all tasks across all days, sorted by date and start time.
func (w *Week) AllTasks() []*Task {
	var result []*Task
	for _, day := range w.Days {
		result = append(result, day.Tasks()...)
	}
	return result
}

// EndDate returns the Sunday of the week.
func (w *Week) EndDate() time.Time {
	return w.StartDate.AddDate(0, 0, 6)
}

// WeekStats holds aggregated statistics for the week.
type WeekStats struct {
	DeepMinutes     int
	ShallowMinutes  int
	PeakDeepMinutes int
	TotalBlocks     int
	CancelledBlocks int
	PostponedBlocks int
	DayStats        [7]DayStats
}

// TotalMinutes returns the sum of deep and shallow minutes.
func (s WeekStats) TotalMinutes() int {
	return s.DeepMinutes + s.ShallowMinutes
}

// DeepPercent returns the percentage of time spent on deep work.
func (s WeekStats) DeepPercent() int {
	if s.TotalMinutes() == 0 {
		return 0
	}
	return (s.DeepMinutes * 100) / s.TotalMinutes()
}

// PeakPercent returns the percentage of deep work during peak hours.
func (s WeekStats) PeakPercent() int {
	if s.DeepMinutes == 0 {
		return 0
	}
	return (s.PeakDeepMinutes * 100) / s.DeepMinutes
}

// Ratio returns the deep:shallow ratio as a string (e.g., "2.5:1").
func (s WeekStats) Ratio() string {
	switch {
	case s.ShallowMinutes > 0:
		r := float64(s.DeepMinutes) / float64(s.ShallowMinutes)
		return fmt.Sprintf("%.1f:1", r)
	case s.DeepMinutes > 0:
		return "∞:1"
	default:
		return "0:0"
	}
}

// BestDay returns the weekday (0=Monday) with the most deep work minutes and the minutes.
func (s WeekStats) BestDay() (weekday int, deepMinutes int) {
	weekday = -1
	for i, ds := range s.DayStats {
		if ds.DeepMinutes > deepMinutes {
			deepMinutes = ds.DeepMinutes
			weekday = i
		}
	}
	return weekday, deepMinutes
}

// Stats calculates statistics for the week.
func (w *Week) Stats() WeekStats {
	var stats WeekStats
	for i, day := range w.Days {
		ds := day.Stats()
		stats.DayStats[i] = ds
		stats.DeepMinutes += ds.DeepMinutes
		stats.ShallowMinutes += ds.ShallowMinutes
		stats.TotalBlocks += ds.TotalBlocks
		stats.CancelledBlocks += ds.CancelledBlocks
		stats.PostponedBlocks += ds.PostponedBlocks
	}
	return stats
}

// StatsWithPeakHours calculates statistics including peak hour alignment.
func (w *Week) StatsWithPeakHours(peakStart, peakEnd string) WeekStats {
	stats := w.Stats()

	for _, day := range w.Days {
		peakStats := day.StatsWithPeakHours(peakStart, peakEnd)
		stats.PeakDeepMinutes += peakStats.PeakDeepMinutes
	}

	return stats
}

// WeekdayName returns the name of the weekday (0=Monday).
func WeekdayName(weekday int) string {
	names := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	if weekday < 0 || weekday > 6 {
		return ""
	}
	return names[weekday]
}

// WeekdayShortName returns the short name of the weekday (0=Monday).
func WeekdayShortName(weekday int) string {
	names := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	if weekday < 0 || weekday > 6 {
		return ""
	}
	return names[weekday]
}

// startOfWeek returns the Monday of the week containing the given date.
func startOfWeek(t time.Time) time.Time {
	t = truncateToDay(t)
	weekday := int(t.Weekday())
	// Convert Sunday (0) to 7 for easier calculation
	if weekday == 0 {
		weekday = 7
	}
	// Go back to Monday (weekday 1)
	return t.AddDate(0, 0, -(weekday - 1))
}

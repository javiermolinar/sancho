// Package scheduler provides time-aware scheduling logic for tasks.
package scheduler

import (
	"strings"
	"time"
)

// Scheduler provides time-aware scheduling operations.
type Scheduler struct {
	workdays map[string]bool
	dayStart string // "HH:MM"
	dayEnd   string // "HH:MM"
}

// New creates a new Scheduler with the given configuration.
func New(workdays []string, dayStart, dayEnd string) *Scheduler {
	wd := make(map[string]bool)
	for _, d := range workdays {
		wd[strings.ToLower(d)] = true
	}
	return &Scheduler{
		workdays: wd,
		dayStart: dayStart,
		dayEnd:   dayEnd,
	}
}

// AvailableSlot represents an available time slot for scheduling.
type AvailableSlot struct {
	Date  time.Time
	Start string // "HH:MM"
	End   string // "HH:MM"
}

// NextAvailableStart returns the next available start time for scheduling.
// If now is before dayStart, returns dayStart of today (if workday) or next workday.
// If now is during work hours, returns now (rounded to next 15 min).
// If now is after dayEnd, returns dayStart of next workday.
func (s *Scheduler) NextAvailableStart(now time.Time) AvailableSlot {
	nowTime := now.Format("15:04")
	weekday := strings.ToLower(now.Weekday().String())

	// If today is a workday and within work hours
	if s.workdays[weekday] {
		if nowTime < s.dayStart {
			// Before work starts - use day start
			return AvailableSlot{
				Date:  now,
				Start: s.dayStart,
				End:   s.dayEnd,
			}
		}
		if nowTime < s.dayEnd {
			// During work hours - round up to next 15 min
			start := roundUpTo15Min(now).Format("15:04")
			if start >= s.dayEnd {
				// Too late today, go to next workday
				return s.nextWorkday(now)
			}
			return AvailableSlot{
				Date:  now,
				Start: start,
				End:   s.dayEnd,
			}
		}
	}

	// After work hours or not a workday - find next workday
	return s.nextWorkday(now)
}

// nextWorkday finds the next workday starting from the day after the given time.
func (s *Scheduler) nextWorkday(from time.Time) AvailableSlot {
	next := from.AddDate(0, 0, 1)
	for range 7 {
		weekday := strings.ToLower(next.Weekday().String())
		if s.workdays[weekday] {
			return AvailableSlot{
				Date:  next,
				Start: s.dayStart,
				End:   s.dayEnd,
			}
		}
		next = next.AddDate(0, 0, 1)
	}
	// Fallback: should never happen if workdays is configured correctly
	return AvailableSlot{
		Date:  from.AddDate(0, 0, 1),
		Start: s.dayStart,
		End:   s.dayEnd,
	}
}

// IsWorkday returns true if the given time falls on a configured workday.
func (s *Scheduler) IsWorkday(t time.Time) bool {
	weekday := strings.ToLower(t.Weekday().String())
	return s.workdays[weekday]
}

// IsWithinWorkHours returns true if the given time is within configured work hours.
func (s *Scheduler) IsWithinWorkHours(t time.Time) bool {
	if !s.IsWorkday(t) {
		return false
	}
	nowTime := t.Format("15:04")
	return nowTime >= s.dayStart && nowTime < s.dayEnd
}

// AvailableMinutes returns the number of minutes available for scheduling
// from the given start time until day end.
func (s *Scheduler) AvailableMinutes(slot AvailableSlot) int {
	start := parseTime(slot.Start)
	end := parseTime(slot.End)
	if start >= end {
		return 0
	}
	return end - start
}

// CanFit returns true if a task of the given duration (in minutes) can fit
// starting at the given time on the given date.
func (s *Scheduler) CanFit(date time.Time, startTime string, durationMinutes int) bool {
	if !s.IsWorkday(date) {
		return false
	}

	start := parseTime(startTime)
	end := parseTime(s.dayEnd)

	if start < parseTime(s.dayStart) || start >= end {
		return false
	}

	return start+durationMinutes <= end
}

// ValidateTimeSlot checks if a time slot is valid for the given date.
// Returns an error message if invalid, empty string if valid.
func (s *Scheduler) ValidateTimeSlot(date time.Time, start, end string) string {
	if !s.IsWorkday(date) {
		return "not a workday"
	}

	startMin := parseTime(start)
	endMin := parseTime(end)
	dayStartMin := parseTime(s.dayStart)
	dayEndMin := parseTime(s.dayEnd)

	if startMin >= endMin {
		return "start time must be before end time"
	}
	if startMin < dayStartMin {
		return "start time is before workday start"
	}
	if endMin > dayEndMin {
		return "end time is after workday end"
	}

	return ""
}

// DayStart returns the configured day start time.
func (s *Scheduler) DayStart() string {
	return s.dayStart
}

// DayEnd returns the configured day end time.
func (s *Scheduler) DayEnd() string {
	return s.dayEnd
}

// ValidateTimeSlotAnyDay validates a time slot without checking if the date is a workday.
// Used when the user explicitly provides a date.
// Returns an error message if invalid, empty string if valid.
func (s *Scheduler) ValidateTimeSlotAnyDay(start, end string) string {
	startMin := parseTime(start)
	endMin := parseTime(end)
	dayStartMin := parseTime(s.dayStart)
	dayEndMin := parseTime(s.dayEnd)

	if startMin >= endMin {
		return "start time must be before end time"
	}
	if startMin < dayStartMin {
		return "start time is before workday start"
	}
	if endMin > dayEndMin {
		return "end time is after workday end"
	}

	return ""
}

// CanFitAnyDay checks if a task of the given duration can fit within day boundaries
// without checking if the date is a workday.
// Used when the user explicitly provides a date.
func (s *Scheduler) CanFitAnyDay(startTime string, durationMinutes int) bool {
	start := parseTime(startTime)
	end := parseTime(s.dayEnd)

	if start < parseTime(s.dayStart) || start >= end {
		return false
	}

	return start+durationMinutes <= end
}

// roundUpTo15Min rounds a time up to the next 15-minute boundary.
func roundUpTo15Min(t time.Time) time.Time {
	minute := t.Minute()
	remainder := minute % 15
	if remainder == 0 && t.Second() == 0 && t.Nanosecond() == 0 {
		return t
	}
	return t.Add(time.Duration(15-remainder) * time.Minute).Truncate(time.Minute)
}

// parseTime parses "HH:MM" to minutes since midnight.
func parseTime(s string) int {
	if len(s) < 5 {
		return 0
	}
	h := int(s[0]-'0')*10 + int(s[1]-'0')
	m := int(s[3]-'0')*10 + int(s[4]-'0')
	return h*60 + m
}

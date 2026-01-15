// Package dateutil provides date parsing and validation utilities.
package dateutil

import (
	"errors"
	"strings"
	"time"
)

// Validation errors.
var (
	ErrInvalidDateFormat  = errors.New("date must be in YYYY-MM-DD format")
	ErrEndDateBeforeStart = errors.New("end date must be on or after start date")
	ErrDateInPast         = errors.New("cannot schedule in the past")
)

// weekdayMap maps weekday names to time.Weekday values.
var weekdayMap = map[string]time.Weekday{
	"sunday":    time.Sunday,
	"monday":    time.Monday,
	"tuesday":   time.Tuesday,
	"wednesday": time.Wednesday,
	"thursday":  time.Thursday,
	"friday":    time.Friday,
	"saturday":  time.Saturday,
}

// DateRange represents a validated date range.
type DateRange struct {
	Start time.Time
	End   time.Time
}

// NewDateRange creates a new DateRange with validation.
// startDate can be empty (defaults to today) or in YYYY-MM-DD format.
// endDate can be empty (defaults to startDate) or in YYYY-MM-DD format.
// Returns an error if endDate is before startDate.
func NewDateRange(startDate, endDate string) (*DateRange, error) {
	start, err := ParseDate(startDate)
	if err != nil {
		return nil, err
	}

	var end time.Time
	if endDate == "" {
		end = start
	} else {
		end, err = ParseDate(endDate)
		if err != nil {
			return nil, err
		}
	}

	if end.Before(start) {
		return nil, ErrEndDateBeforeStart
	}

	return &DateRange{Start: start, End: end}, nil
}

// ParseDate parses a date string in YYYY-MM-DD format.
// If the string is empty, returns today's date.
func ParseDate(s string) (time.Time, error) {
	if s == "" {
		return TruncateToDay(time.Now()), nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, ErrInvalidDateFormat
	}
	return t, nil
}

// WeekRange returns the Monday and Sunday of the ISO week containing t.
func WeekRange(t time.Time) (monday, sunday time.Time) {
	t = TruncateToDay(t)
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday becomes day 7 in ISO week
	}
	monday = t.AddDate(0, 0, -(weekday - 1))
	sunday = monday.AddDate(0, 0, 6)
	return monday, sunday
}

// TruncateToDay returns t with time set to midnight.
func TruncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// ParseRelativeDate parses a date string that can be:
//   - Empty string or "today": returns relativeTo date
//   - Absolute date: "2025-01-15" (YYYY-MM-DD)
//   - Keywords: "tomorrow"
//   - Weekday names: "monday" through "sunday" (next occurrence, always future)
//   - Next prefixed: "next-monday" through "next-sunday", "next-week"
//
// All inputs are case-insensitive.
// Returns ErrDateInPast if the resulting date is before relativeTo (truncated to day).
// Returns ErrInvalidDateFormat for unrecognized input.
func ParseRelativeDate(s string, relativeTo time.Time) (time.Time, error) {
	today := TruncateToDay(relativeTo)
	input := strings.ToLower(strings.TrimSpace(s))

	// Empty or "today"
	if input == "" || input == "today" {
		return today, nil
	}

	// "tomorrow"
	if input == "tomorrow" {
		return today.AddDate(0, 0, 1), nil
	}

	// "next-week" - same weekday, +7 days
	if input == "next-week" {
		return today.AddDate(0, 0, 7), nil
	}

	// "next-monday", "next-tuesday", etc.
	if strings.HasPrefix(input, "next-") {
		weekdayName := strings.TrimPrefix(input, "next-")
		if targetDay, ok := weekdayMap[weekdayName]; ok {
			return nextWeekday(today, targetDay), nil
		}
		return time.Time{}, ErrInvalidDateFormat
	}

	// Weekday names: "monday", "tuesday", etc.
	if targetDay, ok := weekdayMap[input]; ok {
		return nextWeekday(today, targetDay), nil
	}

	// Absolute date: YYYY-MM-DD
	result, err := time.Parse("2006-01-02", input)
	if err != nil {
		return time.Time{}, ErrInvalidDateFormat
	}

	// Check for past date
	if result.Before(today) {
		return time.Time{}, ErrDateInPast
	}

	return result, nil
}

// nextWeekday returns the next occurrence of the given weekday after today.
// If today is the target weekday, returns one week from today.
func nextWeekday(today time.Time, target time.Weekday) time.Time {
	current := today.Weekday()
	daysUntil := int(target) - int(current)
	if daysUntil <= 0 {
		daysUntil += 7
	}
	return today.AddDate(0, 0, daysUntil)
}

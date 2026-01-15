package dwplanner

import (
	"testing"
	"time"

	"github.com/javiermolinar/sancho/internal/config"
	"github.com/javiermolinar/sancho/internal/scheduler"
)

func TestNextWorkdayCalculation(t *testing.T) {
	cfg := &config.Config{
		Schedule: config.ScheduleConfig{
			Workdays: []string{"monday", "tuesday", "wednesday", "thursday", "friday"},
			DayStart: "09:00",
			DayEnd:   "17:00",
		},
	}

	sched := scheduler.New(cfg.Schedule.Workdays, cfg.Schedule.DayStart, cfg.Schedule.DayEnd)

	tests := []struct {
		name         string
		todayDate    time.Time
		expectedNext time.Weekday
	}{
		{
			name:         "Friday -> Monday",
			todayDate:    time.Date(2025, 1, 10, 12, 0, 0, 0, time.Local), // Friday
			expectedNext: time.Monday,
		},
		{
			name:         "Monday -> Tuesday",
			todayDate:    time.Date(2025, 1, 6, 12, 0, 0, 0, time.Local), // Monday
			expectedNext: time.Tuesday,
		},
		{
			name:         "Thursday -> Friday",
			todayDate:    time.Date(2025, 1, 9, 12, 0, 0, 0, time.Local), // Thursday
			expectedNext: time.Friday,
		},
		{
			name:         "Saturday -> Monday",
			todayDate:    time.Date(2025, 1, 11, 12, 0, 0, 0, time.Local), // Saturday
			expectedNext: time.Monday,
		},
		{
			name:         "Sunday -> Monday",
			todayDate:    time.Date(2025, 1, 12, 12, 0, 0, 0, time.Local), // Sunday
			expectedNext: time.Monday,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate finding next workday after today (at end of day)
			nextSlot := sched.NextAvailableStart(
				time.Date(tc.todayDate.Year(), tc.todayDate.Month(), tc.todayDate.Day(),
					23, 59, 0, 0, time.Local),
			)

			if nextSlot.Date.Weekday() != tc.expectedNext {
				t.Errorf("expected next workday to be %s, got %s (date: %s)",
					tc.expectedNext, nextSlot.Date.Weekday(), nextSlot.Date.Format("2006-01-02"))
			}
		})
	}
}

func TestNextAvailableStart_CloseToMidnight(t *testing.T) {
	sched := scheduler.New(
		[]string{"monday", "tuesday", "wednesday", "thursday", "friday"},
		"09:00",
		"17:00",
	)

	tests := []struct {
		name            string
		now             time.Time
		expectedDate    time.Time
		expectedStart   string
		expectedWeekday time.Weekday
	}{
		{
			name:            "Thursday 23:30 -> Friday 09:00",
			now:             time.Date(2025, 1, 9, 23, 30, 0, 0, time.Local), // Thursday 23:30
			expectedDate:    time.Date(2025, 1, 10, 0, 0, 0, 0, time.Local),  // Friday
			expectedStart:   "09:00",
			expectedWeekday: time.Friday,
		},
		{
			name:            "Friday 23:59 -> Monday 09:00",
			now:             time.Date(2025, 1, 10, 23, 59, 0, 0, time.Local), // Friday 23:59
			expectedDate:    time.Date(2025, 1, 13, 0, 0, 0, 0, time.Local),   // Monday
			expectedStart:   "09:00",
			expectedWeekday: time.Monday,
		},
		{
			name:            "Saturday 00:01 -> Monday 09:00",
			now:             time.Date(2025, 1, 11, 0, 1, 0, 0, time.Local), // Saturday 00:01
			expectedDate:    time.Date(2025, 1, 13, 0, 0, 0, 0, time.Local), // Monday
			expectedStart:   "09:00",
			expectedWeekday: time.Monday,
		},
		{
			name:            "Monday 23:00 -> Tuesday 09:00",
			now:             time.Date(2025, 1, 6, 23, 0, 0, 0, time.Local), // Monday 23:00
			expectedDate:    time.Date(2025, 1, 7, 0, 0, 0, 0, time.Local),  // Tuesday
			expectedStart:   "09:00",
			expectedWeekday: time.Tuesday,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			slot := sched.NextAvailableStart(tc.now)

			if slot.Date.Weekday() != tc.expectedWeekday {
				t.Errorf("expected weekday %s, got %s (date: %s)",
					tc.expectedWeekday, slot.Date.Weekday(), slot.Date.Format("2006-01-02"))
			}
			if slot.Start != tc.expectedStart {
				t.Errorf("expected start %s, got %s", tc.expectedStart, slot.Start)
			}
		})
	}
}

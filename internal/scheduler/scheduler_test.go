package scheduler

import (
	"testing"
	"time"
)

func TestNextAvailableStart_BeforeWorkHours(t *testing.T) {
	s := New([]string{"monday", "tuesday", "wednesday", "thursday", "friday"}, "09:00", "17:00")

	// Monday at 7:30 AM
	now := time.Date(2025, 1, 6, 7, 30, 0, 0, time.Local) // Monday
	slot := s.NextAvailableStart(now)

	if slot.Start != "09:00" {
		t.Errorf("expected start 09:00, got %s", slot.Start)
	}
	if slot.End != "17:00" {
		t.Errorf("expected end 17:00, got %s", slot.End)
	}
	if slot.Date.Day() != 6 {
		t.Errorf("expected same day (6), got %d", slot.Date.Day())
	}
}

func TestNextAvailableStart_DuringWorkHours(t *testing.T) {
	s := New([]string{"monday", "tuesday", "wednesday", "thursday", "friday"}, "09:00", "17:00")

	// Monday at 10:23 AM - should round up to 10:30
	now := time.Date(2025, 1, 6, 10, 23, 0, 0, time.Local)
	slot := s.NextAvailableStart(now)

	if slot.Start != "10:30" {
		t.Errorf("expected start 10:30, got %s", slot.Start)
	}
	if slot.Date.Day() != 6 {
		t.Errorf("expected same day (6), got %d", slot.Date.Day())
	}
}

func TestNextAvailableStart_ExactlyOn15Min(t *testing.T) {
	s := New([]string{"monday", "tuesday", "wednesday", "thursday", "friday"}, "09:00", "17:00")

	// Monday at exactly 10:30:00
	now := time.Date(2025, 1, 6, 10, 30, 0, 0, time.Local)
	slot := s.NextAvailableStart(now)

	if slot.Start != "10:30" {
		t.Errorf("expected start 10:30, got %s", slot.Start)
	}
}

func TestNextAvailableStart_AfterWorkHours(t *testing.T) {
	s := New([]string{"monday", "tuesday", "wednesday", "thursday", "friday"}, "09:00", "17:00")

	// Monday at 6:00 PM - should go to Tuesday
	now := time.Date(2025, 1, 6, 18, 0, 0, 0, time.Local)
	slot := s.NextAvailableStart(now)

	if slot.Start != "09:00" {
		t.Errorf("expected start 09:00, got %s", slot.Start)
	}
	if slot.Date.Day() != 7 {
		t.Errorf("expected next day (7), got %d", slot.Date.Day())
	}
}

func TestNextAvailableStart_Weekend(t *testing.T) {
	s := New([]string{"monday", "tuesday", "wednesday", "thursday", "friday"}, "09:00", "17:00")

	// Saturday - should go to Monday
	now := time.Date(2025, 1, 4, 10, 0, 0, 0, time.Local) // Saturday
	slot := s.NextAvailableStart(now)

	if slot.Date.Weekday() != time.Monday {
		t.Errorf("expected Monday, got %s", slot.Date.Weekday())
	}
	if slot.Start != "09:00" {
		t.Errorf("expected start 09:00, got %s", slot.Start)
	}
}

func TestNextAvailableStart_FridayEvening(t *testing.T) {
	s := New([]string{"monday", "tuesday", "wednesday", "thursday", "friday"}, "09:00", "17:00")

	// Friday at 6:00 PM - should go to Monday
	now := time.Date(2025, 1, 10, 18, 0, 0, 0, time.Local) // Friday
	slot := s.NextAvailableStart(now)

	if slot.Date.Weekday() != time.Monday {
		t.Errorf("expected Monday, got %s", slot.Date.Weekday())
	}
	if slot.Date.Day() != 13 {
		t.Errorf("expected day 13, got %d", slot.Date.Day())
	}
}

func TestIsWorkday(t *testing.T) {
	s := New([]string{"monday", "tuesday", "wednesday", "thursday", "friday"}, "09:00", "17:00")

	tests := []struct {
		date time.Time
		want bool
	}{
		{time.Date(2025, 1, 6, 10, 0, 0, 0, time.Local), true},  // Monday
		{time.Date(2025, 1, 7, 10, 0, 0, 0, time.Local), true},  // Tuesday
		{time.Date(2025, 1, 4, 10, 0, 0, 0, time.Local), false}, // Saturday
		{time.Date(2025, 1, 5, 10, 0, 0, 0, time.Local), false}, // Sunday
	}

	for _, tc := range tests {
		t.Run(tc.date.Weekday().String(), func(t *testing.T) {
			got := s.IsWorkday(tc.date)
			if got != tc.want {
				t.Errorf("IsWorkday(%s) = %v, want %v", tc.date.Weekday(), got, tc.want)
			}
		})
	}
}

func TestIsWithinWorkHours(t *testing.T) {
	s := New([]string{"monday", "tuesday", "wednesday", "thursday", "friday"}, "09:00", "17:00")

	tests := []struct {
		name string
		date time.Time
		want bool
	}{
		{"Monday 10am", time.Date(2025, 1, 6, 10, 0, 0, 0, time.Local), true},
		{"Monday 8am", time.Date(2025, 1, 6, 8, 0, 0, 0, time.Local), false},
		{"Monday 5pm", time.Date(2025, 1, 6, 17, 0, 0, 0, time.Local), false}, // exactly at end
		{"Monday 4:59pm", time.Date(2025, 1, 6, 16, 59, 0, 0, time.Local), true},
		{"Saturday 10am", time.Date(2025, 1, 4, 10, 0, 0, 0, time.Local), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := s.IsWithinWorkHours(tc.date)
			if got != tc.want {
				t.Errorf("IsWithinWorkHours = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAvailableMinutes(t *testing.T) {
	s := New([]string{"monday"}, "09:00", "17:00")

	tests := []struct {
		slot AvailableSlot
		want int
	}{
		{AvailableSlot{Start: "09:00", End: "17:00"}, 480},
		{AvailableSlot{Start: "12:00", End: "17:00"}, 300},
		{AvailableSlot{Start: "16:30", End: "17:00"}, 30},
		{AvailableSlot{Start: "17:00", End: "17:00"}, 0},
	}

	for _, tc := range tests {
		t.Run(tc.slot.Start+"-"+tc.slot.End, func(t *testing.T) {
			got := s.AvailableMinutes(tc.slot)
			if got != tc.want {
				t.Errorf("AvailableMinutes = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestCanFit(t *testing.T) {
	s := New([]string{"monday", "tuesday", "wednesday", "thursday", "friday"}, "09:00", "17:00")

	monday := time.Date(2025, 1, 6, 0, 0, 0, 0, time.Local)
	saturday := time.Date(2025, 1, 4, 0, 0, 0, 0, time.Local)

	tests := []struct {
		name     string
		date     time.Time
		start    string
		duration int
		want     bool
	}{
		{"2h at 9am", monday, "09:00", 120, true},
		{"8h at 9am", monday, "09:00", 480, true},
		{"9h at 9am", monday, "09:00", 540, false},
		{"1h at 4pm", monday, "16:00", 60, true},
		{"2h at 4pm", monday, "16:00", 120, false},
		{"before work", monday, "08:00", 60, false},
		{"weekend", saturday, "10:00", 60, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := s.CanFit(tc.date, tc.start, tc.duration)
			if got != tc.want {
				t.Errorf("CanFit = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestValidateTimeSlot(t *testing.T) {
	s := New([]string{"monday", "tuesday", "wednesday", "thursday", "friday"}, "09:00", "17:00")

	monday := time.Date(2025, 1, 6, 0, 0, 0, 0, time.Local)
	saturday := time.Date(2025, 1, 4, 0, 0, 0, 0, time.Local)

	tests := []struct {
		name  string
		date  time.Time
		start string
		end   string
		want  string
	}{
		{"valid slot", monday, "09:00", "11:00", ""},
		{"weekend", saturday, "09:00", "11:00", "not a workday"},
		{"start >= end", monday, "11:00", "09:00", "start time must be before end time"},
		{"before day start", monday, "08:00", "10:00", "start time is before workday start"},
		{"after day end", monday, "16:00", "18:00", "end time is after workday end"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := s.ValidateTimeSlot(tc.date, tc.start, tc.end)
			if got != tc.want {
				t.Errorf("ValidateTimeSlot = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRoundUpTo15Min(t *testing.T) {
	tests := []struct {
		input time.Time
		want  string
	}{
		{time.Date(2025, 1, 6, 10, 0, 0, 0, time.Local), "10:00"},
		{time.Date(2025, 1, 6, 10, 1, 0, 0, time.Local), "10:15"},
		{time.Date(2025, 1, 6, 10, 14, 0, 0, time.Local), "10:15"},
		{time.Date(2025, 1, 6, 10, 15, 0, 0, time.Local), "10:15"},
		{time.Date(2025, 1, 6, 10, 16, 0, 0, time.Local), "10:30"},
		{time.Date(2025, 1, 6, 10, 44, 0, 0, time.Local), "10:45"},
		{time.Date(2025, 1, 6, 10, 46, 0, 0, time.Local), "11:00"},
		{time.Date(2025, 1, 6, 10, 0, 1, 0, time.Local), "10:15"}, // has seconds
	}

	for _, tc := range tests {
		t.Run(tc.input.Format("15:04:05"), func(t *testing.T) {
			got := roundUpTo15Min(tc.input).Format("15:04")
			if got != tc.want {
				t.Errorf("roundUpTo15Min = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"00:00", 0},
		{"09:00", 540},
		{"12:30", 750},
		{"17:00", 1020},
		{"23:59", 1439},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := parseTime(tc.input)
			if got != tc.want {
				t.Errorf("parseTime(%s) = %d, want %d", tc.input, got, tc.want)
			}
		})
	}
}

func TestValidateTimeSlotAnyDay(t *testing.T) {
	s := New([]string{"monday", "tuesday", "wednesday", "thursday", "friday"}, "09:00", "17:00")

	tests := []struct {
		name  string
		start string
		end   string
		want  string
	}{
		{"valid slot", "09:00", "11:00", ""},
		{"valid slot end of day", "15:00", "17:00", ""},
		{"start >= end", "11:00", "09:00", "start time must be before end time"},
		{"start equals end", "10:00", "10:00", "start time must be before end time"},
		{"before day start", "08:00", "10:00", "start time is before workday start"},
		{"after day end", "16:00", "18:00", "end time is after workday end"},
		{"both outside hours", "07:00", "19:00", "start time is before workday start"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := s.ValidateTimeSlotAnyDay(tc.start, tc.end)
			if got != tc.want {
				t.Errorf("ValidateTimeSlotAnyDay = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestValidateTimeSlotAnyDay_Weekend(t *testing.T) {
	// Key test: ValidateTimeSlotAnyDay should NOT check workday
	// Even though scheduler is configured with Mon-Fri, it should accept the slot
	s := New([]string{"monday", "tuesday", "wednesday", "thursday", "friday"}, "09:00", "17:00")

	// This should pass - ValidateTimeSlotAnyDay doesn't check the date
	got := s.ValidateTimeSlotAnyDay("10:00", "12:00")
	if got != "" {
		t.Errorf("ValidateTimeSlotAnyDay should allow any valid time slot, got error: %q", got)
	}
}

func TestCanFitAnyDay(t *testing.T) {
	s := New([]string{"monday", "tuesday", "wednesday", "thursday", "friday"}, "09:00", "17:00")

	tests := []struct {
		name     string
		start    string
		duration int
		want     bool
	}{
		{"2h at 9am", "09:00", 120, true},
		{"8h at 9am", "09:00", 480, true},
		{"9h at 9am", "09:00", 540, false},
		{"1h at 4pm", "16:00", 60, true},
		{"2h at 4pm", "16:00", 120, false},
		{"before work hours", "08:00", 60, false},
		{"at day end", "17:00", 30, false},
		{"exact fit to end", "16:30", 30, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := s.CanFitAnyDay(tc.start, tc.duration)
			if got != tc.want {
				t.Errorf("CanFitAnyDay = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCanFitAnyDay_VsCanFit_WeekendComparison(t *testing.T) {
	s := New([]string{"monday", "tuesday", "wednesday", "thursday", "friday"}, "09:00", "17:00")

	saturday := time.Date(2025, 1, 4, 0, 0, 0, 0, time.Local)

	// CanFit rejects weekend
	if s.CanFit(saturday, "10:00", 60) {
		t.Error("CanFit should reject weekend")
	}

	// CanFitAnyDay allows the same time slot (doesn't check date)
	if !s.CanFitAnyDay("10:00", 60) {
		t.Error("CanFitAnyDay should allow valid time slot regardless of date")
	}
}

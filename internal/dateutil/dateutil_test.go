package dateutil

import (
	"errors"
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	t.Run("valid date", func(t *testing.T) {
		got, err := ParseDate("2025-01-15")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
		if !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("empty defaults to today", func(t *testing.T) {
		got, err := ParseDate("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		today := TruncateToDay(time.Now())
		if !got.Equal(today) {
			t.Errorf("got %v, want %v", got, today)
		}
	})

	t.Run("invalid format", func(t *testing.T) {
		_, err := ParseDate("01-15-2025")
		if !errors.Is(err, ErrInvalidDateFormat) {
			t.Errorf("got error %v, want %v", err, ErrInvalidDateFormat)
		}
	})
}

func TestNewDateRange(t *testing.T) {
	t.Run("valid date range", func(t *testing.T) {
		dr, err := NewDateRange("2025-01-15", "2025-01-20")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expectedStart := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
		expectedEnd := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)
		if !dr.Start.Equal(expectedStart) {
			t.Errorf("got start %v, want %v", dr.Start, expectedStart)
		}
		if !dr.End.Equal(expectedEnd) {
			t.Errorf("got end %v, want %v", dr.End, expectedEnd)
		}
	})

	t.Run("same start and end date", func(t *testing.T) {
		dr, err := NewDateRange("2025-01-15", "2025-01-15")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !dr.Start.Equal(dr.End) {
			t.Errorf("expected start and end to be equal, got %v and %v", dr.Start, dr.End)
		}
	})

	t.Run("empty start defaults to today", func(t *testing.T) {
		dr, err := NewDateRange("", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		today := TruncateToDay(time.Now())
		if !dr.Start.Equal(today) {
			t.Errorf("got start %v, want %v", dr.Start, today)
		}
		if !dr.End.Equal(today) {
			t.Errorf("got end %v, want %v", dr.End, today)
		}
	})

	t.Run("empty end defaults to start", func(t *testing.T) {
		dr, err := NewDateRange("2025-01-15", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expectedDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
		if !dr.Start.Equal(expectedDate) {
			t.Errorf("got start %v, want %v", dr.Start, expectedDate)
		}
		if !dr.End.Equal(expectedDate) {
			t.Errorf("got end %v, want %v", dr.End, expectedDate)
		}
	})
}

func TestNewDateRange_Errors(t *testing.T) {
	tests := []struct {
		name      string
		startDate string
		endDate   string
		wantErr   error
	}{
		{
			name:      "invalid start date format",
			startDate: "01-15-2025",
			endDate:   "",
			wantErr:   ErrInvalidDateFormat,
		},
		{
			name:      "invalid end date format",
			startDate: "2025-01-15",
			endDate:   "01-20-2025",
			wantErr:   ErrInvalidDateFormat,
		},
		{
			name:      "end date before start date",
			startDate: "2025-01-20",
			endDate:   "2025-01-15",
			wantErr:   ErrEndDateBeforeStart,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDateRange(tt.startDate, tt.endDate)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestWeekRange(t *testing.T) {
	tests := []struct {
		name       string
		input      time.Time
		wantMonday time.Time
		wantSunday time.Time
	}{
		{
			name:       "Monday input returns same Monday",
			input:      time.Date(2025, 1, 6, 10, 30, 0, 0, time.UTC), // Monday
			wantMonday: time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC),
			wantSunday: time.Date(2025, 1, 12, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "Wednesday returns previous Monday",
			input:      time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC), // Wednesday
			wantMonday: time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC),
			wantSunday: time.Date(2025, 1, 12, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "Sunday returns previous Monday and same Sunday",
			input:      time.Date(2025, 1, 12, 23, 59, 0, 0, time.UTC), // Sunday
			wantMonday: time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC),
			wantSunday: time.Date(2025, 1, 12, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "Friday returns previous Monday",
			input:      time.Date(2025, 1, 10, 9, 0, 0, 0, time.UTC), // Friday
			wantMonday: time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC),
			wantSunday: time.Date(2025, 1, 12, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "Saturday returns previous Monday",
			input:      time.Date(2025, 1, 11, 12, 0, 0, 0, time.UTC), // Saturday
			wantMonday: time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC),
			wantSunday: time.Date(2025, 1, 12, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMonday, gotSunday := WeekRange(tt.input)
			if !gotMonday.Equal(tt.wantMonday) {
				t.Errorf("monday: got %v, want %v", gotMonday, tt.wantMonday)
			}
			if !gotSunday.Equal(tt.wantSunday) {
				t.Errorf("sunday: got %v, want %v", gotSunday, tt.wantSunday)
			}
		})
	}
}

func TestTruncateToDay(t *testing.T) {
	input := time.Date(2025, 1, 15, 14, 30, 45, 123456789, time.UTC)
	got := TruncateToDay(input)
	want := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParseRelativeDate(t *testing.T) {
	// Reference date: Friday, January 10, 2025
	friday := time.Date(2025, 1, 10, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		input      string
		relativeTo time.Time
		want       time.Time
		wantErr    error
	}{
		// Empty and today
		{
			name:       "empty returns today",
			input:      "",
			relativeTo: friday,
			want:       time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "today keyword",
			input:      "today",
			relativeTo: friday,
			want:       time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "TODAY uppercase",
			input:      "TODAY",
			relativeTo: friday,
			want:       time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
		},

		// Tomorrow
		{
			name:       "tomorrow from friday",
			input:      "tomorrow",
			relativeTo: friday,
			want:       time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC), // Saturday
		},
		{
			name:       "TOMORROW uppercase",
			input:      "TOMORROW",
			relativeTo: friday,
			want:       time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC),
		},

		// Weekday names from Friday
		{
			name:       "saturday from friday",
			input:      "saturday",
			relativeTo: friday,
			want:       time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC), // +1 day
		},
		{
			name:       "sunday from friday",
			input:      "sunday",
			relativeTo: friday,
			want:       time.Date(2025, 1, 12, 0, 0, 0, 0, time.UTC), // +2 days
		},
		{
			name:       "monday from friday",
			input:      "monday",
			relativeTo: friday,
			want:       time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC), // +3 days
		},
		{
			name:       "tuesday from friday",
			input:      "tuesday",
			relativeTo: friday,
			want:       time.Date(2025, 1, 14, 0, 0, 0, 0, time.UTC), // +4 days
		},
		{
			name:       "wednesday from friday",
			input:      "wednesday",
			relativeTo: friday,
			want:       time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), // +5 days
		},
		{
			name:       "thursday from friday",
			input:      "thursday",
			relativeTo: friday,
			want:       time.Date(2025, 1, 16, 0, 0, 0, 0, time.UTC), // +6 days
		},
		{
			name:       "friday from friday returns next friday",
			input:      "friday",
			relativeTo: friday,
			want:       time.Date(2025, 1, 17, 0, 0, 0, 0, time.UTC), // +7 days
		},

		// Weekday names from Monday
		{
			name:       "monday from monday returns next monday",
			input:      "monday",
			relativeTo: time.Date(2025, 1, 13, 10, 0, 0, 0, time.UTC), // Monday
			want:       time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),  // +7 days
		},
		{
			name:       "tuesday from monday",
			input:      "tuesday",
			relativeTo: time.Date(2025, 1, 13, 10, 0, 0, 0, time.UTC), // Monday
			want:       time.Date(2025, 1, 14, 0, 0, 0, 0, time.UTC),  // +1 day
		},
		{
			name:       "sunday from monday",
			input:      "sunday",
			relativeTo: time.Date(2025, 1, 13, 10, 0, 0, 0, time.UTC), // Monday
			want:       time.Date(2025, 1, 19, 0, 0, 0, 0, time.UTC),  // +6 days
		},

		// Next-prefixed weekdays
		{
			name:       "next-monday from friday",
			input:      "next-monday",
			relativeTo: friday,
			want:       time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "next-friday from friday",
			input:      "next-friday",
			relativeTo: friday,
			want:       time.Date(2025, 1, 17, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "next-saturday from friday",
			input:      "next-saturday",
			relativeTo: friday,
			want:       time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "NEXT-MONDAY uppercase",
			input:      "NEXT-MONDAY",
			relativeTo: friday,
			want:       time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC),
		},

		// Next-week
		{
			name:       "next-week from friday",
			input:      "next-week",
			relativeTo: friday,
			want:       time.Date(2025, 1, 17, 0, 0, 0, 0, time.UTC), // Same weekday +7
		},
		{
			name:       "next-week from monday",
			input:      "next-week",
			relativeTo: time.Date(2025, 1, 13, 10, 0, 0, 0, time.UTC),
			want:       time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "next-week from sunday",
			input:      "next-week",
			relativeTo: time.Date(2025, 1, 12, 10, 0, 0, 0, time.UTC),
			want:       time.Date(2025, 1, 19, 0, 0, 0, 0, time.UTC),
		},

		// Absolute dates
		{
			name:       "absolute date today",
			input:      "2025-01-10",
			relativeTo: friday,
			want:       time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "absolute date future",
			input:      "2025-01-15",
			relativeTo: friday,
			want:       time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "absolute date weekend",
			input:      "2025-01-11",
			relativeTo: friday,
			want:       time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "absolute date far future",
			input:      "2030-12-31",
			relativeTo: friday,
			want:       time.Date(2030, 12, 31, 0, 0, 0, 0, time.UTC),
		},

		// Edge case: whitespace
		{
			name:       "input with whitespace",
			input:      "  monday  ",
			relativeTo: friday,
			want:       time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRelativeDate(tt.input, tt.relativeTo)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("got error %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseRelativeDate_Errors(t *testing.T) {
	friday := time.Date(2025, 1, 10, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:    "past absolute date",
			input:   "2025-01-09",
			wantErr: ErrDateInPast,
		},
		{
			name:    "past absolute date far",
			input:   "2020-01-01",
			wantErr: ErrDateInPast,
		},
		{
			name:    "invalid format US style",
			input:   "01-10-2025",
			wantErr: ErrInvalidDateFormat,
		},
		{
			name:    "invalid format slash",
			input:   "10/01/2025",
			wantErr: ErrInvalidDateFormat,
		},
		{
			name:    "typo weekday",
			input:   "mondya",
			wantErr: ErrInvalidDateFormat,
		},
		{
			name:    "typo next-weekday",
			input:   "next-mondya",
			wantErr: ErrInvalidDateFormat,
		},
		{
			name:    "invalid keyword",
			input:   "yesterday",
			wantErr: ErrInvalidDateFormat,
		},
		{
			name:    "random text",
			input:   "foo",
			wantErr: ErrInvalidDateFormat,
		},
		{
			name:    "next- without weekday",
			input:   "next-",
			wantErr: ErrInvalidDateFormat,
		},
		{
			name:    "next-invalid",
			input:   "next-foo",
			wantErr: ErrInvalidDateFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseRelativeDate(tt.input, friday)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
		})
	}
}

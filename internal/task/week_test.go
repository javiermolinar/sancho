package task

import (
	"testing"
	"time"
)

func TestNewWeek(t *testing.T) {
	// Wednesday, January 15, 2025
	date := time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)
	week := NewWeek(date)

	// Monday should be January 13, 2025
	expectedMonday := time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC)
	if !week.StartDate.Equal(expectedMonday) {
		t.Errorf("expected StartDate %v, got %v", expectedMonday, week.StartDate)
	}

	// Check all 7 days are initialized
	for i := 0; i < 7; i++ {
		if week.Days[i] == nil {
			t.Errorf("day %d is nil", i)
			continue
		}
		expectedDate := expectedMonday.AddDate(0, 0, i)
		if !week.Days[i].Date.Equal(expectedDate) {
			t.Errorf("day %d: expected date %v, got %v", i, expectedDate, week.Days[i].Date)
		}
	}
}

func TestNewWeek_Sunday(t *testing.T) {
	// Sunday, January 19, 2025
	date := time.Date(2025, 1, 19, 0, 0, 0, 0, time.UTC)
	week := NewWeek(date)

	// Monday should still be January 13, 2025
	expectedMonday := time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC)
	if !week.StartDate.Equal(expectedMonday) {
		t.Errorf("expected StartDate %v, got %v", expectedMonday, week.StartDate)
	}
}

func TestWeek_EndDate(t *testing.T) {
	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	week := NewWeek(date)

	expectedSunday := time.Date(2025, 1, 19, 0, 0, 0, 0, time.UTC)
	if !week.EndDate().Equal(expectedSunday) {
		t.Errorf("expected EndDate %v, got %v", expectedSunday, week.EndDate())
	}
}

func TestWeek_Day(t *testing.T) {
	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	week := NewWeek(date)

	tests := []struct {
		weekday     int
		expectNil   bool
		expectedDay int // day of month
	}{
		{weekday: 0, expectNil: false, expectedDay: 13}, // Monday
		{weekday: 2, expectNil: false, expectedDay: 15}, // Wednesday
		{weekday: 6, expectNil: false, expectedDay: 19}, // Sunday
		{weekday: -1, expectNil: true},
		{weekday: 7, expectNil: true},
	}

	for _, tt := range tests {
		day := week.Day(tt.weekday)
		if tt.expectNil {
			if day != nil {
				t.Errorf("Day(%d) expected nil, got %v", tt.weekday, day)
			}
		} else {
			if day == nil {
				t.Errorf("Day(%d) expected non-nil", tt.weekday)
				continue
			}
			if day.Date.Day() != tt.expectedDay {
				t.Errorf("Day(%d) expected day %d, got %d", tt.weekday, tt.expectedDay, day.Date.Day())
			}
		}
	}
}

func TestWeek_DayByDate(t *testing.T) {
	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	week := NewWeek(date)

	t.Run("date in week", func(t *testing.T) {
		wednesday := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
		day := week.DayByDate(wednesday)
		if day == nil {
			t.Fatal("expected non-nil day")
		}
		if day.Date.Day() != 15 {
			t.Errorf("expected day 15, got %d", day.Date.Day())
		}
	})

	t.Run("date not in week", func(t *testing.T) {
		nextWeek := time.Date(2025, 1, 22, 0, 0, 0, 0, time.UTC)
		day := week.DayByDate(nextWeek)
		if day != nil {
			t.Errorf("expected nil, got %v", day)
		}
	})
}

func TestNewWeekFromTasks(t *testing.T) {
	monday := time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC)
	wednesday := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	nextWeek := time.Date(2025, 1, 22, 0, 0, 0, 0, time.UTC)

	tasks := []*Task{
		{Description: "Monday task", ScheduledDate: monday, ScheduledStart: "09:00", ScheduledEnd: "10:00", Status: StatusScheduled},
		{Description: "Wednesday task", ScheduledDate: wednesday, ScheduledStart: "11:00", ScheduledEnd: "12:00", Status: StatusScheduled},
		{Description: "Next week task", ScheduledDate: nextWeek, ScheduledStart: "09:00", ScheduledEnd: "10:00", Status: StatusScheduled},
	}

	week := NewWeekFromTasks(wednesday, tasks)

	// Monday should have 1 task
	if week.Days[0].Len() != 1 {
		t.Errorf("expected 1 task on Monday, got %d", week.Days[0].Len())
	}

	// Wednesday should have 1 task
	if week.Days[2].Len() != 1 {
		t.Errorf("expected 1 task on Wednesday, got %d", week.Days[2].Len())
	}

	// Other days should be empty
	if week.Days[1].Len() != 0 {
		t.Errorf("expected 0 tasks on Tuesday, got %d", week.Days[1].Len())
	}
}

func TestWeek_AllTasks(t *testing.T) {
	monday := time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC)
	wednesday := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	tasks := []*Task{
		{Description: "Monday task", ScheduledDate: monday, ScheduledStart: "09:00", ScheduledEnd: "10:00", Status: StatusScheduled},
		{Description: "Wednesday task", ScheduledDate: wednesday, ScheduledStart: "11:00", ScheduledEnd: "12:00", Status: StatusScheduled},
	}

	week := NewWeekFromTasks(wednesday, tasks)
	allTasks := week.AllTasks()

	if len(allTasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(allTasks))
	}
}

func TestWeek_Stats(t *testing.T) {
	monday := time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC)
	wednesday := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	tasks := []*Task{
		{Description: "Monday deep", ScheduledDate: monday, ScheduledStart: "09:00", ScheduledEnd: "11:00", Status: StatusScheduled, Category: CategoryDeep},
		{Description: "Monday shallow", ScheduledDate: monday, ScheduledStart: "11:00", ScheduledEnd: "12:00", Status: StatusScheduled, Category: CategoryShallow},
		{Description: "Wednesday deep", ScheduledDate: wednesday, ScheduledStart: "09:00", ScheduledEnd: "10:00", Status: StatusScheduled, Category: CategoryDeep},
		{Description: "Wednesday cancelled", ScheduledDate: wednesday, ScheduledStart: "14:00", ScheduledEnd: "15:00", Status: StatusCancelled, Category: CategoryDeep},
	}

	week := NewWeekFromTasks(wednesday, tasks)
	stats := week.Stats()

	if stats.TotalBlocks != 4 {
		t.Errorf("expected 4 total blocks, got %d", stats.TotalBlocks)
	}
	if stats.DeepMinutes != 180 { // 120 + 60
		t.Errorf("expected 180 deep minutes, got %d", stats.DeepMinutes)
	}
	if stats.ShallowMinutes != 60 {
		t.Errorf("expected 60 shallow minutes, got %d", stats.ShallowMinutes)
	}
	if stats.CancelledBlocks != 1 {
		t.Errorf("expected 1 cancelled block, got %d", stats.CancelledBlocks)
	}
}

func TestWeekStats_Ratio(t *testing.T) {
	tests := []struct {
		name    string
		deep    int
		shallow int
		want    string
	}{
		{name: "2:1 ratio", deep: 120, shallow: 60, want: "2.0:1"},
		{name: "all deep", deep: 120, shallow: 0, want: "∞:1"},
		{name: "no work", deep: 0, shallow: 0, want: "0:0"},
		{name: "1.5:1 ratio", deep: 90, shallow: 60, want: "1.5:1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := WeekStats{DeepMinutes: tt.deep, ShallowMinutes: tt.shallow}
			if got := stats.Ratio(); got != tt.want {
				t.Errorf("Ratio() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWeekStats_BestDay(t *testing.T) {
	stats := WeekStats{
		DayStats: [7]DayStats{
			{DeepMinutes: 60},  // Monday
			{DeepMinutes: 120}, // Tuesday - best
			{DeepMinutes: 90},  // Wednesday
			{DeepMinutes: 0},   // Thursday
			{DeepMinutes: 30},  // Friday
			{DeepMinutes: 0},   // Saturday
			{DeepMinutes: 0},   // Sunday
		},
	}

	weekday, minutes := stats.BestDay()
	if weekday != 1 {
		t.Errorf("expected best day 1 (Tuesday), got %d", weekday)
	}
	if minutes != 120 {
		t.Errorf("expected 120 minutes, got %d", minutes)
	}
}

func TestWeekStats_BestDay_NoWork(t *testing.T) {
	stats := WeekStats{}

	weekday, minutes := stats.BestDay()
	if weekday != -1 {
		t.Errorf("expected -1 for no work, got %d", weekday)
	}
	if minutes != 0 {
		t.Errorf("expected 0 minutes, got %d", minutes)
	}
}

func TestWeek_StatsWithPeakHours(t *testing.T) {
	monday := time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC)

	// Deep work from 09:00-11:00, peak hours 09:00-10:00
	tasks := []*Task{
		{Description: "Deep work", ScheduledDate: monday, ScheduledStart: "09:00", ScheduledEnd: "11:00", Status: StatusScheduled, Category: CategoryDeep},
	}

	week := NewWeekFromTasks(monday, tasks)
	stats := week.StatsWithPeakHours("09:00", "10:00")

	if stats.DeepMinutes != 120 {
		t.Errorf("expected 120 deep minutes, got %d", stats.DeepMinutes)
	}
	if stats.PeakDeepMinutes != 60 {
		t.Errorf("expected 60 peak deep minutes, got %d", stats.PeakDeepMinutes)
	}
	if stats.PeakPercent() != 50 {
		t.Errorf("expected 50%% peak, got %d%%", stats.PeakPercent())
	}
}

func TestWeekdayName(t *testing.T) {
	tests := []struct {
		weekday int
		want    string
	}{
		{0, "Monday"},
		{1, "Tuesday"},
		{2, "Wednesday"},
		{3, "Thursday"},
		{4, "Friday"},
		{5, "Saturday"},
		{6, "Sunday"},
		{-1, ""},
		{7, ""},
	}

	for _, tt := range tests {
		if got := WeekdayName(tt.weekday); got != tt.want {
			t.Errorf("WeekdayName(%d) = %q, want %q", tt.weekday, got, tt.want)
		}
	}
}

func TestWeekdayShortName(t *testing.T) {
	tests := []struct {
		weekday int
		want    string
	}{
		{0, "Mon"},
		{4, "Fri"},
		{6, "Sun"},
		{-1, ""},
		{7, ""},
	}

	for _, tt := range tests {
		if got := WeekdayShortName(tt.weekday); got != tt.want {
			t.Errorf("WeekdayShortName(%d) = %q, want %q", tt.weekday, got, tt.want)
		}
	}
}

func TestStartOfWeek(t *testing.T) {
	tests := []struct {
		name     string
		date     time.Time
		expected time.Time
	}{
		{
			name:     "Wednesday",
			date:     time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
			expected: time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "Monday",
			date:     time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "Sunday",
			date:     time.Date(2025, 1, 19, 23, 59, 59, 0, time.UTC),
			expected: time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := startOfWeek(tt.date)
			if !got.Equal(tt.expected) {
				t.Errorf("startOfWeek(%v) = %v, want %v", tt.date, got, tt.expected)
			}
		})
	}
}

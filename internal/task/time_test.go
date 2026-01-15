package task

import "testing"

func TestTimeToMinutes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{name: "midnight", input: "00:00", want: 0},
		{name: "9am", input: "09:00", want: 540},
		{name: "noon", input: "12:00", want: 720},
		{name: "5pm", input: "17:00", want: 1020},
		{name: "11:59pm", input: "23:59", want: 1439},
		{name: "with minutes", input: "09:30", want: 570},
		{name: "invalid short", input: "9:00", want: 0},
		{name: "empty", input: "", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TimeToMinutes(tt.input)
			if got != tt.want {
				t.Errorf("TimeToMinutes(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestMinutesToTime(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  string
	}{
		{name: "midnight", input: 0, want: "00:00"},
		{name: "9am", input: 540, want: "09:00"},
		{name: "noon", input: 720, want: "12:00"},
		{name: "5pm", input: 1020, want: "17:00"},
		{name: "11:59pm", input: 1439, want: "23:59"},
		{name: "with minutes", input: 570, want: "09:30"},
		{name: "negative clamps to zero", input: -10, want: "00:00"},
		{name: "over 24h clamps", input: 1500, want: "23:59"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MinutesToTime(tt.input)
			if got != tt.want {
				t.Errorf("MinutesToTime(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestOverlapMinutes(t *testing.T) {
	tests := []struct {
		name                       string
		start1, end1, start2, end2 string
		want                       int
	}{
		{
			name:   "no overlap - before",
			start1: "09:00", end1: "10:00",
			start2: "10:00", end2: "11:00",
			want: 0,
		},
		{
			name:   "no overlap - after",
			start1: "11:00", end1: "12:00",
			start2: "09:00", end2: "10:00",
			want: 0,
		},
		{
			name:   "partial overlap - end overlaps start",
			start1: "09:00", end1: "10:30",
			start2: "10:00", end2: "11:00",
			want: 30,
		},
		{
			name:   "partial overlap - start overlaps end",
			start1: "10:00", end1: "11:00",
			start2: "09:00", end2: "10:30",
			want: 30,
		},
		{
			name:   "full overlap - same range",
			start1: "09:00", end1: "11:00",
			start2: "09:00", end2: "11:00",
			want: 120,
		},
		{
			name:   "full overlap - one inside other",
			start1: "09:00", end1: "12:00",
			start2: "10:00", end2: "11:00",
			want: 60,
		},
		{
			name:   "full overlap - other inside one",
			start1: "10:00", end1: "11:00",
			start2: "09:00", end2: "12:00",
			want: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := OverlapMinutes(tt.start1, tt.end1, tt.start2, tt.end2)
			if got != tt.want {
				t.Errorf("OverlapMinutes(%s-%s, %s-%s) = %d, want %d",
					tt.start1, tt.end1, tt.start2, tt.end2, got, tt.want)
			}
		})
	}
}

func TestTimesOverlap(t *testing.T) {
	tests := []struct {
		name                       string
		start1, end1, start2, end2 string
		want                       bool
	}{
		{
			name:   "no overlap - adjacent",
			start1: "09:00", end1: "10:00",
			start2: "10:00", end2: "11:00",
			want: false,
		},
		{
			name:   "no overlap - gap between",
			start1: "09:00", end1: "10:00",
			start2: "11:00", end2: "12:00",
			want: false,
		},
		{
			name:   "overlap - partial",
			start1: "09:00", end1: "10:30",
			start2: "10:00", end2: "11:00",
			want: true,
		},
		{
			name:   "overlap - same range",
			start1: "09:00", end1: "11:00",
			start2: "09:00", end2: "11:00",
			want: true,
		},
		{
			name:   "overlap - one inside other",
			start1: "09:00", end1: "12:00",
			start2: "10:00", end2: "11:00",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TimesOverlap(tt.start1, tt.end1, tt.start2, tt.end2)
			if got != tt.want {
				t.Errorf("TimesOverlap(%s-%s, %s-%s) = %v, want %v",
					tt.start1, tt.end1, tt.start2, tt.end2, got, tt.want)
			}
		})
	}
}

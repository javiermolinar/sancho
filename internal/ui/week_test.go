package ui

import "testing"

func TestOverlapMinutes(t *testing.T) {
	tests := []struct {
		name   string
		start1 string
		end1   string
		start2 string
		end2   string
		want   int
	}{
		{
			name:   "no overlap - task before peak",
			start1: "07:00",
			end1:   "08:00",
			start2: "09:00",
			end2:   "12:00",
			want:   0,
		},
		{
			name:   "no overlap - task after peak",
			start1: "13:00",
			end1:   "15:00",
			start2: "09:00",
			end2:   "12:00",
			want:   0,
		},
		{
			name:   "task entirely within peak",
			start1: "09:30",
			end1:   "11:00",
			start2: "09:00",
			end2:   "12:00",
			want:   90,
		},
		{
			name:   "peak entirely within task",
			start1: "08:00",
			end1:   "14:00",
			start2: "09:00",
			end2:   "12:00",
			want:   180,
		},
		{
			name:   "task starts before peak, ends during",
			start1: "08:00",
			end1:   "10:00",
			start2: "09:00",
			end2:   "12:00",
			want:   60,
		},
		{
			name:   "task starts during peak, ends after",
			start1: "11:00",
			end1:   "14:00",
			start2: "09:00",
			end2:   "12:00",
			want:   60,
		},
		{
			name:   "exact match",
			start1: "09:00",
			end1:   "12:00",
			start2: "09:00",
			end2:   "12:00",
			want:   180,
		},
		{
			name:   "adjacent - task ends when peak starts",
			start1: "08:00",
			end1:   "09:00",
			start2: "09:00",
			end2:   "12:00",
			want:   0,
		},
		{
			name:   "adjacent - task starts when peak ends",
			start1: "12:00",
			end1:   "14:00",
			start2: "09:00",
			end2:   "12:00",
			want:   0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := OverlapMinutes(tc.start1, tc.end1, tc.start2, tc.end2)
			if got != tc.want {
				t.Errorf("overlapMinutes(%s-%s, %s-%s) = %d, want %d",
					tc.start1, tc.end1, tc.start2, tc.end2, got, tc.want)
			}
		})
	}
}

func TestTimeToMinutes(t *testing.T) {
	tests := []struct {
		time string
		want int
	}{
		{"00:00", 0},
		{"00:01", 1},
		{"01:00", 60},
		{"09:00", 540},
		{"12:30", 750},
		{"17:00", 1020},
		{"23:59", 1439},
	}

	for _, tc := range tests {
		t.Run(tc.time, func(t *testing.T) {
			got := TimeToMinutes(tc.time)
			if got != tc.want {
				t.Errorf("timeToMinutes(%s) = %d, want %d", tc.time, got, tc.want)
			}
		})
	}
}

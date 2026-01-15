package task

import "fmt"

// TimeToMinutes converts "HH:MM" to minutes since midnight.
// Returns 0 for invalid input.
func TimeToMinutes(t string) int {
	if len(t) < 5 {
		return 0
	}
	hours := int(t[0]-'0')*10 + int(t[1]-'0')
	mins := int(t[3]-'0')*10 + int(t[4]-'0')
	return hours*60 + mins
}

// MinutesToTime converts minutes since midnight to "HH:MM" format.
func MinutesToTime(m int) string {
	if m < 0 {
		m = 0
	}
	if m >= 24*60 {
		m = 24*60 - 1
	}
	return fmt.Sprintf("%02d:%02d", m/60, m%60)
}

// OverlapMinutes calculates the overlapping minutes between two time ranges.
// All times are in "HH:MM" format.
// Returns 0 if there is no overlap.
func OverlapMinutes(start1, end1, start2, end2 string) int {
	s1 := TimeToMinutes(start1)
	e1 := TimeToMinutes(end1)
	s2 := TimeToMinutes(start2)
	e2 := TimeToMinutes(end2)

	overlapStart := max(s1, s2)
	overlapEnd := min(e1, e2)

	if overlapEnd <= overlapStart {
		return 0
	}
	return overlapEnd - overlapStart
}

// TimesOverlap returns true if two time ranges overlap.
// Two time ranges overlap if: start1 < end2 AND start2 < end1
func TimesOverlap(start1, end1, start2, end2 string) bool {
	return start1 < end2 && start2 < end1
}

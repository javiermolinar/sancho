package task

import (
	"testing"
	"time"
)

func TestNewWeekWindow(t *testing.T) {
	prev := NewWeek(time.Date(2025, 1, 6, 0, 0, 0, 0, time.Local))  // Jan 6 (Mon)
	curr := NewWeek(time.Date(2025, 1, 13, 0, 0, 0, 0, time.Local)) // Jan 13 (Mon)
	next := NewWeek(time.Date(2025, 1, 20, 0, 0, 0, 0, time.Local)) // Jan 20 (Mon)

	w := NewWeekWindow(prev, curr, next)

	if w.Previous() != prev {
		t.Error("Previous() should return the first week")
	}
	if w.Current() != curr {
		t.Error("Current() should return the center week")
	}
	if w.Next() != next {
		t.Error("Next() should return the last week")
	}
}

func TestWeekWindow_ShiftForward(t *testing.T) {
	week1 := NewWeek(time.Date(2025, 1, 6, 0, 0, 0, 0, time.Local))
	week2 := NewWeek(time.Date(2025, 1, 13, 0, 0, 0, 0, time.Local))
	week3 := NewWeek(time.Date(2025, 1, 20, 0, 0, 0, 0, time.Local))
	week4 := NewWeek(time.Date(2025, 1, 27, 0, 0, 0, 0, time.Local))

	w := NewWeekWindow(week1, week2, week3)

	// Shift forward: [week1, week2, week3] -> [week2, week3, week4]
	w.ShiftForward(week4)

	if w.Previous() != week2 {
		t.Error("after ShiftForward, Previous() should be old current (week2)")
	}
	if w.Current() != week3 {
		t.Error("after ShiftForward, Current() should be old next (week3)")
	}
	if w.Next() != week4 {
		t.Error("after ShiftForward, Next() should be new week (week4)")
	}
}

func TestWeekWindow_ShiftBackward(t *testing.T) {
	week1 := NewWeek(time.Date(2025, 1, 6, 0, 0, 0, 0, time.Local))
	week2 := NewWeek(time.Date(2025, 1, 13, 0, 0, 0, 0, time.Local))
	week3 := NewWeek(time.Date(2025, 1, 20, 0, 0, 0, 0, time.Local))
	week0 := NewWeek(time.Date(2024, 12, 30, 0, 0, 0, 0, time.Local))

	w := NewWeekWindow(week1, week2, week3)

	// Shift backward: [week1, week2, week3] -> [week0, week1, week2]
	w.ShiftBackward(week0)

	if w.Previous() != week0 {
		t.Error("after ShiftBackward, Previous() should be new week (week0)")
	}
	if w.Current() != week1 {
		t.Error("after ShiftBackward, Current() should be old prev (week1)")
	}
	if w.Next() != week2 {
		t.Error("after ShiftBackward, Next() should be old current (week2)")
	}
}

func TestWeekWindow_SetCurrent(t *testing.T) {
	week1 := NewWeek(time.Date(2025, 1, 6, 0, 0, 0, 0, time.Local))
	week2 := NewWeek(time.Date(2025, 1, 13, 0, 0, 0, 0, time.Local))
	week3 := NewWeek(time.Date(2025, 1, 20, 0, 0, 0, 0, time.Local))

	w := NewWeekWindow(week1, week2, week3)

	// Create a new week with the same dates (simulates reload)
	reloaded := NewWeek(time.Date(2025, 1, 13, 0, 0, 0, 0, time.Local))
	w.SetCurrent(reloaded)

	if w.Current() != reloaded {
		t.Error("SetCurrent should replace the current week")
	}
	// Previous and next should be unchanged
	if w.Previous() != week1 {
		t.Error("SetCurrent should not affect Previous")
	}
	if w.Next() != week3 {
		t.Error("SetCurrent should not affect Next")
	}
}

func TestWeekWindow_SetNextAndPrevious(t *testing.T) {
	week1 := NewWeek(time.Date(2025, 1, 6, 0, 0, 0, 0, time.Local))
	week2 := NewWeek(time.Date(2025, 1, 13, 0, 0, 0, 0, time.Local))
	week3 := NewWeek(time.Date(2025, 1, 20, 0, 0, 0, 0, time.Local))

	w := NewWeekWindow(nil, week2, nil) // Start with only current loaded

	w.SetPrevious(week1)
	w.SetNext(week3)

	if w.Previous() != week1 {
		t.Error("SetPrevious should set the previous week")
	}
	if w.Next() != week3 {
		t.Error("SetNext should set the next week")
	}
}

func TestWeekWindow_HasNextAndPrevious(t *testing.T) {
	week2 := NewWeek(time.Date(2025, 1, 13, 0, 0, 0, 0, time.Local))

	t.Run("all nil except current", func(t *testing.T) {
		w := NewWeekWindow(nil, week2, nil)
		if w.HasPrevious() {
			t.Error("HasPrevious should return false when previous is nil")
		}
		if w.HasNext() {
			t.Error("HasNext should return false when next is nil")
		}
	})

	t.Run("all populated", func(t *testing.T) {
		week1 := NewWeek(time.Date(2025, 1, 6, 0, 0, 0, 0, time.Local))
		week3 := NewWeek(time.Date(2025, 1, 20, 0, 0, 0, 0, time.Local))
		w := NewWeekWindow(week1, week2, week3)
		if !w.HasPrevious() {
			t.Error("HasPrevious should return true when previous is set")
		}
		if !w.HasNext() {
			t.Error("HasNext should return true when next is set")
		}
	})
}

func TestWeekWindow_MultipleShifts(t *testing.T) {
	// Test navigating forward multiple times
	week1 := NewWeek(time.Date(2025, 1, 6, 0, 0, 0, 0, time.Local))
	week2 := NewWeek(time.Date(2025, 1, 13, 0, 0, 0, 0, time.Local))
	week3 := NewWeek(time.Date(2025, 1, 20, 0, 0, 0, 0, time.Local))
	week4 := NewWeek(time.Date(2025, 1, 27, 0, 0, 0, 0, time.Local))
	week5 := NewWeek(time.Date(2025, 2, 3, 0, 0, 0, 0, time.Local))

	w := NewWeekWindow(week1, week2, week3)

	// Shift forward twice
	w.ShiftForward(week4)
	w.ShiftForward(week5)

	// Should now be [week3, week4, week5]
	if w.Previous() != week3 {
		t.Errorf("after 2 ShiftForward, Previous() should be week3, got %v", w.Previous().StartDate)
	}
	if w.Current() != week4 {
		t.Errorf("after 2 ShiftForward, Current() should be week4, got %v", w.Current().StartDate)
	}
	if w.Next() != week5 {
		t.Errorf("after 2 ShiftForward, Next() should be week5, got %v", w.Next().StartDate)
	}
}

func TestWeekWindow_ShiftForwardThenBackward(t *testing.T) {
	week1 := NewWeek(time.Date(2025, 1, 6, 0, 0, 0, 0, time.Local))
	week2 := NewWeek(time.Date(2025, 1, 13, 0, 0, 0, 0, time.Local))
	week3 := NewWeek(time.Date(2025, 1, 20, 0, 0, 0, 0, time.Local))
	week4 := NewWeek(time.Date(2025, 1, 27, 0, 0, 0, 0, time.Local))

	w := NewWeekWindow(week1, week2, week3)

	// Shift forward: [week1, week2, week3] -> [week2, week3, week4]
	w.ShiftForward(week4)

	// Shift backward with week1: [week2, week3, week4] -> [week1, week2, week3]
	w.ShiftBackward(week1)

	// Should be back to [week1, week2, week3]
	if w.Previous() != week1 {
		t.Error("after forward+backward, Previous() should be week1")
	}
	if w.Current() != week2 {
		t.Error("after forward+backward, Current() should be week2")
	}
	if w.Next() != week3 {
		t.Error("after forward+backward, Next() should be week3")
	}
}

// Package tui provides the terminal user interface for sancho.
package tui

import "testing"

func TestVisibleSlotsForTable(t *testing.T) {
	tests := []struct {
		name     string
		height   int
		rowLines int
		want     int
	}{
		{name: "height_uses_table_chrome", height: 19, rowLines: 1, want: 15},
		{name: "multi_line_rows", height: 20, rowLines: 2, want: 8},
		{name: "too_small", height: 4, rowLines: 1, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{rowLines: tt.rowLines}
			if got := m.visibleSlotsForTable(tt.height); got != tt.want {
				t.Fatalf("visibleSlotsForTable(%d) = %d, want %d", tt.height, got, tt.want)
			}
		})
	}
}

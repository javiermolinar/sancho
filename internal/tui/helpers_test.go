// Package tui provides the terminal user interface for sancho.
package tui

import (
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/textinput"

	"github.com/javiermolinar/sancho/internal/config"
	"github.com/javiermolinar/sancho/internal/task"
	"github.com/javiermolinar/sancho/internal/tui/theme"
)

func TestGetFooterHeight(t *testing.T) {
	palette := &theme.Theme{
		Bg:          "#101010",
		BgHighlight: "#202020",
		BgSelection: "#303030",
		Fg:          "#ffffff",
		FgMuted:     "#aaaaaa",
		Accent:      "#ff0000",
		Deep:        "#00ff00",
		Shallow:     "#0000ff",
		Current:     "#ffff00",
		Warning:     "#ff00ff",
	}
	styles := NewStyles(palette)

	tests := []struct {
		name       string
		width      int
		height     int
		wantFooter int
	}{
		{
			name:       "full layout",
			width:      100,
			height:     40,
			wantFooter: footerMinHeight,
		},
		{
			name:       "compact layout",
			width:      80,
			height:     10,
			wantFooter: footerCompact,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				width:  tt.width,
				height: tt.height,
				styles: styles,
				prompt: textinput.New(),
			}
			gotFooter := m.getFooterHeight()
			if gotFooter != tt.wantFooter {
				t.Fatalf("footer height = %d, want %d", gotFooter, tt.wantFooter)
			}
		})
	}
}

func TestEnsureCursorVisibleUsesGridHeight(t *testing.T) {
	cfg := config.Default()
	slotConfig := SlotGridConfigFromWeekWindow(nil, cfg.Schedule.DayStart, cfg.Schedule.DayEnd, time.Now, 15)
	m := Model{
		config:    cfg,
		slotState: NewSlotStateManager(slotConfig),
		rowHeight: 15,
		rowLines:  1,
		layoutCache: LayoutCache{
			GridH: 15,
		},
	}

	m.cursor = Position{Day: 0, Slot: 12}
	m.ensureCursorVisible()

	if m.scrollOffset != 2 {
		t.Fatalf("scrollOffset = %d, want %d", m.scrollOffset, 2)
	}
}

func TestFocusCursorOnTaskEndAfterShrink(t *testing.T) {
	cfg := config.Default()
	nowFunc := func() time.Time {
		return time.Date(2030, 1, 1, 8, 0, 0, 0, time.UTC)
	}
	slotConfig := SlotGridConfigFromWeekWindow(nil, cfg.Schedule.DayStart, cfg.Schedule.DayEnd, nowFunc, 15)
	grid := NewSlotGrid(slotConfig)
	taskA := &task.Task{
		ID:          1,
		Description: "Task A",
		Category:    task.CategoryDeep,
		Status:      task.StatusScheduled,
	}
	grid, err := grid.Place(taskA, 7, 0, 2)
	if err != nil {
		t.Fatalf("place task failed: %v", err)
	}

	sm := NewSlotStateManager(slotConfig)
	sm.SetGrid(grid)
	sm.EnterEditMode()

	m := Model{
		config:    cfg,
		slotState: sm,
		rowHeight: 15,
		cursor:    Position{Day: 0, Slot: 1},
	}

	if err := m.slotState.Shrink(taskA); err != nil {
		t.Fatalf("shrink failed: %v", err)
	}
	m.focusCursorOnTaskEnd(taskA)

	if m.cursor.Day != 0 {
		t.Fatalf("cursor day = %d, want 0", m.cursor.Day)
	}
	if m.cursor.Slot != 0 {
		t.Fatalf("cursor slot = %d, want 0", m.cursor.Slot)
	}
}

// Package tui provides the terminal user interface for sancho.
package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/javiermolinar/sancho/internal/config"
	"github.com/javiermolinar/sancho/internal/task"
	"github.com/javiermolinar/sancho/internal/tui/view"
)

func TestRenderModalButtons_UsesModalBodySeparator(t *testing.T) {
	cfg := &config.Config{
		Schedule: config.ScheduleConfig{
			DayStart: "09:00",
			DayEnd:   "17:00",
		},
	}

	m := New(nil, cfg)
	buttons := view.RenderModalButtons(m.modalStyles(), "[Enter] Save", "[Esc] Cancel")
	sep := m.styles.ModalBodyStyle.Render(" ")
	if !strings.Contains(buttons, sep) {
		t.Errorf("expected modal button separator to use modal body style")
	}
}

func TestRenderTaskDetailModal_UsesModalBodyStyleForDescription(t *testing.T) {
	cfg := &config.Config{
		Schedule: config.ScheduleConfig{
			DayStart: "09:00",
			DayEnd:   "17:00",
		},
	}

	m := New(nil, cfg)
	m.modalTask = &task.Task{
		ID:             1,
		Description:    "Prepare release notes",
		Category:       task.CategoryDeep,
		ScheduledDate:  time.Date(2025, 1, 6, 0, 0, 0, 0, time.Local),
		ScheduledStart: "10:00",
		ScheduledEnd:   "10:30",
		Status:         task.StatusScheduled,
	}

	view := m.renderTaskDetailModal()
	expected := m.styles.ModalBodyStyle.Render("Prepare release notes")
	if !strings.Contains(view, expected) {
		t.Errorf("expected modal description to use modal body style")
	}
}

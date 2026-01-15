// Package tui provides the terminal user interface for sancho.
package tui

import (
	"testing"

	"github.com/javiermolinar/sancho/internal/config"
)

func TestNewModel_AppliesModalInputStyles(t *testing.T) {
	cfg := &config.Config{
		Schedule: config.ScheduleConfig{
			DayStart: "09:00",
			DayEnd:   "17:00",
		},
	}

	m := New(nil, cfg)
	if got, want := m.formDesc.TextStyle.Render("x"), m.styles.ModalInputTextStyle.Render("x"); got != want {
		t.Errorf("TextStyle mismatch: got %q, want %q", got, want)
	}
	if got, want := m.formDesc.PromptStyle.Render("x"), m.styles.ModalInputTextStyle.Render("x"); got != want {
		t.Errorf("PromptStyle mismatch: got %q, want %q", got, want)
	}
	if got, want := m.formDesc.Cursor.Style.Render("x"), m.styles.ModalInputCursorStyle.Render("x"); got != want {
		t.Errorf("Cursor style mismatch: got %q, want %q", got, want)
	}
	if got, want := m.formDesc.Cursor.TextStyle.Render("x"), m.styles.ModalInputTextStyle.Render("x"); got != want {
		t.Errorf("Cursor text style mismatch: got %q, want %q", got, want)
	}
}

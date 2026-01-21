// Package tui provides the terminal user interface for sancho.
package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleTaskFormKeys_AllowsTypingH(t *testing.T) {
	input := textinput.New()
	input.Focus()

	m := Model{
		mode:         ModeModal,
		modalType:    ModalTaskForm,
		formFocus:    0,
		formDesc:     input,
		formDuration: 0,
	}

	updated, _ := m.handleTaskFormKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	model := updated.(Model)

	if got := model.formDesc.Value(); got != "h" {
		t.Fatalf("value = %q, want %q", got, "h")
	}
}

func TestHandleTaskFormKeys_UpDownMovesFocus(t *testing.T) {
	input := textinput.New()
	input.Focus()

	m := Model{
		mode:      ModeModal,
		modalType: ModalTaskForm,
		formFocus: 0,
		formDesc:  input,
	}

	updated, _ := m.handleTaskFormKeys(tea.KeyMsg{Type: tea.KeyDown})
	model := updated.(Model)

	if model.formFocus != 1 {
		t.Fatalf("formFocus = %d, want %d", model.formFocus, 1)
	}
	if model.formDesc.Focused() {
		t.Fatal("expected description to be blurred after moving down")
	}

	updated, _ = model.handleTaskFormKeys(tea.KeyMsg{Type: tea.KeyUp})
	model = updated.(Model)

	if model.formFocus != 0 {
		t.Fatalf("formFocus = %d, want %d", model.formFocus, 0)
	}
	if !model.formDesc.Focused() {
		t.Fatal("expected description to be focused after moving up")
	}
}

func TestSaveTaskFromForm_EmptyDescriptionFocusesName(t *testing.T) {
	input := textinput.New()
	input.SetValue("   ")

	m := Model{
		mode:      ModeModal,
		modalType: ModalTaskForm,
		formFocus: 2,
		formDesc:  input,
	}

	updated, _ := m.saveTaskFromForm()
	model := updated.(Model)

	if model.statusMsg != "Description is required" {
		t.Fatalf("statusMsg = %q, want %q", model.statusMsg, "Description is required")
	}
	if model.formFocus != 0 {
		t.Fatalf("formFocus = %d, want %d", model.formFocus, 0)
	}
	if !model.formDesc.Focused() {
		t.Fatal("expected description to be focused after validation error")
	}
	if model.formDesc.Value() != "" {
		t.Fatalf("description value = %q, want %q", model.formDesc.Value(), "")
	}
}

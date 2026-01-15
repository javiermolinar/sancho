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

package view

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderTableIncludesHeader(t *testing.T) {
	state := TableViewState{
		InnerW:       20,
		GridH:        5,
		Headers:      []string{"Hdr"},
		HeaderStyles: []lipgloss.Style{lipgloss.NewStyle()},
		Content: TableContent{
			Rows:       [][]string{{"Cell"}},
			CellStyles: [][]lipgloss.Style{{lipgloss.NewStyle()}},
		},
		BorderStyle: lipgloss.NewStyle(),
		VAlign:      lipgloss.Top,
		Bg:          lipgloss.Color(""),
		Render:      true,
	}

	out := RenderTable(state)
	if !strings.Contains(out, "Hdr") {
		t.Fatalf("expected header in output: %q", out)
	}
}

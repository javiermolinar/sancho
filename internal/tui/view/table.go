package view

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// TableContent contains table rows and cell styles.
type TableContent struct {
	Rows       [][]string
	CellStyles [][]lipgloss.Style
}

// TableViewState holds data needed to render the task grid.
type TableViewState struct {
	InnerW       int
	GridH        int
	Headers      []string
	HeaderStyles []lipgloss.Style
	Content      TableContent
	BorderStyle  lipgloss.Style
	VAlign       lipgloss.Position
	Bg           lipgloss.Color
	Render       bool
}

// RenderTable renders the scrollable task grid using a lipgloss table.
func RenderTable(state TableViewState) string {
	if !state.Render || state.GridH <= 0 {
		return ""
	}

	tableWidth := state.InnerW - 2
	if tableWidth < 0 {
		tableWidth = 0
	}
	tableHeight := state.GridH
	if tableHeight < 0 {
		tableHeight = 0
	}

	t := table.New().
		Headers(state.Headers...).
		Width(tableWidth).
		Height(tableHeight).
		Border(lipgloss.RoundedBorder()).
		BorderTop(true).
		BorderBottom(true).
		BorderLeft(true).
		BorderRight(true).
		BorderHeader(true).
		BorderColumn(true).
		BorderRow(false).
		BorderStyle(state.BorderStyle).
		Rows(state.Content.Rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				if col >= 0 && col < len(state.HeaderStyles) {
					return state.HeaderStyles[col]
				}
				return lipgloss.NewStyle()
			}
			if row < 0 || row >= len(state.Content.CellStyles) || col < 0 || col >= len(state.Content.CellStyles[row]) {
				return lipgloss.NewStyle()
			}
			return state.Content.CellStyles[row][col]
		})

	grid := t.Render()
	return PlaceBox(state.InnerW, state.GridH, state.VAlign, grid, state.Bg)
}

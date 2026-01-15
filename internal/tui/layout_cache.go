// Package tui provides the terminal user interface for sancho.
package tui

import "github.com/charmbracelet/lipgloss"

// LayoutCache stores layout dimensions and styles derived from the window size.
type LayoutCache struct {
	InnerW int
	InnerH int

	FooterH int
	GridH   int

	GridTableStyle lipgloss.Style
	FooterAuxStyle lipgloss.Style
	StatusAuxStyle lipgloss.Style
	HelpAuxStyle   lipgloss.Style

	StatsBarStyle      lipgloss.Style
	PromptStyle        lipgloss.Style
	PromptFocusedStyle lipgloss.Style
	PromptContentWidth int
}

func promptContentWidth(styles *Styles, innerW int) int {
	promptFrameW, _ := styles.PromptStyle.GetFrameSize()
	promptWidth := innerW - promptFrameW
	if promptWidth < 0 {
		promptWidth = 0
	}
	if promptWidth < 20 && innerW >= promptFrameW+20 {
		promptWidth = 20
	}
	return promptWidth
}

func (m Model) buildLayoutCache(width, height int) LayoutCache {
	styles := m.styles
	appH, appV := styles.AppStyle.GetFrameSize()
	innerW := width - appH
	innerH := height - appV

	if innerW < 0 {
		innerW = 0
	}
	if innerH < 0 {
		innerH = 0
	}

	footerH := footerCompact
	if innerH >= footerFullMinHeight {
		promptWidth := promptContentWidth(styles, innerW)
		footerH = m.fullFooterHeight(innerH, promptWidth)
	}

	gridH := innerH - footerH
	if gridH < 2 {
		gridH = 2
	}

	gridTableStyle := styles.TableStyle.
		Width(max(0, innerW-2)).
		Height(gridH).
		Border(lipgloss.RoundedBorder(), false, true, true, true)

	footerAuxStyle := lipgloss.NewStyle().
		Padding(0, 0).
		Width(max(0, innerW)).
		Background(styles.colorBg)
	statusAuxStyle := styles.StatusStyle.Inherit(footerAuxStyle)
	helpAuxStyle := styles.HelpStyle.Inherit(lipgloss.NewStyle().
		Padding(0, 1).
		Width(max(0, innerW-2)).
		Background(styles.colorBg))

	statsWidth := max(0, innerW)
	statsBarStyle := styles.StatsBarStyle.Width(statsWidth)

	promptWidth := promptContentWidth(styles, innerW)
	promptStyle := styles.PromptStyle.Width(promptWidth)
	promptFocusedStyle := styles.PromptFocusedStyle.Width(promptWidth)

	return LayoutCache{
		InnerW:             innerW,
		InnerH:             innerH,
		FooterH:            footerH,
		GridH:              gridH,
		GridTableStyle:     gridTableStyle,
		FooterAuxStyle:     footerAuxStyle,
		StatusAuxStyle:     statusAuxStyle,
		HelpAuxStyle:       helpAuxStyle,
		StatsBarStyle:      statsBarStyle,
		PromptStyle:        promptStyle,
		PromptFocusedStyle: promptFocusedStyle,
		PromptContentWidth: promptWidth,
	}
}

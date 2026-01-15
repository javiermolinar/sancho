package view

import "github.com/charmbracelet/lipgloss"

// FooterViewState holds the strings needed to render the footer section.
type FooterViewState struct {
	InnerW     int
	FooterH    int
	FullFooter bool
	StatsLine  string
	LegendLine string
	PromptLine string
	StatusLine string
	HelpLine   string
	VAlign     lipgloss.Position
	Bg         lipgloss.Color
}

// RenderFooter renders stats, legend, prompt, status, and help lines.
func RenderFooter(state FooterViewState) string {
	if state.FooterH <= 0 {
		return ""
	}

	var s string
	if state.FullFooter {
		s += state.StatsLine + "\n"
		s += state.LegendLine + "\n"
		s += state.PromptLine + "\n"
		s += state.StatusLine + "\n"
		s += state.HelpLine
	} else {
		s += state.StatusLine + "\n"
		s += state.HelpLine
	}

	return PlaceBox(state.InnerW, state.FooterH, state.VAlign, s, state.Bg)
}

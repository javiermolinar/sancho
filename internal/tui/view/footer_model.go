package view

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// FooterModel contains content and styles for rendering the footer.
type FooterModel struct {
	InnerW           int
	FooterH          int
	FullFooter       bool
	StatsLine        string
	LegendText       string
	StatusText       string
	HelpText         string
	PromptLines      []string
	PromptMax        int
	PromptFocus      bool
	ShowPrompt       bool
	FooterStyle      lipgloss.Style
	StatusStyle      lipgloss.Style
	HelpStyle        lipgloss.Style
	PromptStyle      lipgloss.Style
	PromptFocusStyle lipgloss.Style
	VAlign           lipgloss.Position
	Bg               lipgloss.Color
}

// RenderFooterModel builds footer lines and renders the footer.
func RenderFooterModel(model FooterModel) string {
	legendLine := footerLine(model.InnerW, model.FooterStyle, model.LegendText)
	statusLine := footerLine(model.InnerW, model.StatusStyle, model.StatusText)
	helpLine := footerLine(model.InnerW, model.HelpStyle, model.HelpText)

	promptStyle := model.PromptStyle
	if model.PromptFocus {
		promptStyle = model.PromptFocusStyle
	}

	promptLine := RenderPrompt(model.InnerW, promptStyle, model.PromptLines)
	if !model.ShowPrompt {
		promptLine = RenderPromptPlaceholder(model.InnerW, model.PromptStyle, model.PromptMax)
	}

	state := FooterViewState{
		InnerW:     model.InnerW,
		FooterH:    model.FooterH,
		FullFooter: model.FullFooter,
		StatsLine:  model.StatsLine,
		LegendLine: legendLine,
		PromptLine: promptLine,
		StatusLine: statusLine,
		HelpLine:   helpLine,
		VAlign:     model.VAlign,
		Bg:         model.Bg,
	}

	return RenderFooter(state)
}

func footerLine(width int, style lipgloss.Style, content string) string {
	frameW, _ := style.GetFrameSize()
	contentWidth := width - frameW
	if contentWidth < 0 {
		contentWidth = 0
	}
	style = style.Width(contentWidth)
	if contentWidth > 0 {
		content = ansi.Truncate(content, contentWidth, "")
	}
	return style.Render(content)
}

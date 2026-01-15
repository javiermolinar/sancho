// Package tui provides the terminal user interface for sancho.
package tui

import (
	"github.com/javiermolinar/sancho/internal/tui/input"
	"github.com/javiermolinar/sancho/internal/tui/view"
)

var promptCommands = []input.PromptCommand{
	{
		Name:        "/plan",
		Description: "Plan tasks from natural language input",
	},
	{
		Name:        "/week",
		Description: "Summarize the current week",
	},
	{
		Name:        "/help",
		Description: "Show available commands",
	},
	{
		Name:        "/reflect",
		Description: "Reflect on recent work",
	},
}

func (m Model) fullFooterHeight(innerH, promptWidth int) int {
	promptLines := max(promptMinContentLines, m.promptContentLineCount(promptWidth))
	promptHeight := promptLines + promptBorderLines
	desired := footerBaseLines + promptHeight

	maxFooter := innerH - 2
	if maxFooter < footerMinHeight {
		return footerCompact
	}
	if desired > maxFooter {
		desired = maxFooter
	}
	if desired < footerMinHeight {
		desired = footerMinHeight
	}
	return desired
}

func (m Model) promptContentLineCount(contentWidth int) int {
	return len(m.promptLines(contentWidth))
}

func (m Model) promptMaxContentLines() int {
	maxLines := m.layoutCache.FooterH - footerBaseLines - promptBorderLines
	if maxLines < promptMinContentLines {
		return promptMinContentLines
	}
	return maxLines
}

func (m Model) promptLines(contentWidth int) []string {
	state := view.PromptState{
		Value:      m.prompt.Value(),
		Cursor:     m.promptCursor(),
		ModePrompt: m.mode == ModePrompt,
	}
	commands := make([]view.PromptCommand, 0, len(promptCommands))
	for _, cmd := range promptCommands {
		commands = append(commands, view.PromptCommand{Name: cmd.Name, Description: cmd.Description})
	}
	return view.PromptLines(state, contentWidth, commands)
}

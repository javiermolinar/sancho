package view

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// PromptCommand is a command suggestion entry.
type PromptCommand struct {
	Name        string
	Description string
}

// PromptState captures prompt input state for rendering.
type PromptState struct {
	Value      string
	Cursor     string
	ModePrompt bool
}

// PromptLines builds prompt input and suggestion lines for the given width.
func PromptLines(state PromptState, contentWidth int, commands []PromptCommand) []string {
	lines := promptInputLines(state, contentWidth)
	lines = append(lines, promptSuggestionLines(state, contentWidth, commands)...)
	return lines
}

// ClampPromptLines clamps prompt lines to maxLines and adds an ellipsis if needed.
func ClampPromptLines(lines []string, maxLines, width int) []string {
	if maxLines <= 0 {
		return nil
	}
	if len(lines) <= maxLines {
		return lines
	}

	clamped := append([]string(nil), lines[:maxLines]...)
	clamped[maxLines-1] = addEllipsis(clamped[maxLines-1], width)
	return clamped
}

// WrapTextToWidths wraps text across the provided widths.
func WrapTextToWidths(s string, firstWidth, otherWidth int) []string {
	if firstWidth <= 0 || otherWidth <= 0 {
		return []string{""}
	}

	runes := []rune(s)
	if len(runes) == 0 {
		return []string{""}
	}

	lines := make([]string, 0, 4)
	width := firstWidth
	lineStart := 0
	lastSpace := -1
	lineWidth := 0

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == ' ' {
			lastSpace = i
		}

		runeWidth := runewidth.RuneWidth(r)
		if lineWidth+runeWidth > width {
			if lastSpace >= lineStart {
				lines = append(lines, string(runes[lineStart:lastSpace]))
				i = lastSpace
				lineStart = lastSpace + 1
			} else {
				lines = append(lines, string(runes[lineStart:i]))
				lineStart = i
				i--
			}
			width = otherWidth
			lastSpace = -1
			lineWidth = 0
			continue
		}
		lineWidth += runeWidth
	}

	lines = append(lines, string(runes[lineStart:]))
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

// RenderPrompt renders the prompt box with the provided lines.
func RenderPrompt(width int, style lipgloss.Style, lines []string) string {
	frameW, _ := style.GetFrameSize()
	contentWidth := width - frameW
	if contentWidth < 0 {
		contentWidth = 0
	}
	style = style.Width(contentWidth)
	if len(lines) == 0 {
		lines = []string{""}
	}
	return style.Render(strings.Join(lines, "\n"))
}

// RenderPromptPlaceholder renders an empty prompt box with matching height.
func RenderPromptPlaceholder(width int, style lipgloss.Style, maxContentLines int) string {
	frameW, _ := style.GetFrameSize()
	contentWidth := width - frameW
	if contentWidth < 0 {
		contentWidth = 0
	}
	style = style.Width(contentWidth)
	if maxContentLines < 1 {
		maxContentLines = 1
	}
	content := strings.Repeat("\n", maxContentLines-1)
	return style.Render(content)
}

func promptInputLines(state PromptState, contentWidth int) []string {
	value := state.Value + state.Cursor
	return wrapTextWithPrefix(value, "> ", "  ", contentWidth)
}

func promptSuggestionLines(state PromptState, contentWidth int, commands []PromptCommand) []string {
	if !state.ModePrompt {
		return nil
	}

	input := state.Value
	suggestions := promptMatchingCommands(input, commands)
	if len(suggestions) == 0 {
		return nil
	}

	lines := make([]string, 0, len(suggestions))
	for _, cmd := range suggestions {
		line := cmd.Name + " " + cmd.Description
		lines = append(lines, wrapTextWithPrefix(line, "  ", "  ", contentWidth)...)
	}
	return lines
}

func promptMatchingCommands(input string, commands []PromptCommand) []PromptCommand {
	if !strings.HasPrefix(strings.TrimSpace(input), "/") {
		return nil
	}
	if strings.Contains(input, " ") {
		return nil
	}

	prefix := strings.ToLower(strings.TrimSpace(input))
	matches := make([]PromptCommand, 0, len(commands))
	for _, cmd := range commands {
		if strings.HasPrefix(strings.ToLower(cmd.Name), prefix) {
			matches = append(matches, cmd)
		}
	}
	return matches
}

func wrapTextWithPrefix(s, prefix, continuation string, width int) []string {
	if width <= 0 {
		return []string{""}
	}

	firstWidth := width - len(prefix)
	if firstWidth < 0 {
		firstWidth = 0
	}
	otherWidth := width - len(continuation)
	if otherWidth < 0 {
		otherWidth = 0
	}

	lines := WrapTextToWidths(s, firstWidth, otherWidth)
	if len(lines) == 0 {
		return []string{prefix}
	}

	for i := range lines {
		if i == 0 {
			lines[i] = prefix + lines[i]
		} else {
			lines[i] = continuation + lines[i]
		}
	}
	return lines
}

func addEllipsis(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if width == 1 {
		return "."
	}
	if width == 2 {
		return ".."
	}
	if len(s) >= width {
		return s[:width-3] + "..."
	}
	if len(s)+3 > width {
		if width-3 < 0 {
			return "..."
		}
		return s[:width-3] + "..."
	}
	return s + "..."
}

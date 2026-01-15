package view

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// PlaceBox renders content in a lipgloss.Place box with background fill.
func PlaceBox(w, h int, vAlign lipgloss.Position, content string, bg lipgloss.Color) string {
	placed := lipgloss.Place(
		w,
		h,
		lipgloss.Left,
		vAlign,
		content,
		lipgloss.WithWhitespaceBackground(bg),
	)
	return PadLinesWithBackground(placed, w, h, bg)
}

// PadLinesWithBackground pads content to width/height with a background color.
func PadLinesWithBackground(content string, width, height int, bg lipgloss.Color) string {
	if width <= 0 || height <= 0 {
		return content
	}
	lines := strings.Split(content, "\n")
	paddingStyle := lipgloss.NewStyle().Background(bg)
	for len(lines) < height {
		lines = append(lines, "")
	}
	for i := 0; i < height; i++ {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		lineWidth := lipgloss.Width(line)
		if lineWidth > width {
			lines[i] = line
			continue
		}
		lines[i] = line + paddingStyle.Render(strings.Repeat(" ", width-lineWidth))
	}
	if len(lines) > height {
		lines = lines[:height]
	}
	return strings.Join(lines, "\n")
}

// RenderModalOverlay centers modalContent and splices it over the base content.
func RenderModalOverlay(baseContent, modalContent string, width, height int, modalBg lipgloss.Color) string {
	modalLines := strings.Split(modalContent, "\n")
	modalHeight := len(modalLines)
	if modalHeight == 0 {
		return baseContent
	}

	modalWidth := 0
	for _, line := range modalLines {
		if w := lipgloss.Width(line); w > modalWidth {
			modalWidth = w
		}
	}
	if modalWidth == 0 {
		return baseContent
	}
	if modalWidth > width {
		modalWidth = width
	}

	top := (height - modalHeight) / 2
	left := (width - modalWidth) / 2
	if top < 0 {
		top = 0
	}
	if left < 0 {
		left = 0
	}

	for i, line := range modalLines {
		lineWidth := lipgloss.Width(line)
		if lineWidth > modalWidth {
			line = ansi.Cut(line, 0, modalWidth)
		}
		if lineWidth < modalWidth {
			paddingStyle := lipgloss.NewStyle().Background(modalBg)
			line += paddingStyle.Render(strings.Repeat(" ", modalWidth-lineWidth))
		}
		line = ApplyModalBackgroundResets(line, modalBg)
		modalLines[i] = line + ansi.ResetStyle
	}

	emptyBg := lipgloss.Color("")
	baseLines := strings.Split(PadLinesWithBackground(baseContent, width, height, emptyBg), "\n")
	if len(baseLines) < height {
		for len(baseLines) < height {
			baseLines = append(baseLines, "")
		}
	}

	lines := make([]string, 0, height)
	for row := 0; row < height; row++ {
		if row < top || row >= top+modalHeight {
			lines = append(lines, baseLines[row])
			continue
		}

		modalLine := modalLines[row-top]
		baseLine := baseLines[row]
		leftSlice := ansi.Cut(baseLine, 0, left)
		rightSlice := ansi.Cut(baseLine, left+modalWidth, width)
		lines = append(lines, leftSlice+modalLine+rightSlice)
	}

	return strings.Join(lines, "\n")
}

// ApplyModalBackgroundResets reapplies modal background after ANSI resets.
func ApplyModalBackgroundResets(line string, modalBg lipgloss.Color) string {
	bgSeq := ModalBackgroundSeq(modalBg)
	if bgSeq == "" {
		return line
	}
	line = strings.ReplaceAll(line, ansi.ResetStyle, ansi.ResetStyle+bgSeq)
	line = strings.ReplaceAll(line, "\x1b[0m", "\x1b[0m"+bgSeq)
	line = strings.ReplaceAll(line, "\x1b[49m", "\x1b[49m"+bgSeq)
	return line
}

// ModalBackgroundSeq returns the background escape sequence for the modal color.
func ModalBackgroundSeq(modalBg lipgloss.Color) string {
	if modalBg == "" {
		return ""
	}
	return ansi.Style{}.BackgroundColor(ansi.HexColor(string(modalBg))).String()
}

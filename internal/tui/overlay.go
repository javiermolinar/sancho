package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const (
	overlayMinWidth  = 18
	overlayMinHeight = 5
	overlayMaxWidth  = 48
	overlayMaxHeight = 12
)

// OverlayModel renders a simple opaque overlay box.
type OverlayModel struct {
	active  bool
	bgColor lipgloss.Color
}

// NewOverlayModel initializes an overlay model.
func NewOverlayModel() OverlayModel {
	return OverlayModel{
		active:  false,
		bgColor: lipgloss.Color(""),
	}
}

// Toggle flips the overlay visibility.
func (o *OverlayModel) Toggle() {
	o.active = !o.active
}

// Active reports whether the overlay is visible.
func (o OverlayModel) Active() bool {
	return o.active
}

// SetBackground updates the overlay background color.
func (o *OverlayModel) SetBackground(color lipgloss.Color) {
	o.bgColor = color
}

// Render draws the overlay on top of base content.
func (o OverlayModel) Render(base string, width, height int, content string) string {
	if !o.active {
		return base
	}
	if width <= 0 || height <= 0 {
		return base
	}

	contentLines := o.contentLines(content)
	contentW, contentH := o.contentSize(contentLines)

	boxW, boxH := o.boxSize(width, height)
	if contentW > boxW {
		boxW = contentW
	}
	if contentH > boxH {
		boxH = contentH
	}
	if boxW > width {
		boxW = width
	}
	if boxH > height {
		boxH = height
	}
	if boxW <= 0 || boxH <= 0 {
		return base
	}

	top := (height - boxH) / 2
	left := (width - boxW) / 2
	if top < 0 {
		top = 0
	}
	if left < 0 {
		left = 0
	}

	baseLines := o.normalizeBase(base, width, height)
	overlayLines := o.overlayLines(boxW, boxH)
	overlayLines = o.applyContent(overlayLines, contentLines, boxW, boxH)

	lines := make([]string, 0, height)
	for row := 0; row < height; row++ {
		if row < top || row >= top+boxH {
			lines = append(lines, baseLines[row])
			continue
		}

		overlayLine := overlayLines[row-top]
		baseLine := baseLines[row]
		leftSlice := ansi.Cut(baseLine, 0, left)
		rightSlice := ansi.Cut(baseLine, left+boxW, width)
		lines = append(lines, leftSlice+overlayLine+rightSlice)
	}

	return strings.Join(lines, "\n")
}

func (o OverlayModel) boxSize(width, height int) (int, int) {
	if width <= 0 || height <= 0 {
		return 0, 0
	}

	boxW := width / 2
	boxH := height / 3

	if boxW < overlayMinWidth {
		boxW = overlayMinWidth
	}
	if boxH < overlayMinHeight {
		boxH = overlayMinHeight
	}
	if boxW > overlayMaxWidth {
		boxW = overlayMaxWidth
	}
	if boxH > overlayMaxHeight {
		boxH = overlayMaxHeight
	}
	if boxW > width {
		boxW = width
	}
	if boxH > height {
		boxH = height
	}

	return boxW, boxH
}

func (o OverlayModel) overlayLines(width, height int) []string {
	if width <= 0 || height <= 0 {
		return nil
	}

	fill := strings.Repeat(" ", width)
	bgSeq := ansi.Style{}.BackgroundColor(ansi.HexColor(string(o.bgColor))).String()
	resetSeq := ansi.ResetStyle
	line := bgSeq + fill + resetSeq

	lines := make([]string, height)
	for i := 0; i < height; i++ {
		lines[i] = line
	}

	return lines
}

func (o OverlayModel) applyContent(lines []string, content []string, width, height int) []string {
	if len(lines) == 0 || len(content) == 0 || width <= 0 || height <= 0 {
		return lines
	}

	contentW, contentH := o.contentSize(content)
	if contentW == 0 || contentH == 0 {
		return lines
	}
	if contentW > width {
		contentW = width
	}
	if contentH > height {
		contentH = height
	}

	top := (height - contentH) / 2
	left := (width - contentW) / 2
	if top < 0 {
		top = 0
	}
	if left < 0 {
		left = 0
	}

	bgSeq := ansi.Style{}.BackgroundColor(ansi.HexColor(string(o.bgColor))).String()
	for i := 0; i < contentH; i++ {
		idx := top + i
		if idx >= len(lines) {
			break
		}
		line := content[i]
		lineWidth := lipgloss.Width(line)
		if lineWidth > contentW {
			line = ansi.Cut(line, 0, contentW)
			lineWidth = contentW
		}
		if lineWidth < contentW {
			line += strings.Repeat(" ", contentW-lineWidth)
		}
		line = o.applyOverlayBackgroundResets(line, bgSeq)

		leftPad := left
		rightPad := width - left - contentW
		if rightPad < 0 {
			rightPad = 0
		}
		lines[idx] = bgSeq + strings.Repeat(" ", leftPad) + line + bgSeq + strings.Repeat(" ", rightPad) + ansi.ResetStyle
	}

	return lines
}

func (o OverlayModel) contentLines(content string) []string {
	if content == "" {
		return nil
	}
	lines := strings.Split(content, "\n")
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func (o OverlayModel) contentSize(lines []string) (int, int) {
	if len(lines) == 0 {
		return 0, 0
	}
	maxWidth := 0
	for _, line := range lines {
		if w := lipgloss.Width(line); w > maxWidth {
			maxWidth = w
		}
	}
	return maxWidth, len(lines)
}

func (o OverlayModel) applyOverlayBackgroundResets(line, bgSeq string) string {
	if bgSeq == "" || line == "" {
		return line
	}
	line = strings.ReplaceAll(line, ansi.ResetStyle, ansi.ResetStyle+bgSeq)
	line = strings.ReplaceAll(line, "\x1b[0m", "\x1b[0m"+bgSeq)
	line = strings.ReplaceAll(line, "\x1b[49m", "\x1b[49m"+bgSeq)
	return line
}

func (o OverlayModel) normalizeBase(base string, width, height int) []string {
	lines := strings.Split(base, "\n")
	for len(lines) < height {
		lines = append(lines, "")
	}
	if len(lines) > height {
		lines = lines[:height]
	}

	for i, line := range lines {
		lineWidth := lipgloss.Width(line)
		if lineWidth > width {
			lines[i] = ansi.Cut(line, 0, width)
			continue
		}
		if lineWidth < width {
			lines[i] = line + strings.Repeat(" ", width-lineWidth)
		}
	}

	return lines
}

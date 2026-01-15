// Package theme provides color themes for the TUI.
package theme

import (
	"math"

	"github.com/charmbracelet/lipgloss"
)

// Palette holds precomputed colors derived from a Theme.
type Palette struct {
	Bg          lipgloss.Color
	BgHighlight lipgloss.Color
	BgSelection lipgloss.Color
	Fg          lipgloss.Color
	FgMuted     lipgloss.Color
	Accent      lipgloss.Color
	Deep        lipgloss.Color
	Shallow     lipgloss.Color
	Current     lipgloss.Color
	Warning     lipgloss.Color

	DeepBg           lipgloss.Color
	ShallowBg        lipgloss.Color
	DeepBgAlt        lipgloss.Color
	ShallowBgAlt     lipgloss.Color
	DeepPastBg       lipgloss.Color
	ShallowPastBg    lipgloss.Color
	DeepPastBgAlt    lipgloss.Color
	ShallowPastBgAlt lipgloss.Color

	TextOnAccent  lipgloss.Color
	TextOnWarning lipgloss.Color
	TextOnCurrent lipgloss.Color
	TextOnDeep    lipgloss.Color
	TextOnShallow lipgloss.Color

	Modal ModalColors
}

// ModalColors holds modal-specific colors derived from a Theme.
type ModalColors struct {
	Bg          lipgloss.Color
	Border      lipgloss.AdaptiveColor
	Text        lipgloss.AdaptiveColor
	Muted       lipgloss.AdaptiveColor
	Highlight   lipgloss.AdaptiveColor
	Panel       lipgloss.AdaptiveColor
	ReverseText lipgloss.AdaptiveColor
	Backdrop    lipgloss.Color
}

// NewPalette derives a Palette from the provided Theme.
func NewPalette(t *Theme) *Palette {
	if t == nil {
		t, _ = Load("mocha")
	}

	isLight := isLightTheme(t.Bg)
	deepBgHex := taskBaseBg(t.Deep, t.Bg, isLight)
	shallowBgHex := taskBaseBg(t.Shallow, t.Bg, isLight)
	deepPastHex := taskMutedBg(t.Deep, t.Bg, isLight)
	shallowPastHex := taskMutedBg(t.Shallow, t.Bg, isLight)
	deepBgAltHex := alternateShade(deepBgHex, isLight)
	shallowBgAltHex := alternateShade(shallowBgHex, isLight)
	deepPastAltHex := alternateShade(deepPastHex, isLight)
	shallowPastAltHex := alternateShade(shallowPastHex, isLight)

	modalPalette := t.Modal()
	modalBgHex := coalesce(modalPalette.BaseBg, t.BgHighlight, t.Bg)
	modalTextHex := coalesce(modalPalette.TextPrimary, t.Fg)
	modalMutedHex := coalesce(modalPalette.TextMuted, t.FgMuted)
	modalHighlightHex := coalesce(modalPalette.Highlight, t.BgSelection, t.Accent)
	modalBorderHex := coalesce(modalPalette.ModalBorder, t.Accent)
	modalPanelHex := coalesce(t.BgSelection, t.BgHighlight, t.Bg)
	modalBackdropHex := coalesce(t.BgSelection, t.BgHighlight, t.Bg)

	return &Palette{
		Bg:          lipgloss.Color(t.Bg),
		BgHighlight: lipgloss.Color(t.BgHighlight),
		BgSelection: lipgloss.Color(t.BgSelection),
		Fg:          lipgloss.Color(t.Fg),
		FgMuted:     lipgloss.Color(t.FgMuted),
		Accent:      lipgloss.Color(t.Accent),
		Deep:        lipgloss.Color(t.Deep),
		Shallow:     lipgloss.Color(t.Shallow),
		Current:     lipgloss.Color(t.Current),
		Warning:     lipgloss.Color(t.Warning),

		DeepBg:           lipgloss.Color(deepBgHex),
		ShallowBg:        lipgloss.Color(shallowBgHex),
		DeepBgAlt:        lipgloss.Color(deepBgAltHex),
		ShallowBgAlt:     lipgloss.Color(shallowBgAltHex),
		DeepPastBg:       lipgloss.Color(deepPastHex),
		ShallowPastBg:    lipgloss.Color(shallowPastHex),
		DeepPastBgAlt:    lipgloss.Color(deepPastAltHex),
		ShallowPastBgAlt: lipgloss.Color(shallowPastAltHex),

		TextOnAccent:  lipgloss.Color(chooseTextColor(t.Accent, t.Bg, t.Fg)),
		TextOnWarning: lipgloss.Color(chooseTextColor(t.Warning, t.Bg, t.Fg)),
		TextOnCurrent: lipgloss.Color(chooseTextColor(t.Current, t.Bg, t.Fg)),
		TextOnDeep:    lipgloss.Color(chooseTextColor(t.Deep, t.Bg, t.Fg)),
		TextOnShallow: lipgloss.Color(chooseTextColor(t.Shallow, t.Bg, t.Fg)),

		Modal: ModalColors{
			Bg:          lipgloss.Color(modalBgHex),
			Border:      adaptiveColor(modalBorderHex),
			Text:        adaptiveColor(modalTextHex),
			Muted:       adaptiveColor(modalMutedHex),
			Highlight:   adaptiveColor(modalHighlightHex),
			Panel:       adaptiveColor(modalPanelHex),
			ReverseText: reverseTextColor(modalBgHex, modalTextHex),
			Backdrop:    lipgloss.Color(modalBackdropHex),
		},
	}
}

func isLightTheme(bg string) bool {
	return relativeLuminance(bg) > 0.55
}

func taskBaseBg(accent, bg string, isLight bool) string {
	if isLight {
		return blendColors(accent, bg, 0.75)
	}
	return darkenColor(accent)
}

func taskMutedBg(accent, bg string, isLight bool) string {
	if isLight {
		return blendColors(accent, bg, 0.88)
	}
	return muteColor(accent)
}

// darkenColor creates a darker version of a hex color for backgrounds.
// It reduces the brightness by blending towards black, with a minimum floor
// to ensure visibility on dark themes.
func darkenColor(hex string) string {
	if len(hex) != 7 || hex[0] != '#' {
		return hex
	}

	var r, g, b int
	parseHex(hex[1:3], &r)
	parseHex(hex[3:5], &g)
	parseHex(hex[5:7], &b)

	factor := 0.50
	r = int(float64(r) * factor)
	g = int(float64(g) * factor)
	b = int(float64(b) * factor)

	minBrightness := 40
	if r < minBrightness {
		r = minBrightness
	}
	if g < minBrightness {
		g = minBrightness
	}
	if b < minBrightness {
		b = minBrightness
	}

	return formatHexColor(r, g, b)
}

// muteColor creates a more heavily muted version of a hex color for past tasks.
func muteColor(hex string) string {
	if len(hex) != 7 || hex[0] != '#' {
		return hex
	}

	var r, g, b int
	parseHex(hex[1:3], &r)
	parseHex(hex[3:5], &g)
	parseHex(hex[5:7], &b)

	factor := 0.30
	r = int(float64(r) * factor)
	g = int(float64(g) * factor)
	b = int(float64(b) * factor)

	minBrightness := 30
	if r < minBrightness {
		r = minBrightness
	}
	if g < minBrightness {
		g = minBrightness
	}
	if b < minBrightness {
		b = minBrightness
	}

	return formatHexColor(r, g, b)
}

// alternateShade creates a subtle alternate shade for adjacent tasks.
func alternateShade(hex string, isLight bool) string {
	if len(hex) != 7 || hex[0] != '#' {
		return hex
	}

	if isLight {
		return blendColors(hex, "#000000", 0.10)
	}
	return blendColors(hex, "#ffffff", 0.30)
}

// parseHex parses a 2-character hex string into an integer.
func parseHex(s string, v *int) {
	var val int
	for i := 0; i < len(s); i++ {
		val *= 16
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
			val += int(c - '0')
		case c >= 'a' && c <= 'f':
			val += int(c - 'a' + 10)
		case c >= 'A' && c <= 'F':
			val += int(c - 'A' + 10)
		}
	}
	*v = val
}

// formatHexColor formats RGB values as a hex color string.
func formatHexColor(r, g, b int) string {
	const hex = "0123456789abcdef"
	result := make([]byte, 7)
	result[0] = '#'
	result[1] = hex[r>>4]
	result[2] = hex[r&0xf]
	result[3] = hex[g>>4]
	result[4] = hex[g&0xf]
	result[5] = hex[b>>4]
	result[6] = hex[b&0xf]
	return string(result)
}

func adaptiveColor(hex string) lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{
		Dark:  hex,
		Light: hex,
	}
}

func reverseTextColor(darkBg, lightText string) lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{
		Dark:  darkBg,
		Light: lightText,
	}
}

func chooseTextColor(bg, lightText, darkText string) string {
	lightContrast := contrastRatio(bg, lightText)
	darkContrast := contrastRatio(bg, darkText)
	if lightContrast >= darkContrast {
		return lightText
	}
	return darkText
}

func contrastRatio(a, b string) float64 {
	l1 := relativeLuminance(a)
	l2 := relativeLuminance(b)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

func relativeLuminance(hex string) float64 {
	if len(hex) != 7 || hex[0] != '#' {
		return 0
	}
	var r, g, b int
	parseHex(hex[1:3], &r)
	parseHex(hex[3:5], &g)
	parseHex(hex[5:7], &b)
	return 0.2126*srgbToLinear(r) + 0.7152*srgbToLinear(g) + 0.0722*srgbToLinear(b)
}

func srgbToLinear(c int) float64 {
	v := float64(c) / 255.0
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

func blendColors(a, b string, ratio float64) string {
	if len(a) != 7 || a[0] != '#' || len(b) != 7 || b[0] != '#' {
		return a
	}
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	var ar, ag, ab int
	var br, bg, bb int
	parseHex(a[1:3], &ar)
	parseHex(a[3:5], &ag)
	parseHex(a[5:7], &ab)
	parseHex(b[1:3], &br)
	parseHex(b[3:5], &bg)
	parseHex(b[5:7], &bb)

	r := int(float64(ar)*(1-ratio) + float64(br)*ratio)
	g := int(float64(ag)*(1-ratio) + float64(bg)*ratio)
	bv := int(float64(ab)*(1-ratio) + float64(bb)*ratio)

	return formatHexColor(r, g, bv)
}

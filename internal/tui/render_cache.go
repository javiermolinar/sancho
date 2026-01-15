// Package tui provides the terminal user interface for sancho.
package tui

// RenderCache stores pre-rendered ANSI strings for hot view paths.
type RenderCache struct {
	EmptyCell        string
	TimeLabelPrefix  []string
	TimeBlankPrefix  string
	VerticalSep      string
	HorizontalSep    string
	ExtraLinePadding string
}

// Package view provides view composition helpers for the TUI.
package view

// OverlayRenderer renders modal overlays on top of base content.
type OverlayRenderer interface {
	Render(base string, width, height int, content string) string
}

// ViewState contains pre-rendered content and overlay metadata.
type ViewState struct {
	Width            int
	Height           int
	BaseContent      string
	ModalContent     string
	ShowModal        bool
	Overlay          OverlayRenderer
	EmptyPlaceholder string
}

// Render composes the final view output.
func Render(state ViewState) string {
	if state.Width == 0 || state.Height == 0 {
		if state.EmptyPlaceholder != "" {
			return state.EmptyPlaceholder
		}
		return "Loading..."
	}

	base := state.BaseContent
	if state.ShowModal && state.Overlay != nil {
		return state.Overlay.Render(base, state.Width, state.Height, state.ModalContent)
	}

	return base
}

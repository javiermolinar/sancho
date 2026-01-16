// Package view provides rendering helpers for the TUI.
package view

import "fmt"

// FormatDuration formats minutes as "Xh Ym".
func FormatDuration(minutes int) string {
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	h := minutes / 60
	m := minutes % 60
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh %dm", h, m)
}

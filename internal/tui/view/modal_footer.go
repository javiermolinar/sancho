// Package view provides rendering helpers for the TUI.
package view

// TaskFormFooter renders the footer for the task form modal.
func TaskFormFooter(styles ModalStyles) string {
	return RenderModalButtons(styles, "[Enter] Save", "[Esc] Cancel")
}

// TaskDetailFooter renders the footer for the task detail modal.
func TaskDetailFooter(isPast bool, styles ModalStyles) string {
	if isPast {
		return RenderModalButtons(styles, "[o] Outcome", "[Esc] Close")
	}
	return RenderModalButtonsCompact(styles, "[o] Outcome", "[e] Edit", "[x] Cancel", "[Esc] Close")
}

// ConfirmDeleteFooter renders the footer for the confirm delete modal.
func ConfirmDeleteFooter(styles ModalStyles) string {
	return RenderModalButtons(styles, "[y/Enter] Confirm", "[n/Esc] Cancel")
}

// PlanResultFooter renders the footer for the plan result modal.
func PlanResultFooter(hasValidationErrors bool, styles ModalStyles) string {
	if hasValidationErrors {
		return RenderModalButtons(styles, "[m] Amend", "[Esc/c] Cancel")
	}
	return RenderModalButtons(styles, "[Enter/a] Apply", "[m] Amend", "[Esc/c] Cancel")
}

// WeekSummaryFooter renders the footer for the week summary modal.
func WeekSummaryFooter(showTasks bool, styles ModalStyles) string {
	if showTasks {
		return RenderModalButtonsCompact(styles, "[s] Summary", "[y] Copy", "[Esc] Close")
	}
	return RenderModalButtonsCompact(styles, "[w] Tasks", "[y] Copy", "[Esc] Close")
}

// InitFooter renders the footer for the init modal.
func InitFooter(styles ModalStyles) string {
	return RenderModalButtons(styles, "[Enter] Allow", "[Esc] Quit")
}

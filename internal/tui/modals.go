// Package tui provides the terminal user interface for sancho.
package tui

import "github.com/javiermolinar/sancho/internal/tui/view"

// renderModal renders the current modal.
func (m Model) renderModal() string {
	switch m.modalType {
	case ModalTaskForm:
		return m.renderTaskFormModal()
	case ModalTaskDetail:
		return m.renderTaskDetailModal()
	case ModalConfirmDelete:
		return m.renderConfirmDeleteModal()
	case ModalPlanResult:
		return m.renderPlanResultModal()
	case ModalWeekSummary:
		return m.renderWeekSummaryModal()
	case ModalInit:
		return m.renderInitModal()
	default:
		return ""
	}
}

func (m Model) modalStyles() view.ModalStyles {
	return view.ModalStyles{
		ModalHeaderStyle:       m.styles.ModalHeaderStyle,
		ModalTitleStyle:        m.styles.ModalTitleStyle,
		ModalFooterStyle:       m.styles.ModalFooterStyle,
		ModalStyle:             m.styles.ModalStyle,
		ModalButtonStyle:       m.styles.ModalButtonStyle,
		ModalButtonActiveStyle: m.styles.ModalButtonActiveStyle,
		ModalBodyStyle:         m.styles.ModalBodyStyle,
	}
}

// renderTaskFormModal renders the task creation form.
func (m Model) renderTaskFormModal() string {
	vm := m.taskFormModalViewModel()
	body := view.RenderTaskFormBody(vm.Model, vm.Styles)
	footer := view.TaskFormFooter(m.modalStyles())
	return view.RenderModalFrame(vm.Title, body, footer, m.modalStyles())
}

// renderTaskDetailModal renders the task detail popup.
func (m Model) renderTaskDetailModal() string {
	vm, ok := m.taskDetailModalViewModel()
	if !ok {
		return ""
	}
	body := view.RenderTaskDetailBody(vm.Model, vm.Styles)
	footer := view.TaskDetailFooter(vm.IsPast, m.modalStyles())
	return view.RenderModalFrame("Task Details", body, footer, m.modalStyles())
}

// renderConfirmDeleteModal renders the delete confirmation modal.
func (m Model) renderConfirmDeleteModal() string {
	vm := m.confirmDeleteModalViewModel()
	body := view.RenderConfirmDeleteBody(vm.Model, vm.Styles)
	footer := view.ConfirmDeleteFooter(m.modalStyles())
	return view.RenderModalFrame("Confirm Cancel", body, footer, m.modalStyles())
}

// renderPlanResultModal renders the LLM planning result modal.
func (m Model) renderPlanResultModal() string {
	vm, ok := m.planResultModalViewModel()
	if !ok {
		return ""
	}
	body := view.RenderPlanResultBody(vm.Model, vm.Styles)
	footer := view.PlanResultFooter(vm.HasValidationErrors, m.modalStyles())
	return view.RenderModalFrame("LLM Draft", body, footer, m.modalStyles())
}

// renderInitModal renders the startup initialization modal.
func (m Model) renderInitModal() string {
	vm := m.initModalViewModel()
	body := view.RenderInitBody(vm.Model, vm.Styles)
	footer := view.InitFooter(m.modalStyles())
	return view.RenderModalFrame("Initialize Sancho", body, footer, m.modalStyles())
}

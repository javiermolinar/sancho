// Package tui provides the terminal user interface for sancho.
package tui

import "github.com/javiermolinar/sancho/internal/tui/view"

const weekSummaryFallbackWidth = 60

func (m Model) renderWeekSummaryModal() string {
	if m.weekSummary == nil {
		return ""
	}
	body := m.weekSummaryBody()
	footer := m.weekSummaryFooter()
	return view.RenderModalFrame("Week Summary", body, footer, m.modalStyles())
}

func (m Model) weekSummaryFooter() string {
	return view.WeekSummaryFooter(m.weekSummaryView == weekSummaryViewTasks, m.modalStyles())
}

func (m Model) weekSummaryBody() string {
	vm := m.weekSummaryBodyViewModel()
	return view.RenderWeekSummaryBody(vm.Lines, vm.Styles, vm.Width)
}

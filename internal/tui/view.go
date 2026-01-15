package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/javiermolinar/sancho/internal/tui/view"
)

// View renders the TUI using a boxed, parent-controlled layout.
func (m Model) View() string {
	state := m.viewState()
	return view.Render(state)
}

func (m Model) viewState() view.ViewState {
	base := m.renderAppContent()
	showModal := m.mode == ModeModal && m.modalType != ModalNone
	modal := ""
	if showModal {
		modal = m.renderModal()
		m.overlay.active = true
		m.overlay.SetBackground(m.styles.ModalBackdropColor)
	} else {
		m.overlay.active = false
	}

	return view.ViewState{
		Width:            m.width,
		Height:           m.height,
		BaseContent:      base,
		ModalContent:     modal,
		ShowModal:        showModal,
		Overlay:          m.overlay,
		EmptyPlaceholder: "Loading...",
	}
}

func (m Model) renderAppContent() string {
	layout := m.layoutCache
	if layout.InnerW <= 0 || layout.InnerH <= 0 {
		return "Terminal too small"
	}

	// 3. Render Sections into Boxes
	gridBox := view.RenderTable(m.tableViewState(layout))
	footerBox := view.RenderFooterModel(m.footerViewState(layout))

	// 4. Assemble Final View
	content := lipgloss.JoinVertical(lipgloss.Left, gridBox, footerBox)
	app := m.styles.AppStyle.Render(content)
	return view.PadLinesWithBackground(app, m.width, m.height, m.styles.colorBg)
}

// placeBox is a helper to render content in an explicit lipgloss box.
func (m Model) placeBox(w, h int, vAlign lipgloss.Position, content string) string {
	return view.PlaceBox(w, h, vAlign, content, m.styles.colorBg)
}

func (m Model) tableViewState(layout LayoutCache) view.TableViewState {
	if layout.GridH <= 0 {
		return view.TableViewState{Render: false}
	}

	totalSlots := m.maxSlots()
	if totalSlots <= 0 {
		return view.TableViewState{Render: false}
	}

	visibleSlots := m.visibleSlotsForTable(layout.GridH)
	if visibleSlots > totalSlots-m.scrollOffset {
		visibleSlots = totalSlots - m.scrollOffset
	}
	if visibleSlots < 0 {
		visibleSlots = 0
	}

	headers, todayCols := view.HeaderLabels(m.weekStart, time.Now())
	rows, cellStyles := m.buildGridTableRows(visibleSlots)

	headerStyles := make([]lipgloss.Style, len(headers))
	if len(headers) > 0 {
		headerStyles[0] = m.styles.TimeColumnStyle.Width(6)
	}
	for i := 1; i < len(headers); i++ {
		style := m.styleCache.DayHeader
		if todayCols[i] {
			style = m.styleCache.DayHeaderToday
		}
		headerStyles[i] = style
	}

	borderStyle := lipgloss.NewStyle().
		Foreground(m.styles.colorAccent).
		Background(m.styles.colorBg)

	return view.TableViewState{
		InnerW:       layout.InnerW,
		GridH:        layout.GridH,
		Headers:      headers,
		HeaderStyles: headerStyles,
		Content: view.TableContent{
			Rows:       rows,
			CellStyles: cellStyles,
		},
		BorderStyle: borderStyle,
		VAlign:      lipgloss.Top,
		Bg:          m.styles.colorBg,
		Render:      true,
	}
}

func (m Model) footerViewState(layout LayoutCache) view.FooterModel {
	legendText := m.renderLegend()
	statusText := m.statusMsgOrDefault()
	helpText := m.renderHelp()

	contentWidth := layout.PromptContentWidth
	lines := m.promptLines(contentWidth)
	lines = view.ClampPromptLines(lines, m.promptMaxContentLines(), contentWidth)

	showPrompt := m.mode != ModeMove && (m.mode != ModeModal || m.modalType == ModalNone)

	return view.FooterModel{
		InnerW:           layout.InnerW,
		FooterH:          layout.FooterH,
		FullFooter:       layout.FooterH >= footerMinHeight,
		StatsLine:        m.renderStatsBar(layout.InnerW),
		LegendText:       legendText,
		StatusText:       statusText,
		HelpText:         helpText,
		PromptLines:      lines,
		PromptMax:        m.promptMaxContentLines(),
		PromptFocus:      m.mode == ModePrompt,
		ShowPrompt:       showPrompt,
		FooterStyle:      layout.FooterAuxStyle,
		StatusStyle:      layout.StatusAuxStyle,
		HelpStyle:        layout.HelpAuxStyle,
		PromptStyle:      layout.PromptStyle,
		PromptFocusStyle: layout.PromptFocusedStyle,
		VAlign:           lipgloss.Bottom,
		Bg:               m.styles.colorBg,
	}
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

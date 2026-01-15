package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/javiermolinar/sancho/internal/tui/commands"
)

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		updated, cmd := m.handleKeyMsg(msg)
		if model, ok := updated.(Model); ok {
			model.refreshCachesIfNeeded()
			return model, cmd
		}
		return updated, cmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.colWidth = m.calculateColWidth()
		m.calculateLayout() // Calculate rowHeight and rowLines together
		m.styleCache = NewStyleCache(m.styles, m.colWidth)
		m.layoutCache = m.buildLayoutCache(m.width, m.height)
		m.refreshViewCaches()
		return m, nil

	case commands.WeekLoadedMsg:
		// Single week reload (after mutations)
		// Get the current week window, update the current week, and rebuild the grid
		ww := m.slotState.WeekWindow()
		if ww != nil {
			ww.SetCurrent(msg.Week)
		}
		// Rebuild slot grid from updated week window
		slotGrid := WeekWindowToSlotGrid(ww, m.slotState.Config())
		m.slotState.SetGrid(slotGrid)
		m.loading = false
		m.refreshViewCaches()
		return m, nil

	case commands.InitialLoadMsg:
		// Initial load of 3 weeks - update config and convert to slot grid
		newConfig := SlotGridConfigFromWeekWindow(msg.Window, m.config.Schedule.DayStart, m.config.Schedule.DayEnd, m.nowFunc(), m.rowHeight)
		m.slotState.UpdateConfig(newConfig)
		slotGrid := WeekWindowToSlotGrid(msg.Window, newConfig)
		m.slotState.SetGrid(slotGrid)
		m.loading = false
		m.focusCursorOnCurrentTaskOrTime()
		m.refreshViewCaches()
		return m, nil

	case commands.WeekShiftedMsg:
		// Shift prev/next week - shift the window and set the newly loaded edge week
		ww := m.slotState.WeekWindow()
		if ww != nil {
			if msg.Forward {
				// Shift forward: prev←current, current←next, then set new next
				ww.ShiftForward(msg.Week)
			} else {
				// Shift backward: next←current, current←prev, then set new prev
				ww.ShiftBackward(msg.Week)
			}
		}
		// Update config with new FirstDate from the shifted week window
		newConfig := SlotGridConfigFromWeekWindow(ww, m.config.Schedule.DayStart, m.config.Schedule.DayEnd, m.nowFunc(), m.rowHeight)
		m.slotState.UpdateConfig(newConfig)
		// Rebuild slot grid from updated week window with new config
		slotGrid := WeekWindowToSlotGrid(ww, newConfig)
		m.slotState.SetGrid(slotGrid)
		m.loading = false
		m.focusCursorOnCurrentTaskOrTime()
		m.refreshViewCaches()
		return m, nil

	case commands.ErrMsg:
		m.err = msg.Err
		m.statusMsg = fmt.Sprintf("Error: %v", msg.Err)
		m.statusTime = time.Now().Add(5 * time.Second)
		return m, nil

	case commands.StatusMsgCmd:
		m.statusMsg = msg.Msg
		m.statusTime = time.Now().Add(3 * time.Second)
		return m, tea.Tick(3*time.Second, func(time.Time) tea.Msg {
			return commands.ClearStatusMsg{}
		})

	case commands.ClearStatusMsg:
		if time.Now().After(m.statusTime) {
			m.statusMsg = ""
		}
		return m, nil

	case commands.PlanStartedMsg:
		m.statusMsg = "Planning..."
		return m, nil

	case commands.PlanResultMsg:
		m.planner = msg.Planner
		m.planResult = msg.Result
		m.mode = ModeModal
		m.modalType = ModalPlanResult
		m.statusMsg = ""
		return m, nil

	case commands.PlanSavedMsg:
		m.statusMsg = fmt.Sprintf("Saved %d tasks", msg.Count)
		m.planResult = nil
		m.planner = nil
		m.mode = ModeNormal
		m.modalType = ModalNone
		return m, commands.LoadWeek(m.repo, m.weekStart)

	case commands.WeekSummaryMsg:
		m.weekSummary = msg.Summary
		m.weekSummaryView = weekSummaryViewSummary
		m.weekSummarySummaryText = buildWeekSummaryLines(msg.Summary, m.config.HasPeakHours())
		m.weekSummaryTasksText = buildWeekTasksLines(msg.Summary)
		m.weekSummaryCopyText = buildWeekTasksCopyText(m.weekSummaryTasksText)
		m.mode = ModeModal
		m.modalType = ModalWeekSummary
		m.statusMsg = ""
		return m, nil
	}

	// Handle prompt input when in prompt mode
	if m.mode == ModePrompt {
		var cmd tea.Cmd
		m.prompt, cmd = m.prompt.Update(msg)
		m.calculateLayout()
		m.layoutCache = m.buildLayoutCache(m.width, m.height)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

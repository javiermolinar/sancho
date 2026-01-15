// Package tui provides the terminal user interface for sancho.
package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"

	"github.com/javiermolinar/sancho/internal/config"
	"github.com/javiermolinar/sancho/internal/dwplanner"
	"github.com/javiermolinar/sancho/internal/task"
	"github.com/javiermolinar/sancho/internal/tui/view"
)

func refreshCachesForTest(m *Model) {
	m.styleCache = NewStyleCache(m.styles, m.colWidth)
	m.markCacheDirty()
	m.refreshCachesIfNeeded()
}

func TestPlaceBox_WhitespaceBackground(t *testing.T) {
	prevProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() {
		lipgloss.SetColorProfile(prevProfile)
	})

	bg := lipgloss.Color("#112233")
	m := &Model{
		styles: &Styles{
			colorBg: bg,
		},
	}
	out := m.placeBox(5, 1, lipgloss.Top, "x")
	bgSeq := "\x1b[48;2;17;34;51m"
	bgIndex := strings.Index(out, bgSeq)
	if bgIndex == -1 {
		t.Fatalf("expected background whitespace in output: %q", out)
	}
	if strings.Index(out, "x") > bgIndex {
		t.Fatalf("expected background after content, got %q", out)
	}
}

func TestRenderCell_TitleOnlyOnFirstSlot(t *testing.T) {
	cfg := &config.Config{
		Schedule: config.ScheduleConfig{
			DayStart: "09:00",
			DayEnd:   "17:00",
		},
	}

	m := New(nil, cfg)
	m.rowHeight = 15
	m.rowLines = 2
	m.colWidth = 20
	m.cursor = Position{Day: 6, Slot: 0}

	monday := time.Date(2025, 1, 6, 0, 0, 0, 0, time.Local)
	testTask := &task.Task{
		ID:             1,
		Description:    "Review 3 PRs",
		Category:       task.CategoryShallow,
		ScheduledDate:  monday,
		ScheduledStart: "11:30",
		ScheduledEnd:   "13:15",
		Status:         task.StatusScheduled,
	}

	week := task.NewWeek(monday)
	if err := week.Day(0).AddTask(testTask); err != nil {
		t.Fatalf("add task: %v", err)
	}
	ww := task.NewWeekWindow(nil, week, nil)

	slotConfig := SlotGridConfigFromWeekWindow(ww, cfg.Schedule.DayStart, cfg.Schedule.DayEnd, time.Now, m.rowHeight)
	m.slotState = NewSlotStateManager(slotConfig)
	m.slotState.SetGrid(WeekWindowToSlotGrid(ww, m.slotState.Config()))

	refreshCachesForTest(m)
	dayStart := task.TimeToMinutes(cfg.Schedule.DayStart)
	firstSlot := (task.TimeToMinutes("11:30") - dayStart) / m.rowHeight

	shadeByDay := m.cachedShadeMap
	dayTasks := m.gridCache[0]
	cursorTask := m.cachedCursorTask()
	firstCell := m.renderCell(firstSlot, 0, dayTasks[firstSlot], dayTasks, cursorTask, shadeByDay)
	if !strings.Contains(firstCell, testTask.Description) {
		t.Errorf("first slot missing title, got %q", firstCell)
	}

	nextSlot := firstSlot + 1
	nextCell := m.renderCell(nextSlot, 0, dayTasks[nextSlot], dayTasks, cursorTask, shadeByDay)
	if strings.Contains(nextCell, "[S]") {
		t.Errorf("next slot unexpectedly repeated indicator, got %q", nextCell)
	}
}

func TestRenderCell_SingleLineShowsTimeWhenSpace(t *testing.T) {
	cfg := &config.Config{
		Schedule: config.ScheduleConfig{
			DayStart: "09:00",
			DayEnd:   "17:00",
		},
	}

	m := New(nil, cfg)
	m.rowHeight = 15
	m.rowLines = 1
	m.colWidth = 26
	m.cursor = Position{Day: 6, Slot: 0}

	monday := time.Date(2025, 1, 6, 0, 0, 0, 0, time.Local)
	testTask := &task.Task{
		ID:             1,
		Description:    "Plan sprint backlog",
		Category:       task.CategoryDeep,
		ScheduledDate:  monday,
		ScheduledStart: "11:30",
		ScheduledEnd:   "11:45",
		Status:         task.StatusScheduled,
	}

	week := task.NewWeek(monday)
	if err := week.Day(0).AddTask(testTask); err != nil {
		t.Fatalf("add task: %v", err)
	}
	ww := task.NewWeekWindow(nil, week, nil)

	slotConfig := SlotGridConfigFromWeekWindow(ww, cfg.Schedule.DayStart, cfg.Schedule.DayEnd, time.Now, m.rowHeight)
	m.slotState = NewSlotStateManager(slotConfig)
	m.slotState.SetGrid(WeekWindowToSlotGrid(ww, m.slotState.Config()))

	refreshCachesForTest(m)
	dayStart := task.TimeToMinutes(cfg.Schedule.DayStart)
	slot := (task.TimeToMinutes("11:30") - dayStart) / m.rowHeight

	shadeByDay := m.cachedShadeMap
	dayTasks := m.gridCache[0]
	cursorTask := m.cachedCursorTask()
	cell := m.renderCell(slot, 0, dayTasks[slot], dayTasks, cursorTask, shadeByDay)
	if !strings.Contains(cell, "Plan") {
		t.Errorf("single-line cell missing description, got %q", cell)
	}
	if !strings.Contains(cell, "11:30-11:45") {
		t.Errorf("single-line cell missing time range, got %q", cell)
	}
}

func TestTaskShadeMap_AdjacentSameCategoryAlternates(t *testing.T) {
	cfg := &config.Config{
		Schedule: config.ScheduleConfig{
			DayStart: "09:00",
			DayEnd:   "17:00",
		},
	}

	m := New(nil, cfg)
	m.rowHeight = 60
	m.rowLines = 2
	m.colWidth = 20

	monday := time.Date(2025, 1, 6, 0, 0, 0, 0, time.Local)
	tasks := []*task.Task{
		{
			ID:             1,
			Description:    "Deep 1",
			Category:       task.CategoryDeep,
			ScheduledDate:  monday,
			ScheduledStart: "09:00",
			ScheduledEnd:   "10:00",
			Status:         task.StatusScheduled,
		},
		{
			ID:             2,
			Description:    "Deep 2",
			Category:       task.CategoryDeep,
			ScheduledDate:  monday,
			ScheduledStart: "10:00",
			ScheduledEnd:   "11:00",
			Status:         task.StatusScheduled,
		},
		{
			ID:             3,
			Description:    "Deep 3",
			Category:       task.CategoryDeep,
			ScheduledDate:  monday,
			ScheduledStart: "11:00",
			ScheduledEnd:   "12:00",
			Status:         task.StatusScheduled,
		},
		{
			ID:             4,
			Description:    "Shallow 1",
			Category:       task.CategoryShallow,
			ScheduledDate:  monday,
			ScheduledStart: "12:00",
			ScheduledEnd:   "13:00",
			Status:         task.StatusScheduled,
		},
		{
			ID:             5,
			Description:    "Shallow 2",
			Category:       task.CategoryShallow,
			ScheduledDate:  monday,
			ScheduledStart: "13:00",
			ScheduledEnd:   "14:00",
			Status:         task.StatusScheduled,
		},
		{
			ID:             6,
			Description:    "Deep gap",
			Category:       task.CategoryDeep,
			ScheduledDate:  monday,
			ScheduledStart: "15:00",
			ScheduledEnd:   "16:00",
			Status:         task.StatusScheduled,
		},
	}

	week := task.NewWeek(monday)
	for _, tsk := range tasks {
		if err := week.Day(0).AddTask(tsk); err != nil {
			t.Fatalf("add task %d: %v", tsk.ID, err)
		}
	}

	ww := task.NewWeekWindow(nil, week, nil)
	slotConfig := SlotGridConfigFromWeekWindow(ww, cfg.Schedule.DayStart, cfg.Schedule.DayEnd, time.Now, m.rowHeight)
	m.slotState = NewSlotStateManager(slotConfig)
	m.slotState.SetGrid(WeekWindowToSlotGrid(ww, m.slotState.Config()))

	refreshCachesForTest(m)
	shadeByDay := m.cachedShadeMap
	dayShade := shadeByDay[0]

	if dayShade == nil {
		t.Fatal("expected day 0 shade map to be set")
	}

	if dayShade[1] {
		t.Errorf("task 1 should use base shade")
	}
	if !dayShade[2] {
		t.Errorf("task 2 should use alternate shade")
	}
	if dayShade[3] {
		t.Errorf("task 3 should use base shade after alternation")
	}
	if dayShade[4] {
		t.Errorf("task 4 should use base shade for new category")
	}
	if !dayShade[5] {
		t.Errorf("task 5 should use alternate shade for adjacent shallow task")
	}
	if dayShade[6] {
		t.Errorf("task 6 should use base shade after gap")
	}
}

func TestRenderLegendIncludesLabels(t *testing.T) {
	cfg := &config.Config{
		Schedule: config.ScheduleConfig{
			DayStart: "09:00",
			DayEnd:   "17:00",
		},
	}

	m := New(nil, cfg)
	legend := m.renderLegend()
	if !strings.Contains(legend, "Legend:") {
		t.Errorf("legend missing label prefix, got %q", legend)
	}
	if !strings.Contains(legend, "[D] Deep") {
		t.Errorf("legend missing deep label, got %q", legend)
	}
	if !strings.Contains(legend, "[S] Shallow") {
		t.Errorf("legend missing shallow label, got %q", legend)
	}
}

func TestRenderModalDimensions(t *testing.T) {
	content := "Modal"
	width := 20
	height := 5

	baseLine := strings.Repeat(" ", width)
	base := strings.Repeat(baseLine+"\n", height-1) + baseLine
	got := view.RenderModalOverlay(base, content, width, height, lipgloss.Color("#000000"))
	lines := strings.Split(got, "\n")
	if len(lines) != height {
		t.Fatalf("RenderModal lines = %d, want %d", len(lines), height)
	}
	for i, line := range lines {
		lineWidth := lipgloss.Width(line)
		if lineWidth != width {
			t.Fatalf("RenderModal line %d width = %d, want %d", i, lineWidth, width)
		}
	}
}

func TestRenderModalOverlay_PaddedLineKeepsBackground(t *testing.T) {
	content := "Hi\nH"
	width := 6
	height := 3
	modalBg := lipgloss.Color("#222222")

	baseLine := strings.Repeat(" ", width)
	base := strings.Repeat(baseLine+"\n", height-1) + baseLine
	got := view.RenderModalOverlay(base, content, width, height, modalBg)
	lines := strings.Split(got, "\n")
	if len(lines) != height {
		t.Fatalf("RenderModal lines = %d, want %d", len(lines), height)
	}

	top := (height - 2) / 2
	row := top + 1
	paddingStyle := lipgloss.NewStyle().Background(modalBg)
	paddedLine := "H" + paddingStyle.Render(" ")
	expectedLine := view.ApplyModalBackgroundResets(paddedLine, modalBg) + ansi.ResetStyle
	if !strings.Contains(lines[row], expectedLine) {
		t.Errorf("expected background to be re-applied after padding reset")
	}
}

func TestApplyModalBackgroundResets_ReappliesBackground(t *testing.T) {
	bg := lipgloss.Color("#101010")
	line := "Hello" + ansi.ResetStyle + "World"
	got := view.ApplyModalBackgroundResets(line, bg)
	bgSeq := view.ModalBackgroundSeq(bg)
	if !strings.Contains(got, ansi.ResetStyle+bgSeq) {
		t.Errorf("expected background to be re-applied after reset")
	}
	line = "Hello\x1b[49mWorld"
	got = view.ApplyModalBackgroundResets(line, bg)
	if !strings.Contains(got, "\x1b[49m"+bgSeq) {
		t.Errorf("expected background to be re-applied after default background")
	}
}

func TestRenderCell_WrapsDescriptionAcrossLines(t *testing.T) {
	cfg := &config.Config{
		Schedule: config.ScheduleConfig{
			DayStart: "09:00",
			DayEnd:   "17:00",
		},
	}

	m := New(nil, cfg)
	m.rowHeight = 15
	m.rowLines = 3
	m.colWidth = 16

	monday := time.Date(2025, 1, 6, 0, 0, 0, 0, time.Local)
	testTask := &task.Task{
		ID:             1,
		Description:    "Write thesis introduction",
		Category:       task.CategoryDeep,
		ScheduledDate:  monday,
		ScheduledStart: "11:30",
		ScheduledEnd:   "12:15",
		Status:         task.StatusScheduled,
	}

	week := task.NewWeek(monday)
	if err := week.Day(0).AddTask(testTask); err != nil {
		t.Fatalf("add task: %v", err)
	}
	ww := task.NewWeekWindow(nil, week, nil)

	slotConfig := SlotGridConfigFromWeekWindow(ww, cfg.Schedule.DayStart, cfg.Schedule.DayEnd, time.Now, m.rowHeight)
	m.slotState = NewSlotStateManager(slotConfig)
	m.slotState.SetGrid(WeekWindowToSlotGrid(ww, m.slotState.Config()))

	dayStart := task.TimeToMinutes(cfg.Schedule.DayStart)
	slot := (task.TimeToMinutes("11:30") - dayStart) / m.rowHeight

	refreshCachesForTest(m)
	shadeByDay := m.cachedShadeMap
	dayTasks := m.gridCache[0]
	cursorTask := m.cachedCursorTask()
	line0 := m.renderCell(slot, 0, dayTasks[slot], dayTasks, cursorTask, shadeByDay)
	line1 := m.renderCell(slot, 1, dayTasks[slot], dayTasks, cursorTask, shadeByDay)
	line2 := m.renderCell(slot, 2, dayTasks[slot], dayTasks, cursorTask, shadeByDay)

	if !strings.Contains(line0, "Write") {
		t.Errorf("line0 missing first wrapped line, got %q", line0)
	}
	if !strings.Contains(line1, "thesis") {
		t.Errorf("line1 missing wrapped continuation, got %q", line1)
	}
	if !strings.Contains(line2, "introduction") && !strings.Contains(line2, "thesis") {
		t.Errorf("line2 missing wrapped continuation, got %q", line2)
	}

	timeIndex := len(m.cachedTaskLines[testTask.ID])
	timeSlot := slot + (timeIndex / m.rowLines)
	timeLine := timeIndex % m.rowLines
	timeCell := m.renderCell(timeSlot, timeLine, dayTasks[timeSlot], dayTasks, cursorTask, shadeByDay)
	if !strings.Contains(timeCell, "11:30-12:15") {
		t.Errorf("time line missing time range, got %q", timeCell)
	}
}

func TestRenderCell_WrapsAcrossSlotsWhenSpace(t *testing.T) {
	cfg := &config.Config{
		Schedule: config.ScheduleConfig{
			DayStart: "09:00",
			DayEnd:   "17:00",
		},
	}

	m := New(nil, cfg)
	m.rowHeight = 15
	m.rowLines = 2
	m.colWidth = 14

	monday := time.Date(2025, 1, 6, 0, 0, 0, 0, time.Local)
	desc := "Alpha Beta Gamma Delta Epsilon Zeta Eta Theta"
	testTask := &task.Task{
		ID:             2,
		Description:    desc,
		Category:       task.CategoryDeep,
		ScheduledDate:  monday,
		ScheduledStart: "10:00",
		ScheduledEnd:   "10:45",
		Status:         task.StatusScheduled,
	}

	week := task.NewWeek(monday)
	if err := week.Day(0).AddTask(testTask); err != nil {
		t.Fatalf("add task: %v", err)
	}
	ww := task.NewWeekWindow(nil, week, nil)

	slotConfig := SlotGridConfigFromWeekWindow(ww, cfg.Schedule.DayStart, cfg.Schedule.DayEnd, time.Now, m.rowHeight)
	m.slotState = NewSlotStateManager(slotConfig)
	m.slotState.SetGrid(WeekWindowToSlotGrid(ww, m.slotState.Config()))

	refreshCachesForTest(m)
	dayStart := task.TimeToMinutes(cfg.Schedule.DayStart)
	slot := (task.TimeToMinutes("10:00") - dayStart) / m.rowHeight

	firstWidth := m.colWidth - 5
	if firstWidth < 1 {
		firstWidth = 1
	}
	otherWidth := m.colWidth - 1
	if otherWidth < 1 {
		otherWidth = 1
	}
	maxLines := (m.taskSlotSpan(testTask) * m.rowLines) - 1
	descLines := wrapTextWithWidths(desc, firstWidth, otherWidth, maxLines)
	if len(descLines) < 3 {
		t.Fatalf("expected wrapped description lines, got %v", descLines)
	}

	shadeByDay := m.cachedShadeMap
	dayTasks := m.gridCache[0]
	cursorTask := m.cachedCursorTask()
	nextSlotLine := m.renderCell(slot+1, 0, dayTasks[slot+1], dayTasks, cursorTask, shadeByDay)
	continuations := []string{"Beta", "Gamma", "Delta", "Epsilon", "Zeta", "Eta", "Theta"}
	found := false
	for _, word := range continuations {
		if strings.Contains(nextSlotLine, word) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected continuation line on next slot, got %q", nextSlotLine)
	}
}

func TestRenderTaskFormModal_ShowsDurationOptions(t *testing.T) {
	cfg := &config.Config{
		Schedule: config.ScheduleConfig{
			DayStart: "09:00",
			DayEnd:   "17:00",
		},
	}

	m := New(nil, cfg)
	m.weekStart = time.Date(2025, 1, 6, 0, 0, 0, 0, time.Local)
	m.cursor = Position{Day: 1, Slot: 4}
	m.formDesc.SetValue("Write summary")
	m.formFocus = 1
	m.formDuration = 1

	view := m.renderTaskFormModal()
	if !strings.Contains(view, "TASK NAME") {
		t.Errorf("expected task name label, got %q", view)
	}
	if !strings.Contains(view, "Write summary") {
		t.Errorf("expected task name value, got %q", view)
	}
	for _, label := range []string{"12m", "30m", "1h"} {
		if !strings.Contains(view, label) {
			t.Errorf("expected duration option %q, got %q", label, view)
		}
	}
}

func TestRenderPlanResultModal_ShowsAmendSection(t *testing.T) {
	cfg := &config.Config{
		Schedule: config.ScheduleConfig{
			DayStart: "09:00",
			DayEnd:   "17:00",
		},
	}

	m := New(nil, cfg)
	m.planResult = &dwplanner.PlanResult{
		SortedDates: []string{"2025-01-06"},
		TasksByDate: map[string][]dwplanner.PlannedTask{
			"2025-01-06": {
				{
					Description:    "Outline quarterly goals",
					Category:       "deep",
					ScheduledDate:  "2025-01-06",
					ScheduledStart: "10:00",
					ScheduledEnd:   "11:00",
				},
			},
		},
	}

	view := m.renderPlanResultModal()
	if !strings.Contains(view, "AMEND") {
		t.Errorf("expected amend section, got %q", view)
	}
	if !strings.Contains(view, "Press m to amend") {
		t.Errorf("expected amend hint, got %q", view)
	}
}

func TestRenderPlanResultModal_ShowsApplyWhenValid(t *testing.T) {
	cfg := &config.Config{
		Schedule: config.ScheduleConfig{
			DayStart: "09:00",
			DayEnd:   "17:00",
		},
	}

	m := New(nil, cfg)
	m.planResult = &dwplanner.PlanResult{
		SortedDates: []string{"2025-01-06"},
		TasksByDate: map[string][]dwplanner.PlannedTask{
			"2025-01-06": {
				{
					Description:    "Plan quarterly review",
					Category:       "deep",
					ScheduledDate:  "2025-01-06",
					ScheduledStart: "10:00",
					ScheduledEnd:   "11:00",
				},
			},
		},
	}

	view := m.renderPlanResultModal()
	if !strings.Contains(view, "Apply") {
		t.Errorf("expected apply action, got %q", view)
	}
	if !strings.Contains(view, "Enter/a") {
		t.Errorf("expected apply shortcut, got %q", view)
	}
}

func TestTaskShadeMap_DisplaySlotAdjacency(t *testing.T) {
	cfg := &config.Config{
		Schedule: config.ScheduleConfig{
			DayStart: "09:00",
			DayEnd:   "17:00",
		},
	}

	m := New(nil, cfg)
	m.rowHeight = 60
	m.rowLines = 2
	m.colWidth = 20

	monday := time.Date(2025, 1, 6, 0, 0, 0, 0, time.Local)
	tasks := []*task.Task{
		{
			ID:             10,
			Description:    "Shallow A",
			Category:       task.CategoryShallow,
			ScheduledDate:  monday,
			ScheduledStart: "11:00",
			ScheduledEnd:   "12:00",
			Status:         task.StatusScheduled,
		},
		{
			ID:             11,
			Description:    "Shallow B",
			Category:       task.CategoryShallow,
			ScheduledDate:  monday,
			ScheduledStart: "12:00",
			ScheduledEnd:   "13:00",
			Status:         task.StatusScheduled,
		},
	}

	week := task.NewWeek(monday)
	for _, tsk := range tasks {
		if err := week.Day(0).AddTask(tsk); err != nil {
			t.Fatalf("add task %d: %v", tsk.ID, err)
		}
	}

	ww := task.NewWeekWindow(nil, week, nil)
	slotConfig := SlotGridConfigFromWeekWindow(ww, cfg.Schedule.DayStart, cfg.Schedule.DayEnd, time.Now, m.rowHeight)
	m.slotState = NewSlotStateManager(slotConfig)
	m.slotState.SetGrid(WeekWindowToSlotGrid(ww, m.slotState.Config()))

	refreshCachesForTest(m)
	shadeByDay := m.cachedShadeMap
	dayShade := shadeByDay[0]
	if dayShade == nil {
		t.Fatal("expected day 0 shade map to be set")
	}

	if dayShade[10] {
		t.Errorf("task 10 should use base shade")
	}
	if !dayShade[11] {
		t.Errorf("task 11 should use alternate shade for adjacent display slot")
	}
}

func TestRenderCell_PrefersTaskStartingInSlot(t *testing.T) {
	cfg := &config.Config{
		Schedule: config.ScheduleConfig{
			DayStart: "09:00",
			DayEnd:   "17:00",
		},
	}

	m := New(nil, cfg)
	m.rowHeight = 60
	m.rowLines = 2
	m.colWidth = 20

	monday := time.Date(2025, 1, 6, 0, 0, 0, 0, time.Local)
	previous := &task.Task{
		ID:             21,
		Description:    "Earlier task",
		Category:       task.CategoryDeep,
		ScheduledDate:  monday,
		ScheduledStart: "09:15",
		ScheduledEnd:   "11:15",
		Status:         task.StatusScheduled,
	}
	next := &task.Task{
		ID:             22,
		Description:    "Starts in slot",
		Category:       task.CategoryShallow,
		ScheduledDate:  monday,
		ScheduledStart: "11:15",
		ScheduledEnd:   "12:15",
		Status:         task.StatusScheduled,
	}

	week := task.NewWeek(monday)
	if err := week.Day(0).AddTask(previous); err != nil {
		t.Fatalf("add previous task: %v", err)
	}
	if err := week.Day(0).AddTask(next); err != nil {
		t.Fatalf("add next task: %v", err)
	}

	ww := task.NewWeekWindow(nil, week, nil)
	slotConfig := SlotGridConfigFromWeekWindow(ww, cfg.Schedule.DayStart, cfg.Schedule.DayEnd, time.Now, m.rowHeight)
	m.slotState = NewSlotStateManager(slotConfig)
	m.slotState.SetGrid(WeekWindowToSlotGrid(ww, m.slotState.Config()))

	dayStart := task.TimeToMinutes(cfg.Schedule.DayStart)
	slot := (task.TimeToMinutes("11:00") - dayStart) / m.rowHeight
	refreshCachesForTest(m)
	shadeByDay := m.cachedShadeMap
	dayTasks := m.gridCache[0]
	cursorTask := m.cachedCursorTask()
	cell := m.renderCell(slot, 0, dayTasks[slot], dayTasks, cursorTask, shadeByDay)
	if !strings.Contains(cell, next.Description) {
		t.Errorf("expected slot to show next task, got %q", cell)
	}
	if strings.Contains(cell, previous.Description) {
		t.Errorf("expected slot to avoid previous task, got %q", cell)
	}
}

func TestRenderCell_GapWhenTaskEndsMidSlot(t *testing.T) {
	cfg := &config.Config{
		Schedule: config.ScheduleConfig{
			DayStart: "09:00",
			DayEnd:   "17:00",
		},
	}

	m := New(nil, cfg)
	m.rowHeight = 60
	m.rowLines = 2
	m.colWidth = 20

	monday := time.Date(2025, 1, 6, 0, 0, 0, 0, time.Local)
	taskA := &task.Task{
		ID:             31,
		Description:    "Ends at half",
		Category:       task.CategoryDeep,
		ScheduledDate:  monday,
		ScheduledStart: "13:00",
		ScheduledEnd:   "15:30",
		Status:         task.StatusScheduled,
	}
	taskB := &task.Task{
		ID:             32,
		Description:    "Next hour",
		Category:       task.CategoryShallow,
		ScheduledDate:  monday,
		ScheduledStart: "16:00",
		ScheduledEnd:   "17:00",
		Status:         task.StatusScheduled,
	}

	week := task.NewWeek(monday)
	if err := week.Day(0).AddTask(taskA); err != nil {
		t.Fatalf("add task A: %v", err)
	}
	if err := week.Day(0).AddTask(taskB); err != nil {
		t.Fatalf("add task B: %v", err)
	}

	ww := task.NewWeekWindow(nil, week, nil)
	slotConfig := SlotGridConfigFromWeekWindow(ww, cfg.Schedule.DayStart, cfg.Schedule.DayEnd, time.Now, m.rowHeight)
	m.slotState = NewSlotStateManager(slotConfig)
	m.slotState.SetGrid(WeekWindowToSlotGrid(ww, m.slotState.Config()))

	dayStart := task.TimeToMinutes(cfg.Schedule.DayStart)
	slot := (task.TimeToMinutes("15:00") - dayStart) / m.rowHeight
	refreshCachesForTest(m)
	shadeByDay := m.cachedShadeMap
	dayTasks := m.gridCache[0]
	cursorTask := m.cachedCursorTask()
	cell := m.renderCell(slot, 0, dayTasks[slot], dayTasks, cursorTask, shadeByDay)
	if strings.Contains(cell, taskA.Description) {
		t.Errorf("expected 15:00 slot to be empty, got %q", cell)
	}
	if strings.Contains(cell, taskB.Description) {
		t.Errorf("expected 15:00 slot to be empty, got %q", cell)
	}
}

func TestRenderCell_TimeRangeAfterTitle(t *testing.T) {
	cfg := &config.Config{
		Schedule: config.ScheduleConfig{
			DayStart: "09:00",
			DayEnd:   "17:00",
		},
	}

	m := New(nil, cfg)
	m.rowHeight = 60
	m.rowLines = 2
	m.colWidth = 24

	monday := time.Date(2025, 1, 6, 0, 0, 0, 0, time.Local)
	taskA := &task.Task{
		ID:             41,
		Description:    "Long task",
		Category:       task.CategoryDeep,
		ScheduledDate:  monday,
		ScheduledStart: "13:00",
		ScheduledEnd:   "15:30",
		Status:         task.StatusScheduled,
	}

	week := task.NewWeek(monday)
	if err := week.Day(0).AddTask(taskA); err != nil {
		t.Fatalf("add task A: %v", err)
	}

	ww := task.NewWeekWindow(nil, week, nil)
	slotConfig := SlotGridConfigFromWeekWindow(ww, cfg.Schedule.DayStart, cfg.Schedule.DayEnd, time.Now, m.rowHeight)
	m.slotState = NewSlotStateManager(slotConfig)
	m.slotState.SetGrid(WeekWindowToSlotGrid(ww, m.slotState.Config()))

	dayStart := task.TimeToMinutes(cfg.Schedule.DayStart)
	slot := (task.TimeToMinutes("14:00") - dayStart) / m.rowHeight
	refreshCachesForTest(m)
	shadeByDay := m.cachedShadeMap
	dayTasks := m.gridCache[0]
	cursorTask := m.cachedCursorTask()
	cell := m.renderCell(slot-1, 1, dayTasks[slot-1], dayTasks, cursorTask, shadeByDay)
	if !strings.Contains(cell, "13:00-15:30") {
		t.Errorf("expected time range after title, got %q", cell)
	}
}

// Package tui provides the terminal user interface for sancho.
package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/javiermolinar/sancho/internal/config"
	"github.com/javiermolinar/sancho/internal/dwplanner"
	"github.com/javiermolinar/sancho/internal/summary"
	"github.com/javiermolinar/sancho/internal/task"
	"github.com/javiermolinar/sancho/internal/tui/commands"
	"github.com/javiermolinar/sancho/internal/tui/theme"
	"github.com/javiermolinar/sancho/internal/tui/view"
)

// Mode represents the current interaction mode.
type Mode int

const (
	ModeNormal Mode = iota
	ModeEdit        // In edit mode - changes are in-memory until saved
	ModeMove        // Moving a task (can be within edit mode)
	ModePrompt
	ModeModal
)

// ModalType identifies the type of modal.
type ModalType int

const (
	ModalNone       ModalType = iota
	ModalTaskForm             // New task creation
	ModalTaskDetail           // View existing task
	ModalConfirmDelete
	ModalPlanResult // Show LLM planning results
	ModalWeekSummary
	ModalInit
)

type weekSummaryView int

const (
	weekSummaryViewSummary weekSummaryView = iota
	weekSummaryViewTasks
)

// Duration options for task form.
var durationOptions = []int{12, 30, 60}

// Position represents a cursor position in the grid.
type Position struct {
	Day  int // 0=Monday, 6=Sunday
	Slot int // Row index in grid (based on rowHeight)
}

// Model is the main TUI model.
type Model struct {
	// Dependencies
	repo   task.Repository
	config *config.Config

	// Theme and styles
	theme  *theme.Theme
	styles *Styles

	// State manager (slot-based)
	slotState *SlotStateManager

	// State
	weekStart time.Time // Monday of current week
	cursor    Position  // Current cursor position
	mode      Mode
	loading   bool // True when loading week data

	// Move mode (minimal state for UI - SlotStateManager owns the move session)
	moveOriginalDay  int // Original day index of moving task (for view logic)
	moveOriginalSlot int // Original slot of moving task (for view logic)

	// Modal state
	modalType      ModalType       // Current modal type
	modalTask      *task.Task      // Task being viewed/edited (nil for new)
	formDesc       textinput.Model // Description input
	formCategory   int             // 0=deep, 1=shallow
	formDuration   int             // Index into durationOptions
	formFocus      int             // Which field is focused (0=desc, 1=duration)
	confirmMessage string          // Message for confirm modal
	initState      InitState       // Startup initialization state
	initError      string          // Initialization error for modal display

	// Planning state
	planner    *dwplanner.Planner    // LLM planner (created on demand)
	planResult *dwplanner.PlanResult // Current planning result
	planInput  string                // Original plan input (for modify)

	// Overlay state
	overlay OverlayModel

	// Summary state
	weekSummary            *summary.WeekSummary
	weekSummaryView        weekSummaryView
	weekSummarySummaryText []view.WeekSummaryLine
	weekSummaryTasksText   []view.WeekSummaryLine
	weekSummaryCopyText    string

	// Components
	prompt textinput.Model

	// Terminal dimensions and layout
	width        int
	height       int
	rowHeight    int // Minutes per slot (15, 30, or 60)
	rowLines     int // Terminal lines per slot (1, 2, or 3)
	colWidth     int // Dynamic column width based on terminal width
	scrollOffset int // For scrolling the grid

	// Cached render data
	styleCache       StyleCache
	layoutCache      LayoutCache
	renderCache      RenderCache
	gridCache        [7][]*task.Task
	cachedShadeMap   map[int]map[int64]bool
	cachedTaskLines  map[int64][]string
	cacheNeedsUpdate bool

	// Messages
	statusMsg  string    // Temporary status/error message
	statusTime time.Time // When to clear message

	// Error state
	err error
}

// ModelOption configures optional model behavior.
type ModelOption func(*Model)

// WithInitState sets the startup initialization state.
func WithInitState(state InitState) ModelOption {
	return func(m *Model) {
		m.initState = state
		if state.NeedsInit {
			m.mode = ModeModal
			m.modalType = ModalInit
		}
	}
}

// New creates a new TUI model.
func New(repo task.Repository, cfg *config.Config, opts ...ModelOption) *Model {
	ti := textinput.New()
	ti.Placeholder = "/plan ..."

	// Form description input
	formDesc := textinput.New()
	formDesc.Placeholder = "Task name"
	formDesc.CharLimit = 256
	formDesc.Width = 40

	// Load theme from config
	t, err := theme.Load(cfg.UI.Theme)
	if err != nil {
		// Fallback to mocha on error
		t, _ = theme.Load("mocha")
	}

	// Create styles from theme
	styles := NewStyles(t)

	formDesc.PlaceholderStyle = styles.ModalPlaceholderStyle
	formDesc.TextStyle = styles.ModalInputTextStyle
	formDesc.PromptStyle = styles.ModalInputTextStyle
	formDesc.Cursor.Style = styles.ModalInputCursorStyle
	formDesc.Cursor.TextStyle = styles.ModalInputTextStyle

	// Create new slot-based state manager
	// Use current week start as the middle of the 3-week window
	weekStart := startOfWeek(time.Now())
	defaultRowHeight := 60 // Default to 60-min blocks until layout calculated
	slotConfig := SlotGridConfigFromWeekWindow(nil, cfg.Schedule.DayStart, cfg.Schedule.DayEnd, time.Now, defaultRowHeight)
	slotState := NewSlotStateManager(slotConfig)

	m := &Model{
		repo:             repo,
		config:           cfg,
		theme:            t,
		styles:           styles,
		slotState:        slotState,
		weekStart:        weekStart,
		cursor:           Position{Day: weekdayIndex(time.Now()), Slot: 0},
		mode:             ModeNormal,
		prompt:           ti,
		formDesc:         formDesc,
		formCategory:     0, // Default to deep
		formDuration:     1, // Default to 30 min (index 1)
		overlay:          NewOverlayModel(),
		rowHeight:        15, // Default to 15min slots
		rowLines:         1,  // Default to 1 line per slot
		colWidth:         defaultColWidth,
		styleCache:       NewStyleCache(styles, defaultColWidth),
		cacheNeedsUpdate: true,
	}
	m.layoutCache = m.buildLayoutCache(0, 0)

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	if m.initState.NeedsInit {
		return nil
	}
	return commands.LoadInitialWeeks(m.repo, m.weekStart)
}

// Run starts the TUI.
func Run(repo task.Repository, cfg *config.Config) error {
	return RunWithDebug(repo, cfg, false)
}

// RunWithDebug starts the TUI with optional debug logging.
func RunWithDebug(repo task.Repository, cfg *config.Config, debug bool) error {
	if err := InitDebugLogger(debug); err != nil {
		return err
	}
	defer CloseDebugLogger()

	initialRepo := repo
	var initState InitState

	if repo == nil {
		state, err := DetectInitState(cfg)
		if err != nil {
			return err
		}
		initState = state
		if !state.NeedsInit {
			repo, err = openRepo(state.DBPath)
			if err != nil {
				return err
			}
		}
	}

	model := New(repo, cfg, WithInitState(initState))
	model.layoutCache = model.buildLayoutCache(0, 0)
	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if initialRepo == nil {
		if m, ok := finalModel.(Model); ok && m.repo != nil {
			_ = m.repo.Close()
		}
	}
	return err
}

// Package db provides SQLite storage implementation.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite" // SQLite driver

	"github.com/javiermolinar/sancho/internal/task"
)

// SQLite implements task.Repository using SQLite.
type SQLite struct {
	db *sql.DB
}

// New creates a new SQLite repository and runs migrations.
func New(path string) (*SQLite, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	s := &SQLite{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return s, nil
}

// CreateTask adds a new task to the repository.
// Returns ErrTimeBlockOverlap if the task overlaps with an existing scheduled task.
func (s *SQLite) CreateTask(ctx context.Context, t *task.Task) error {
	// Check for overlapping tasks
	if err := s.checkOverlap(ctx, t.ScheduledDate, t.ScheduledStart, t.ScheduledEnd); err != nil {
		return err
	}

	query := `
		INSERT INTO tasks (
			description, category, scheduled_date, scheduled_start, scheduled_end,
			status, outcome, postponed_from, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.ExecContext(ctx, query,
		t.Description,
		t.Category,
		t.ScheduledDate.Format("2006-01-02"),
		t.ScheduledStart,
		t.ScheduledEnd,
		t.Status,
		t.Outcome,
		t.PostponedFrom,
		t.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("inserting task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting last insert id: %w", err)
	}
	t.ID = id

	return nil
}

// GetTask retrieves a task by ID.
func (s *SQLite) GetTask(ctx context.Context, id int64) (*task.Task, error) {
	query := `
		SELECT id, description, category, scheduled_date, scheduled_start, scheduled_end,
		       status, outcome, postponed_from, created_at
		FROM tasks
		WHERE id = ?
	`

	var (
		t             task.Task
		scheduledDate string
		createdAt     string
		outcome       sql.NullString
		postponedFrom sql.NullInt64
	)

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID,
		&t.Description,
		&t.Category,
		&scheduledDate,
		&t.ScheduledStart,
		&t.ScheduledEnd,
		&t.Status,
		&outcome,
		&postponedFrom,
		&createdAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying task: %w", err)
	}

	t.ScheduledDate, err = parseDate(scheduledDate)
	if err != nil {
		return nil, fmt.Errorf("parsing scheduled date: %w", err)
	}

	t.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("parsing created at: %w", err)
	}

	if outcome.Valid {
		o := task.Outcome(outcome.String)
		t.Outcome = &o
	}

	if postponedFrom.Valid {
		t.PostponedFrom = &postponedFrom.Int64
	}

	return &t, nil
}

// CancelTask marks a task as cancelled.
func (s *SQLite) CancelTask(ctx context.Context, id int64) error {
	query := `UPDATE tasks SET status = ? WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, task.StatusCancelled, id)
	if err != nil {
		return fmt.Errorf("cancelling task: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("task %d not found", id)
	}

	return nil
}

// SetTaskOutcome sets the outcome of a task during review.
func (s *SQLite) SetTaskOutcome(ctx context.Context, id int64, outcome task.Outcome) error {
	query := `UPDATE tasks SET outcome = ? WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, outcome, id)
	if err != nil {
		return fmt.Errorf("setting task outcome: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("task %d not found", id)
	}

	return nil
}

// ListTasksByDateRange returns all tasks scheduled within the date range (inclusive).
func (s *SQLite) ListTasksByDateRange(ctx context.Context, start, end time.Time) ([]*task.Task, error) {
	query := `
		SELECT id, description, category, scheduled_date, scheduled_start, scheduled_end,
		       status, outcome, postponed_from, created_at
		FROM tasks
		WHERE scheduled_date >= ? AND scheduled_date <= ?
		ORDER BY scheduled_date, scheduled_start
	`

	rows, err := s.db.QueryContext(ctx, query, start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("querying tasks: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var tasks []*task.Task
	for rows.Next() {
		var (
			t             task.Task
			scheduledDate string
			createdAt     string
			outcome       sql.NullString
			postponedFrom sql.NullInt64
		)

		err := rows.Scan(
			&t.ID,
			&t.Description,
			&t.Category,
			&scheduledDate,
			&t.ScheduledStart,
			&t.ScheduledEnd,
			&t.Status,
			&outcome,
			&postponedFrom,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning task: %w", err)
		}

		t.ScheduledDate, err = parseDate(scheduledDate)
		if err != nil {
			return nil, fmt.Errorf("parsing scheduled date: %w", err)
		}

		t.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, fmt.Errorf("parsing created at: %w", err)
		}

		if outcome.Valid {
			o := task.Outcome(outcome.String)
			t.Outcome = &o
		}

		if postponedFrom.Valid {
			t.PostponedFrom = &postponedFrom.Int64
		}

		tasks = append(tasks, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating tasks: %w", err)
	}

	return tasks, nil
}

// CreateTasks adds multiple tasks in a batch using a transaction.
// Returns ErrTimeBlockOverlap if any task overlaps with existing or other new tasks.
func (s *SQLite) CreateTasks(ctx context.Context, tasks []*task.Task) error {
	if len(tasks) == 0 {
		return nil
	}

	// First, check for overlaps between the new tasks themselves
	if err := checkBatchOverlap(tasks); err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Check for overlaps with existing tasks in the database
	for _, t := range tasks {
		if err := checkOverlapTx(ctx, tx, t.ScheduledDate, t.ScheduledStart, t.ScheduledEnd); err != nil {
			return err
		}
	}

	query := `
		INSERT INTO tasks (
			description, category, scheduled_date, scheduled_start, scheduled_end,
			status, outcome, postponed_from, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, t := range tasks {
		result, err := stmt.ExecContext(ctx,
			t.Description,
			t.Category,
			t.ScheduledDate.Format("2006-01-02"),
			t.ScheduledStart,
			t.ScheduledEnd,
			t.Status,
			t.Outcome,
			t.PostponedFrom,
			t.CreatedAt.Format(time.RFC3339),
		)
		if err != nil {
			return fmt.Errorf("inserting task %q: %w", t.Description, err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("getting last insert id: %w", err)
		}
		t.ID = id
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

// PostponeTask atomically marks the original task as postponed and creates a new task.
// Returns the newly created task with PostponedFrom pointing to the original.
// Returns ErrTimeBlockOverlap if the new time slot overlaps with an existing task.
func (s *SQLite) PostponeTask(ctx context.Context, taskID int64, newDate time.Time, newStart, newEnd string) (*task.Task, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Check for overlapping tasks at the new time slot
	if err := checkOverlapTx(ctx, tx, newDate, newStart, newEnd); err != nil {
		return nil, err
	}

	// Get the original task
	var (
		original      task.Task
		scheduledDate string
		createdAt     string
		outcome       sql.NullString
		postponedFrom sql.NullInt64
	)

	query := `
		SELECT id, description, category, scheduled_date, scheduled_start, scheduled_end,
		       status, outcome, postponed_from, created_at
		FROM tasks
		WHERE id = ?
	`
	err = tx.QueryRowContext(ctx, query, taskID).Scan(
		&original.ID,
		&original.Description,
		&original.Category,
		&scheduledDate,
		&original.ScheduledStart,
		&original.ScheduledEnd,
		&original.Status,
		&outcome,
		&postponedFrom,
		&createdAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task %d not found", taskID)
	}
	if err != nil {
		return nil, fmt.Errorf("querying original task: %w", err)
	}

	// Mark original as postponed
	_, err = tx.ExecContext(ctx, `UPDATE tasks SET status = ? WHERE id = ?`, task.StatusPostponed, taskID)
	if err != nil {
		return nil, fmt.Errorf("marking task as postponed: %w", err)
	}

	// Create new task with reference to original
	insertQuery := `
		INSERT INTO tasks (
			description, category, scheduled_date, scheduled_start, scheduled_end,
			status, outcome, postponed_from, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := tx.ExecContext(ctx, insertQuery,
		original.Description,
		original.Category,
		newDate.Format("2006-01-02"),
		newStart,
		newEnd,
		task.StatusScheduled,
		nil, // new task has no outcome yet
		taskID,
		time.Now().Format(time.RFC3339),
	)
	if err != nil {
		return nil, fmt.Errorf("inserting new task: %w", err)
	}

	newID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("getting new task id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	// Return the new task
	newTask := &task.Task{
		ID:             newID,
		Description:    original.Description,
		Category:       original.Category,
		ScheduledDate:  newDate,
		ScheduledStart: newStart,
		ScheduledEnd:   newEnd,
		Status:         task.StatusScheduled,
		PostponedFrom:  &taskID,
		CreatedAt:      time.Now(),
	}

	return newTask, nil
}

// Close releases database resources.
func (s *SQLite) Close() error {
	return s.db.Close()
}

// UpdateTask updates a task's scheduled times in place.
// Returns ErrTimeBlockOverlap if the new times conflict with another task.
func (s *SQLite) UpdateTask(ctx context.Context, id int64, newStart, newEnd string) error {
	// Get existing task to find date
	t, err := s.GetTask(ctx, id)
	if err != nil {
		return fmt.Errorf("getting task: %w", err)
	}
	if t == nil {
		return fmt.Errorf("task %d not found", id)
	}

	// Check for overlaps (excluding self)
	if err := s.checkOverlapExcluding(ctx, t.ScheduledDate, newStart, newEnd, id); err != nil {
		return err
	}

	query := `UPDATE tasks SET scheduled_start = ?, scheduled_end = ? WHERE id = ?`
	_, err = s.db.ExecContext(ctx, query, newStart, newEnd, id)
	if err != nil {
		return fmt.Errorf("updating task times: %w", err)
	}
	return nil
}

// UpdateTaskDescription updates a task description in place.
func (s *SQLite) UpdateTaskDescription(ctx context.Context, id int64, description string) error {
	description = strings.TrimSpace(description)
	if description == "" {
		return task.ErrEmptyDescription
	}

	t, err := s.GetTask(ctx, id)
	if err != nil {
		return fmt.Errorf("getting task: %w", err)
	}
	if t == nil {
		return fmt.Errorf("task %d not found", id)
	}

	query := `UPDATE tasks SET description = ? WHERE id = ?`
	_, err = s.db.ExecContext(ctx, query, description, id)
	if err != nil {
		return fmt.Errorf("updating task description: %w", err)
	}
	return nil
}

// parseDate parses a date string in various formats SQLite might return.
// Date-only values (midnight) are parsed in local timezone to match time.Now() behavior.
func parseDate(s string) (time.Time, error) {
	// Date-only format: use local timezone (midnight local, not UTC)
	// This ensures dates match when compared with time.Now() based dates
	if t, err := time.ParseInLocation("2006-01-02", s, time.Local); err == nil {
		return t, nil
	}

	// SQLite returns DATE columns as "2006-01-02T00:00:00Z" - extract date and parse as local
	// This is a date-only value stored by SQLite, should be treated as local midnight
	if len(s) == 20 && s[10] == 'T' && s[19] == 'Z' {
		dateOnly := s[:10] // Extract "2006-01-02"
		if t, err := time.ParseInLocation("2006-01-02", dateOnly, time.Local); err == nil {
			return t, nil
		}
	}

	// Formats with actual time components (not midnight placeholders)
	formats := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized date format: %s", s)
}

// checkOverlap checks if a time block overlaps with existing tasks on the same day.
// It uses the database connection (either *sql.DB or *sql.Tx) to query.
// Two time ranges overlap if: start1 < end2 AND start2 < end1
func (s *SQLite) checkOverlap(ctx context.Context, date time.Time, start, end string) error {
	query := `
		SELECT id, scheduled_start, scheduled_end, description
		FROM tasks
		WHERE scheduled_date = ?
		  AND status = ?
		  AND scheduled_start < ?
		  AND scheduled_end > ?
		LIMIT 1
	`

	var (
		id          int64
		existStart  string
		existEnd    string
		description string
	)

	err := s.db.QueryRowContext(ctx, query,
		date.Format("2006-01-02"),
		task.StatusScheduled,
		end,
		start,
	).Scan(&id, &existStart, &existEnd, &description)

	if err == sql.ErrNoRows {
		return nil // No overlap
	}
	if err != nil {
		return fmt.Errorf("checking overlap: %w", err)
	}

	return fmt.Errorf("%w: conflicts with #%d %q (%s-%s)",
		task.ErrTimeBlockOverlap, id, description, existStart, existEnd)
}

// checkOverlapTx is like checkOverlap but uses a transaction.
func checkOverlapTx(ctx context.Context, tx *sql.Tx, date time.Time, start, end string) error {
	query := `
		SELECT id, scheduled_start, scheduled_end, description
		FROM tasks
		WHERE scheduled_date = ?
		  AND status = ?
		  AND scheduled_start < ?
		  AND scheduled_end > ?
		LIMIT 1
	`

	var (
		id          int64
		existStart  string
		existEnd    string
		description string
	)

	err := tx.QueryRowContext(ctx, query,
		date.Format("2006-01-02"),
		task.StatusScheduled,
		end,
		start,
	).Scan(&id, &existStart, &existEnd, &description)

	if err == sql.ErrNoRows {
		return nil // No overlap
	}
	if err != nil {
		return fmt.Errorf("checking overlap: %w", err)
	}

	return fmt.Errorf("%w: conflicts with #%d %q (%s-%s)",
		task.ErrTimeBlockOverlap, id, description, existStart, existEnd)
}

// checkBatchOverlap checks for overlaps between tasks in the same batch.
// Two time ranges overlap if: start1 < end2 AND start2 < end1
func checkBatchOverlap(tasks []*task.Task) error {
	for i := 0; i < len(tasks); i++ {
		for j := i + 1; j < len(tasks); j++ {
			t1, t2 := tasks[i], tasks[j]

			// Only check tasks on the same day
			if !t1.ScheduledDate.Equal(t2.ScheduledDate) {
				continue
			}

			// Check if time ranges overlap
			if task.TimesOverlap(t1.ScheduledStart, t1.ScheduledEnd, t2.ScheduledStart, t2.ScheduledEnd) {
				return fmt.Errorf("%w: %q (%s-%s) conflicts with %q (%s-%s)",
					task.ErrTimeBlockOverlap,
					t1.Description, t1.ScheduledStart, t1.ScheduledEnd,
					t2.Description, t2.ScheduledStart, t2.ScheduledEnd,
				)
			}
		}
	}
	return nil
}

// checkOverlapExcluding checks for overlaps with existing tasks, excluding a specific task ID.
// Used for update operations where the task being updated should not conflict with itself.
func (s *SQLite) checkOverlapExcluding(ctx context.Context, date time.Time, start, end string, excludeID int64) error {
	query := `
		SELECT id, scheduled_start, scheduled_end, description
		FROM tasks
		WHERE scheduled_date = ?
		  AND status = ?
		  AND id != ?
		  AND scheduled_start < ?
		  AND scheduled_end > ?
		LIMIT 1
	`

	var (
		id          int64
		existStart  string
		existEnd    string
		description string
	)

	err := s.db.QueryRowContext(ctx, query,
		date.Format("2006-01-02"),
		task.StatusScheduled,
		excludeID,
		end,
		start,
	).Scan(&id, &existStart, &existEnd, &description)

	if err == sql.ErrNoRows {
		return nil // No overlap
	}
	if err != nil {
		return fmt.Errorf("checking overlap: %w", err)
	}

	return fmt.Errorf("%w: conflicts with #%d %q (%s-%s)",
		task.ErrTimeBlockOverlap, id, description, existStart, existEnd)
}

// BatchUpdateTaskTimes updates multiple tasks' times atomically in a single transaction.
// It validates that the final state has no overlaps before applying changes.
// This is used for move operations where multiple tasks shift positions simultaneously.
func (s *SQLite) BatchUpdateTaskTimes(ctx context.Context, date time.Time, updates []task.TaskTimeUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// 1. Get all scheduled tasks on this date
	query := `
		SELECT id, description, scheduled_start, scheduled_end
		FROM tasks
		WHERE scheduled_date = ?
		  AND status = ?
	`
	rows, err := tx.QueryContext(ctx, query, date.Format("2006-01-02"), task.StatusScheduled)
	if err != nil {
		return fmt.Errorf("querying tasks: %w", err)
	}

	type taskTime struct {
		id          int64
		description string
		start       string
		end         string
	}
	var currentTasks []taskTime
	for rows.Next() {
		var t taskTime
		if err := rows.Scan(&t.id, &t.description, &t.start, &t.end); err != nil {
			_ = rows.Close()
			return fmt.Errorf("scanning task: %w", err)
		}
		currentTasks = append(currentTasks, t)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("closing rows: %w", err)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterating tasks: %w", err)
	}

	// 2. Build final state by applying updates
	updateMap := make(map[int64]task.TaskTimeUpdate)
	for _, u := range updates {
		updateMap[u.ID] = u
	}

	finalState := make([]taskTime, len(currentTasks))
	for i, t := range currentTasks {
		if u, ok := updateMap[t.id]; ok {
			// Apply the update
			finalState[i] = taskTime{
				id:          t.id,
				description: t.description,
				start:       u.NewStart,
				end:         u.NewEnd,
			}
		} else {
			// Keep original times
			finalState[i] = t
		}
	}

	// 3. Check for overlaps in the final state
	for i := 0; i < len(finalState); i++ {
		for j := i + 1; j < len(finalState); j++ {
			t1, t2 := finalState[i], finalState[j]
			if task.TimesOverlap(t1.start, t1.end, t2.start, t2.end) {
				return fmt.Errorf("%w: %q (%s-%s) conflicts with %q (%s-%s)",
					task.ErrTimeBlockOverlap,
					t1.description, t1.start, t1.end,
					t2.description, t2.start, t2.end,
				)
			}
		}
	}

	// 4. Execute all updates
	updateQuery := `UPDATE tasks SET scheduled_start = ?, scheduled_end = ? WHERE id = ?`
	stmt, err := tx.PrepareContext(ctx, updateQuery)
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, u := range updates {
		if _, err := stmt.ExecContext(ctx, u.NewStart, u.NewEnd, u.ID); err != nil {
			return fmt.Errorf("updating task %d: %w", u.ID, err)
		}
	}

	// 5. Commit
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}
